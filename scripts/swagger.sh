#!/bin/sh

create_doc() {
    GO111MODULE=off go get -v github.com/go-swagger/go-swagger/cmd/swagger
    swagger generate spec -m -o docs/swagger.json --exclude-deps
}

validate_doc() {
    GO111MODULE=off go get -v github.com/go-swagger/go-swagger/cmd/swagger
    swagger validate docs/swagger.json
}

check_for_change() {
    if git diff-index --quiet HEAD; then
        ret=0
    else
        echo "Swagger doc needs to be updated. Try running 'make update-doc' to generate the update-swagger doc"
        exit 1
    fi
}

$*