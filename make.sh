#!/usr/bin/env bash

go install
GOOS=linux GOARCH=amd64 go install
GOOS=windows GOARCH=amd64 go install
GOOS=windows GOARCH=386 go install
