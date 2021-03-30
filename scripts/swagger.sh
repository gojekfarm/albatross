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

check_doc_ci() {
    download_url=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/latest | \
    jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url')
    curl -o /usr/local/bin/swagger -L'#' "$download_url"
    chmod +x /usr/local/bin/swagger
    swagger generate spec -m -o docs/swagger.json --exclude-deps
    swagger validate docs/swagger.json
}

$*