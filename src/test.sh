#!/bin/bash

go test -coverprofile test.out ./...
go tool cover -html=test.out
