package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"

	"github.com/giantswarm/apigen"
)

var config = apigen.Config{}

func main() {
	localRepoPtr := flag.String("local-repo", "", "local repository")
	orgPtr := flag.String("org", "kubernetes-sigs", "GitHub organization name")
	repoPtr := flag.String("repo", "cluster-api", "GitHub repo name")
	tagPtr := flag.String("tag", "", "Project version (GitHub release/tag name)")
	debugPtr := flag.Bool("debug", false, "Run in debug mode")
	targetDirPtr := flag.String("target-dir", "", "Where to generate code")
	flag.Parse()

	if debugPtr != nil && *debugPtr {
		config.DebugMode = true
	} else {
		config.DebugMode = false
	}

	if localRepoPtr != nil && *localRepoPtr != "" {
		config.LocalRepo = *localRepoPtr
	} else if orgPtr != nil && repoPtr != nil {
		if *orgPtr == "" {
			printError(errors.New("Flag 'org' cannot must be set"))
			return
		}
		config.Org = *orgPtr

		if *repoPtr == "" {
			printError(errors.New("Flag 'repo' cannot must be set"))
			return
		}
		config.Repo = *repoPtr

		if tagPtr != nil && *tagPtr != "" {
			config.Tag = *tagPtr
		}
	}

	if targetDirPtr != nil && *targetDirPtr != "" {
		if _, err := os.Stat(*targetDirPtr); os.IsNotExist(err) {
			printError(errors.Errorf("Target directory %s not found", *targetDirPtr))
			return
		}

		config.TargetDir = *targetDirPtr
	}

	err := apigen.Clone(config)
	if err != nil {
		printError(err)
	}
}

func printError(err error) {
	if config.DebugMode {
		log.Fatal(err)
	} else {
		fmt.Println(err.Error())
	}
}
