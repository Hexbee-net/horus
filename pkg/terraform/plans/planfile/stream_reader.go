package planfile

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/hexbee-net/horus/pkg/terraform/configs"
	"github.com/hexbee-net/horus/pkg/terraform/configs/configload"
	"github.com/hexbee-net/horus/pkg/terraform/states/statefile"
	"github.com/hexbee-net/horus/pkg/terraform/tfdiags"
	"golang.org/x/xerrors"
	"io"

	"github.com/hexbee-net/horus/pkg/terraform/plans"
)

type StreamReader struct {
	zip *zip.Reader
}

// OpenStream creates a StreamReader for the given file, or returns an error
// if the file doesn't seem to be a planfile.
func OpenStream(reader io.ReaderAt, size int64) (*StreamReader, error) {
	r, err := zip.NewReader(reader, size)
	if err != nil {
		// To give a better error message, we'll sniff to see if this looks like
		// the  old plan format from versions prior to 0.12.
		b := make([]byte, len("tfplan"))
		if _, err := reader.ReadAt(b, 0); err != nil {
			if bytes.HasPrefix(b, []byte("tfplan")) {
				return nil, xerrors.Errorf("the given plan file was created by an earlier version of Terraform; plan files before Terraform 0.12 are not compatible")
			}
		}

		return nil, err
	}

	// Sniff to make sure this looks like a plan file, as opposed to any other
	// random zip file the user might have around.
	var planFile *zip.File
	for _, file := range r.File {
		if file.Name == tfplanFilename {
			planFile = file
			break
		}
	}
	if planFile == nil {
		return nil, xerrors.Errorf("the given file is not a valid plan file")
	}

	// For now, we'll just accept the presence of the tfplan file as enough,
	// and wait to validate the version when the caller requests the plan
	// itself.

	return &StreamReader{
		zip: r,
	}, nil
}

// ReadPlan reads the plan embedded in the plan file.
//
// Errors can be returned for various reasons, including if the plan file
// is not of an appropriate format version, if it was created by a different
// version of Terraform, if it is invalid, etc.
func (r *StreamReader) ReadPlan() (*plans.Plan, error) {
	var planFile *zip.File
	for _, file := range r.zip.File {
		if file.Name == tfplanFilename {
			planFile = file
			break
		}
	}
	if planFile == nil {
		// This should never happen because we checked for this file during
		// Open, but we'll check anyway to be safe.
		return nil, xerrors.Errorf("the plan file is invalid")
	}

	pr, err := planFile.Open()
	if err != nil {
		return nil, xerrors.Errorf("failed to retrieve plan from plan file: %w", err)
	}
	defer func(pr io.ReadCloser) {
		_ = pr.Close()
	}(pr)

	ret, err := readTfplan(pr)
	if err != nil {
		return nil, xerrors.Errorf("failed to read plan from plan file: %w", err)
	}

	prevRunStateFile, err := r.ReadPrevStateFile()
	if err != nil {
		return nil, xerrors.Errorf("failed to read previous run state from plan file: %w", err)
	}
	priorStateFile, err := r.ReadStateFile()
	if err != nil {
		return nil, xerrors.Errorf("failed to read prior state from plan file: %w", err)
	}

	ret.PrevRunState = prevRunStateFile.State
	ret.PriorState = priorStateFile.State

	return ret, nil
}

// ReadStateFile reads the state file embedded in the plan file, which
// represents the "PriorState" as defined in plans.Plan.
//
// If the plan file contains no embedded state file, the returned error is
// statefile.ErrNoState.
func (r *StreamReader) ReadStateFile() (*statefile.File, error) {
	for _, file := range r.zip.File {
		if file.Name == tfstateFilename {
			r, err := file.Open()
			if err != nil {
				return nil, xerrors.Errorf("failed to extract state from plan file: %w", err)
			}
			return statefile.Read(r)
		}
	}
	return nil, statefile.ErrNoState
}

// ReadPrevStateFile reads the previous state file embedded in the plan file, which
// represents the "PrevRunState" as defined in plans.Plan.
//
// If the plan file contains no embedded previous state file, the returned error is
// statefile.ErrNoState.
func (r *StreamReader) ReadPrevStateFile() (*statefile.File, error) {
	for _, file := range r.zip.File {
		if file.Name == tfstatePreviousFilename {
			r, err := file.Open()
			if err != nil {
				return nil, xerrors.Errorf("failed to extract previous state from plan file: %w", err)
			}
			return statefile.Read(r)
		}
	}
	return nil, statefile.ErrNoState
}

// ReadConfigSnapshot reads the configuration snapshot embedded in the plan
// file.
//
// This is a lower-level alternative to ReadConfig that just extracts the
// source files, without attempting to parse them.
func (r *StreamReader) ReadConfigSnapshot() (*configload.Snapshot, error) {
	return readConfigSnapshot(r.zip)
}

// ReadConfig reads the configuration embedded in the plan file.
//
// Internally this function delegates to the configs/configload package to
// parse the embedded configuration and so it returns diagnostics (rather than
// a native Go error as with other methods on Reader).
func (r *StreamReader) ReadConfig() (*configs.Config, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	snap, err := r.ReadConfigSnapshot()
	if err != nil {
		diags = diags.Append(tfdiags.Sourceless(
			tfdiags.Error,
			"Failed to read configuration from plan file",
			fmt.Sprintf("The configuration file snapshot in the plan file could not be read: %s.", err),
		))
		return nil, diags
	}

	loader := configload.NewLoaderFromSnapshot(snap)
	rootDir := snap.Modules[""].Dir // Root module base directory
	config, configDiags := loader.LoadConfig(rootDir)

	diags = diags.Append(configDiags)

	return config, diags
}
