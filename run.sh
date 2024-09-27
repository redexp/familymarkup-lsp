#!/bin/sh

# Проверка первого аргумента с использованием case
case "$1" in
    build)
        echo "go build"
        go build -o build/main main.go
        ;;
    *)
        echo "go run"
        go run .
        ;;
esac
