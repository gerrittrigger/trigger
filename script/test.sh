#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

if [ "$1" = "report" ]; then
  go test -cover -covermode=atomic -coverprofile=coverage.txt -parallel 2 -race -v ./...
else
  list="$(go list ./... | grep -v test)"
  old=$IFS IFS=$'\n'
  for item in $list; do
    go test -cover -covermode=atomic -parallel 2 -race -v "$item"
  done
  IFS=$old
fi
