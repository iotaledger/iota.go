#!/bin/bash
 docker run --rm -v "$PWD":/usr/src/iota.go -w /usr/src/iota.go golang:1.14 ./fuzzing/run.sh