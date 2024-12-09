package main

import (
	"fmt"

	"github.com/yakuter/go-mage/pkg/buildvars"
)

const projectName = "go-mage"

func main() {
	fmt.Println(helloMessage())
}

func helloMessage() string {
	message := ""
	message = fmt.Sprintf("%s\nVersion: %s\nCommit ID: %s\nBuild Time: %s\nBuild Mode: %s",
		projectName, buildvars.GetVersion(), buildvars.GetCommitID(), buildvars.GetBuildTime(), buildvars.GetBuildMode())
	return message
}
