package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/giantswarm/apigen"
)

var config = apigen.Config{}

var rootCmd = &cobra.Command{
	Use: "goclone",
	Example: `
  # Copy from a remote GitHub repo
  goclone --org kubernetes-sigs --repo cluster-api --tag v1.0.2 --target-dir ./out/

  # Copy from a local repo
  goclone --local-repo ../cluster-api --target-dir ./out/

	# Ignore files matching pattern
  goclone --local-repo ../cluster-api --target-dir ./out/ --exclude "*_test.go" --exclude "doc.go"

  # Copy additional directories
  goclone --org kubernetes-sigs --repo cluster-api-provider-aws --tag v1.0.0 --target-dir ./out --additional-dir exp/api`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if config.LocalRepo == "" {
			cmd.MarkFlagRequired("org")
			cmd.MarkFlagRequired("repo")
		}

		if config.TargetDir != "" {
			if _, err := os.Stat(config.TargetDir); os.IsNotExist(err) {
				return errors.Errorf("Target directory %s not found", config.TargetDir)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return apigen.Clone(config)
	},
}

func init() {
	rootCmd.Flags().StringVar(&config.LocalRepo, "local-repo", "", "the local repository")
	rootCmd.Flags().StringVar(&config.Org, "org", "", "the GitHub organization name")
	rootCmd.Flags().StringVar(&config.Repo, "repo", "", "the GitHub repo name")
	rootCmd.Flags().StringVar(&config.Tag, "tag", "", "Project version (GitHub release/tag name)")
	rootCmd.Flags().StringVar(&config.TargetDir, "target-dir", "", "Where to generate code")
	rootCmd.Flags().StringArrayVar(&config.AdditionalDirs, "additional-dir", []string{}, "additional directories to copy from source repo")
	rootCmd.Flags().StringArrayVar(&config.ExcludeGlobs, "exclude", []string{}, "glob patterns to exclude")
	rootCmd.Flags().BoolVar(&config.DebugMode, "debug", false, "Run in debug mode")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if config.DebugMode {
			log.Fatal(err)
		} else {
			fmt.Println(err.Error())
		}
	}
}
