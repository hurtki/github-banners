#!/bin/sh
cd api
go test -count=1 ./... | grep -Ev "no test files|skipped"
cd ../renderer
go test -count=1 ./... | grep -Ev "no test files|skipped"
cd ../storage
go test -count=1 ./... | grep -Ev "no test files|skipped"
