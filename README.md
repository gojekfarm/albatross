# Albatross
![Build Status](https://github.com/gojekfarm/albatross/actions/workflows/master.yml/badge.svg?branch=master)
[![codecov](https://codecov.io/gh/gojekfarm/albatross/branch/master/graph/badge.svg)](https://codecov.io/gh/gojekfarm/albatross)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg)](https://github.com/goreleaser)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


Albatross wraps the helm package to expose helm operations as HTTP APIs.

## Getting Started

### Prerequisites
* Go >= version 1.12
* Make sure gomodules is enabled(GO111MODULES=on) if the source path is part of GOPATH

### Building from source
Clone the repository and run:
```
make build
```
This places a binary at *bin/albatross*.

### Running
```
make run
```

## Status

Albatross is under development, and there will be breaking changes as part of it's evolution.

## License

```
Copyright 2020 GO-JEK Tech

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```


