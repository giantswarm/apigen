package apigen

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Lockfile struct {
	Generated []string `json:"generated"`
}

var lockfile Lockfile = Lockfile{}

func addGeneratedFileToLockfile(path string) error {
	lockfile.Generated = append(lockfile.Generated, path)
	return nil
}

func writeLockfile() error {
	lockfilePath := filepath.Join(config.TargetDir, "apigen.lock")
	lockfileContents, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal lockfile contents to JSON:\n%s", lockfile)
	}

	err = os.WriteFile(lockfilePath, lockfileContents, 0664)
	if err != nil {
		return errors.Wrapf(err, "failed to write lockfile with contents:\n%s", lockfileContents)
	}

	return nil
}
