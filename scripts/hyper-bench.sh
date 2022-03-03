#!/usr/bin/env bash
set -e
shopt -s expand_aliases
alias time='date; time'

scriptdir=$(cd $(dirname $0); pwd -P)
sourcedir=$(cd $scriptdir/..; pwd -P)

# Micro benchmarks without OpenAll and Commit
time go test -v ./vcs -bench=BenchmarkPrunedVCS -run=BenchmarkPrunedVCS -benchtime 4x -benchmem -timeout 10800m -json

# Benchmarks of Hyperproofs aggregation
time go test -v ./vcs -bench=BenchmarkVCSAgg -run=BenchmarkVCSAgg -benchtime 2x -benchmem -timeout 360m -json

# This computes the estimate verification time of SNARK based Merkle aggregation.
go build && time ./hyperproofs-go 1

# # WARNING: Benchmarking OpenAll and Commit takes around 6.5 hours.
# go build && time ./hyperproofs-go 2 # This benchmarks OpenAll and Commit
