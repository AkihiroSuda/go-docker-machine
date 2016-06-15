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
