#!/bin/bash
GOOS=js GOARCH=wasm go build -o aqlparse.wasm && go run server/main.go `pwd`
