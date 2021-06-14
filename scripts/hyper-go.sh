#!/usr/bin/env bash
set -e
shopt -s expand_aliases
alias time='date; time'

scriptdir=$(cd $(dirname $0); pwd -P)
sourcedir=$(cd $scriptdir/..; pwd -P)

go build && time ./hyperproofs-go
time go test -v ./vcs -bench=BenchmarkPrunedVCS -run=BenchmarkPrunedVCS -benchtime 4x -timeout 360m
# go build && time ./go-mcl-benchmarks -test.benchtime 1x
