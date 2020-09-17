#!/bin/bash
 docker run --rm -v "$PWD":/usr/src/iota.go -w /usr/src/iota.go golang:1.14-buster \
 go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build \
 cd fuzzing \
go-fuzz-build \
go-fuzz