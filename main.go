package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"

	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var cniPath = flag.String("cni.path", "/var/lib/cni/networks/podnet", "Path to lookup CNI IP Assigments")

func main() {
	flag.Parse()

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal("Unable to connect to docker agent", err.Error())
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Quiet: true})
	if err != nil {
		log.Fatal("Unable to list Containers", err.Error())
	}

	assignmentFiles, err := ioutil.ReadDir(*cniPath)
	if err != nil {
		log.Fatal("Unable to list assigned IPs", err.Error())
	}

	for _, assignment := range assignmentFiles {

		assignmentFile := assignment.Name()

		assignmentFileContents, err := ioutil.ReadFile(path.Join(*cniPath, assignmentFile))
		if err != nil {
			log.Fatal("Unable to open assignment file", err.Error())
		}

		isIPAddress, err := regexp.MatchString(`\d+[.]\d+[.]\d+[.]\d+`, assignmentFile)

		if !isIPAddress {
			continue
		}

		containerID := string(assignmentFileContents)

		if isRunningContainer(containerID, containers) {
			fmt.Println("Container found. Assignment active")

		} else {
			fmt.Println("Container " + containerID + " not found. Removing assigned IP")
			os.Remove(path.Join(*cniPath, assignmentFile))
		}
	}
}

func isRunningContainer(containerID string, containers []types.Container) bool {

	for _, container := range containers {

		if container.ID == containerID {
			return true
		}
	}

	return false
}
