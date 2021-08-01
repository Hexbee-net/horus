terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 3.27"
    }
  }

  required_version = ">= 0.14.9"
}

provider "aws" {
  profile = "default"
  region = "eu-west-3"
}


resource "aws_instance" "simple_resource" {
  ami = "ami-062fdd189639d3e93"
  instance_type = "t2.micro"

  tags = {
    Name = "ExampleAppServerInstance 1"
  }
}

resource "aws_instance" "multiple_resource" {
  count = 3
  ami = "ami-062fdd189639d3e93"
  instance_type = "t2.micro"

  tags = {
    Name = "ExampleAppServerInstance 2"
  }
}

resource "null_resource" "foo" {
  triggers = {
    foo = "bar"
  }
}
