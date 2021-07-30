package warden

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		args    []*Options
		wantErr bool
	}{
		{
			name:    "no options",
			args:    nil,
			wantErr: false,
		}, {
			name:    "empty options",
			args:    []*Options{{}},
			wantErr: false,
		}, {
			name: "user module with no content",
			args: []*Options{{
				UserModules: &[]UserModule{{
					Name:   "no_content",
					Script: "",
				}},
			}},
			wantErr: false,
		}, {
			name: "invalid syntax in user module",
			args: []*Options{{
				UserModules: &[]UserModule{{
					Name:   "no_content",
					Script: `definitely not lua code`,
				}},
			}},
			wantErr: true,
		}, {
			name: "valid user module",
			args: []*Options{{
				UserModules: &[]UserModule{{
					Name:   "user_module",
					Script: `print("this looks ok")`,
				}},
			}},
			wantErr: false,
		}, {
			name: "invalid syntax in main script",
			args: []*Options{{
				Script: `definitely not lua code`,
			}},
			wantErr: true,
		}, {
			name: "valid script",
			args: []*Options{{
				Script: `print("this looks ok")`,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestWarden_ValidatePlan_NilPlanFile(t *testing.T) {
	w, err := New()
	require.NoError(t, err)

	assert.PanicsWithError(t, "runtime error: invalid memory address or nil pointer dereference", func() {
		_, _ = w.ValidatePlan(nil)
	})
}

func TestWarden_ValidatePlan_EmptyFile(t *testing.T) {
	w, err := New()
	require.NoError(t, err)

	fs := afero.NewMemMapFs()
	file, err := fs.Create("ts-planfile")
	require.NoError(t, err)

	_, err = w.ValidatePlan(file)
	assert.Error(t, err)
}

func TestWarden_ValidatePlan_ValidFile(t *testing.T) {
	w, err := New()
	require.NoError(t, err)

	fs := afero.NewReadOnlyFs(afero.NewOsFs())
	planFile, err := fs.Open("../../testData/tf-planfile")
	require.NoError(t, err)

	_, err = w.ValidatePlan(planFile)
	assert.NoError(t, err)
}

func TestWarden_ValidatePlan_ReturnValues(t *testing.T) {
	testFs := afero.NewReadOnlyFs(afero.NewOsFs())

	tests := []struct {
		name    string
		options Options
		file    string
		issues  []string
		wantErr bool
	}{
		{
			name: "no return",
			options: Options{
				Script: "",
			},
			file:    "../../testData/tf-planfile",
			issues:  nil,
			wantErr: false,
		}, {
			name: "boolean - true",
			options: Options{
				Script: `return true`,
			},
			file:    "../../testData/tf-planfile",
			issues:  nil,
			wantErr: false,
		}, {
			name: "boolean - false",
			options: Options{
				Script: `return false`,
			},
			file:    "../../testData/tf-planfile",
			issues:  []string{"validation failed"},
			wantErr: true,
		}, {
			name: "string - one",
			options: Options{
				Script: `return "test - found one issue"`,
			},
			file:    "../../testData/tf-planfile",
			issues:  []string{"test - found one issue"},
			wantErr: true,
		}, {
			name: "string - several",
			options: Options{
				Script: `return { "test - issue 1", "test - issue 2", "test - issue 3", "test - issue 4" }`,
			},
			file: "../../testData/tf-planfile",
			issues: []string{
				"test - issue 1",
				"test - issue 2",
				"test - issue 3",
				"test - issue 4",
			},
			wantErr: true,
		}, {
			name: "other",
			options: Options{
				Script: `return 123`,
			},
			file:    "../../testData/tf-planfile",
			issues:  []string{"validation failed (123)"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := New(&tt.options)
			require.NoError(t, err)

			planFile, err := testFs.Open(tt.file)
			require.NoError(t, err)

			issues, err := w.ValidatePlan(planFile)

			if tt.wantErr {
				assert.ErrorIs(t, err, ErrValidationFailed)
			} else {
				assert.NoError(t, err)
			}

			assert.ElementsMatch(t, tt.issues, issues)
		})
	}
}

func TestWarden_ValidatePlan_UserModules(t *testing.T) {
	testFs := afero.NewReadOnlyFs(afero.NewOsFs())

	tests := []struct {
		name    string
		options Options
		file    string
		issues  []string
		wantErr bool
	}{
		{
			name: "missing module",
			options: Options{
				Script: `
local UserModule = require 'missing-module'
return true
`,
			},
			file:    "../../testData/tf-planfile",
			issues:  nil,
			wantErr: true,
		}, {
			name: "module not included",
			options: Options{
				UserModules: &[]UserModule{
					{
						Name: "mod1",
						Script: `
local function ok(v)
    return true
end
`,
					},
				},
				Script: `return mod1.ok()`,
			},
			file:    "../../testData/tf-planfile",
			issues:  nil,
			wantErr: true,
		}, {
			name: "module ok",
			options: Options{
				UserModules: &[]UserModule{
					{
						Name: "mod1",
						Script: `
local function ok(v)
    return true
end
`,
					},
				},
				Script: `
local mod1 = require 'mod1'
return mod1.ok()
`,
			},
			file:    "../../testData/tf-planfile",
			issues:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := New(&tt.options)
			require.NoError(t, err)

			planFile, err := testFs.Open(tt.file)
			require.NoError(t, err)

			issues, err := w.ValidatePlan(planFile)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.ElementsMatch(t, tt.issues, issues)
		})
	}
}
