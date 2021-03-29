#!/bin/sh

create_doc() {
    GO111MODULE=off go get -v github.com/go-swagger/go-swagger/cmd/swagger
    swagger generate spec -m -o docs/swagger.json --exclude-deps
}

validate_doc() {
    GO111MODULE=off go get -v github.com/go-swagger/go-swagger/cmd/swagger
    swagger validate docs/swagger.json
}

$*