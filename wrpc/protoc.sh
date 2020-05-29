#!/bin/bash

# The shell shall write to standard error a trace for each command after it expands the command and
# before it executes it.
set -x

protoc --go_out=plugins=grpc:./ tasks.proto