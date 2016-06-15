# go-docker-machine: Go binding for Docker Machine

[![GoDoc](https://godoc.org/github.com/AkihiroSuda/go-docker-machine?status.svg)](https://godoc.org/github.com/AkihiroSuda/go-docker-machine)
[![Build Status](https://travis-ci.org/AkihiroSuda/go-docker-machine.svg?branch=master)](https://travis-ci.org/AkihiroSuda/go-docker-machine)
[![Go Report Card](https://goreportcard.com/badge/github.com/AkihiroSuda/go-docker-machine)](https://goreportcard.com/report/github.com/AkihiroSuda/go-docker-machine)


go-docker-machine provides a Go binding for Docker Machine.
Inspired by [python-docker-machine](https://github.com/gijzelaerr/python-docker-machine).

Currently, go-docker-machine is tested with Docker Machine 0.6.0.

## Example
[example/main.go](example/main.go):

```go
package main

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	dm "github.com/AkihiroSuda/go-docker-machine"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s MACHINENAME\n", os.Args[0])
		os.Exit(1)
	}
	mname := os.Args[1]

	dockerMachine := dm.NewDockerMachine()

	// client is github.com/docker/engine-api/client.(*Client)
	client, err := dockerMachine.Client(mname)
	if err != nil {
		panic(err)
	}
	info, err := client.Info(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s(%s) has %d containers\n",
		info.Name, info.ServerVersion, info.Containers)
}
```

```bash
$ docker-machine create -d virtualbox dm1
$ go run example/main.go dm1
dm1(1.11.1) has 0 containers
```
