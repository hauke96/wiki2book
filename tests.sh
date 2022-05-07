#!/bin/bash

cd src
go test ./...

cd ../integration-tests
./run.sh
