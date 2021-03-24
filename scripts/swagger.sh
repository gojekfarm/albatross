#!/bin/sh

create_doc() {
    GO111MODULE=off go get -v github.com/go-swagger/go-swagger/cmd/swagger
    swagger generate spec -m -o swagger.yaml --exclude-deps
}

validate_doc() {
    GO111MODULE=off go get -v github.com/go-swagger/go-swagger/cmd/swagger
    swagger validate swagger.yaml
}

$*