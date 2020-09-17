#!/bin/bash
echo "install go-fuzz..."
go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
chmod -R 777 ./fuzzing
cd ./fuzzing/message || exit
echo "go-fuzz-build..."
go-fuzz-build
echo "go-fuzz..."
go-fuzz