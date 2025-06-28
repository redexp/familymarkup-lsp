#!/bin/sh

case "$1" in
    build)
        echo "build linux"
        go build -o build/linux-x64 main.go
        ;;
    win)
        echo "build windows"
        GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o build/main.exe main.go
        ;;
    gowasm)
        echo "go build wasm"
        GOOS=js GOARCH=wasm go build -o build/main.wasm main.go
        ;;
    wasm)
        echo "tinygo build wasm"
        GOOS=wasip1 GOARCH=wasm tinygo build -o build/main.wasm main.go
        ls -la build/main.wasm
        ;;
    *)
        echo "go run"
        go run main.go
        ;;
esac
