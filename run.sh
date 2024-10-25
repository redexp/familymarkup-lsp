#!/bin/sh

# Проверка первого аргумента с использованием case
case "$1" in
    build)
        echo "go build"
        go build -o build/main main.go
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
