#!/usr/bin/env bash

mkdir -p coverage
go test ./... -coverprofile=coverage/coverage.out
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

python3 -m http.server -d coverage/
