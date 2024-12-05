#!/bin/sh

case "$1" in
    build)
        echo "go build"
        go build -o build/main main.go
        ;;
    win)
        echo "go build windows"
        GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o build/main.exe main.go
        ;;
    wasm)
        echo "go build wasm"
        GOOS=js GOARCH=wasm go build -o build/main.wasm main.go
        ;;
    *)
        echo "go run"
        go run main.go
        ;;
esac
