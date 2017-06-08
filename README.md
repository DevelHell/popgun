# POPgun

[![Build Status](https://circleci.com/gh/DevelHell/popgun.svg?style=shield&circle-token=:circle-token)](https://circleci.com/gh/DevelHell/popgun) [![Coverage Status](https://coveralls.io/repos/github/DevelHell/popgun/badge.svg?branch=master)](https://coveralls.io/github/DevelHell/popgun?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/DevelHell/popgun)](https://goreportcard.com/report/github.com/DevelHell/popgun)

POPgun is a lightweight POP3 server implementation in Go. POPgun meets [RFC1939](https://www.ietf.org/rfc/rfc1939.txt)
and was mainly created for [develmail.com](https://develmail.com).

## Getting Started

POPgun is meant to be used as a package and you need to create your own implementations
of `Authorizator` and `Backend` interfaces.

#### 1. Import the POPgun package
```go
import (
    "github.com/DevelHell/popgun"
)
```

#### 2. Implement Authorizator and Backend interfaces

`Authorizator` is used for user authorization and there's only one function `Authorize(user, pass string)`. Be aware that single instance is shared
across all client connections.

`Backend` is used for mail storage access, e.g. database storage. Single `Backend` instance is shared across all client connections connections as well. 

Example dummy implementations can be found in `backend` package, see comments in these files for more information. When your're done, create an instance of both of them:
```go
backend := backends.DummyBackend{}
authorizator := backends.DummyAuthorizator{}
```

#### 3. Configure and run the server
There is only one configuration field for now - `ListenInterface`, which defines interface (ip address) and port to listen on.
Server is started in separate go routine, so be sure to keep the server busy, e.g. using wait groups:

```go
cfg := popgun.Config{
    ListenInterface: "localhost:1100",
}

var wg sync.WaitGroup
wg.Add(1)

server := NewServer(cfg, authorizator, backend)
err := server.Start()
if err != nil {
    log.Fatal(err)
    return
}
wg.Wait()
```
Server is logging to `stderr` using `log` package.

## License and Contribution

POPgun is released under MIT license. Feel free to fork, redistribute or contribute!