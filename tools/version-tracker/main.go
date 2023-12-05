package main

import (
	"os"

	"github.com/aws/eks-anywhere-build-tooling/tools/version-tracker/cmd"
)

func main() {
	if cmd.Execute() == nil {
		os.Exit(0)
	}
	os.Exit(-1)
}
