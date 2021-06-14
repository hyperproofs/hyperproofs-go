#!/usr/bin/env bash
set -e
shopt -s expand_aliases
alias time='date; time'

scriptdir=$(cd $(dirname $0); pwd -P)
sourcedir=$(cd $scriptdir/..; pwd -P)

# time go test -v ./vcs -bench=. -run=Bench -benchtime 2x -timeout 360m
# time go test -v ./vcs -bench=BenchmarkVCSAgg -run=BenchmarkVCSAgg -benchtime 2x -timeout 360m
time go test -v ./vcs -bench=BenchmarkPrunedVCS -run=BenchmarkPrunedVCS -benchtime 4x -benchmem -timeout 10800m -json
#time go test -v -bench=BenchmarkHashing -run=BenchmarkHashing -benchtime=8000x -benchmem -timeout 10800m -json
# time go test -v ./gipa -bench=. -run=Bench -benchtime 4x
# time go test -v ./agggipa -bench=. -run=Bench -benchtime 1x
# time go test -v ./... -bench=. -run=Bench -benchtime 4x
#time go test -v ./... -bench=. -benchtime 4x
