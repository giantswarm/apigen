package apigen

import (
	"strings"
)

type Config struct {
	LocalRepo      string
	Org            string
	Repo           string
	Tag            string
	TargetDir      string
	APIVersion     string
	AdditionalDirs []string
	ExcludeGlobs   []string

	DebugMode bool
}

func (c *Config) UseLocalRepo() bool {
	return c.LocalRepo != ""
}

func (c *Config) ShouldCopyAPIVersion(apiVersion string) bool {
	return c.APIVersion == "" || strings.ToLower(c.APIVersion) == strings.ToLower(apiVersion)
}
