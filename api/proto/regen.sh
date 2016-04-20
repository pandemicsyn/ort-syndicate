#!/bin/bash
set -x
protoc --go_out=plugins=grc:. *.proto
go install .
