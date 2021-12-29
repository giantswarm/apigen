package apigen

type Config struct {
	LocalRepo string
	Org       string
	Repo      string
	Tag       string
	TargetDir string

	DebugMode bool
}

func (c *Config) UseLocalRepo() bool {
	return c.LocalRepo != ""
}
