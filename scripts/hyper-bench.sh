#!/usr/bin/env bash
set -e
shopt -s expand_aliases
alias time='date; time'

scriptdir=$(cd $(dirname $0); pwd -P)
sourcedir=$(cd $scriptdir/..; pwd -P)

time go test -v ./vcs -bench=BenchmarkPrunedVCS -run=BenchmarkPrunedVCS -benchtime 4x -benchmem -timeout 10800m -json
# Note: Benchtime has to be specified (Ex: 1x, 2x, 4x, 8x)
time go test -v ./vcs -bench=BenchmarkVCSAgg -run=BenchmarkVCSAgg -benchtime 2x -benchmem -timeout 360m -json
