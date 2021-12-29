package apigen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

var (
	config Config

	srcRoot    string
	srcModFile *modfile.File
	dstModFile *modfile.File

	localImports       = NewStringSet()
	copiedLocalImports = NewStringSet()
	externalImports    = NewStringSet()

	srcFilesystem billy.Filesystem
)

func Clone(c Config) error {
	var err error
	config = c

	if config.UseLocalRepo() {
		srcFilesystem = osfs.New("/")
		srcRoot = config.LocalRepo
	} else {
		srcFilesystem = memfs.New()
		srcRoot = "/"
		memStorage := memory.NewStorage()

		cloneOptions := git.CloneOptions{
			URL: fmt.Sprintf("https://github.com/%s/%s", config.Org, config.Repo),
		}
		if config.Tag != "" {
			cloneOptions.ReferenceName = plumbing.NewTagReferenceName(config.Tag)
		}

		_, err := git.Clone(memStorage, srcFilesystem, &cloneOptions)
		if err != nil {
			panic(err)
		}
	}

	err = readSrcModFile(filepath.Join(srcRoot, "go.mod"))
	if err != nil {
		return errors.Wrapf(err, "failed to read src go.mod")
	}

	err = readDstModFile()
	if err != nil {
		return errors.Wrapf(err, "failed to read dst go.mod")
	}

	// tag := "v1.0.2"
	//
	// memStorage := memory.NewStorage()
	// memFs := memfs.New()
	//
	// cloneOptions := git.CloneOptions{
	// 	URL:           fmt.Sprintf("https://github.com/%s/%s", org, repo),
	// 	ReferenceName: plumbing.NewTagReferenceName(tag),
	// }
	// _, err := git.Clone(memStorage, memFs, &cloneOptions)
	// if err != nil {
	// 	panic(err)
	// }

	// List of subdirectories in the repo where to look for api package
	// srcDirs := []string{
	// 	"/",
	// }

	apiDir := "api"

	// Generate API directory in target project where we are calling `go generate`
	dstApiPath := filepath.Join(".", apiDir)

	srcApiPath := fmt.Sprintf("%s/%s", srcRoot, apiDir)
	srcApiDirEntries, err := srcFilesystem.ReadDir(srcApiPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read src api dir %s", srcApiPath)
	}

	for _, srcApiDirEntry := range srcApiDirEntries {
		if srcApiDirEntry.IsDir() && strings.HasPrefix(srcApiDirEntry.Name(), "v") {
			apiVersionName := srcApiDirEntry.Name()
			srcApiVersionDirPath := filepath.Join(srcApiPath, apiVersionName)
			if config.DebugMode {
				log.Printf("Found API version %s", apiVersionName)
			}

			// Copy API version directory
			dstApiVersionDirPath := filepath.Join(dstApiPath, apiVersionName)
			err = copyDirectory(srcApiVersionDirPath, dstApiVersionDirPath)
			if err != nil {
				return errors.Wrapf(err, "failed to copy api version from %s to %s", srcApiVersionDirPath, dstApiVersionDirPath)
			}
		} else {
			if config.DebugMode {
				log.Printf("Skipping entry %s", srcApiDirEntry.Name())
			}
		}
	}

	err = fixDstModFile()
	if err != nil {
		return errors.Wrapf(err, "failed to fix go.mod in generated project")
	}

	return nil
}

func copyDirectory(srcDirPath, dstDirPath string) error {
	err := os.MkdirAll(dstDirPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failed to create directory with path %s", dstDirPath)
	}

	entries, err := srcFilesystem.ReadDir(srcDirPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read src dir %s", srcDirPath)
	}

	for _, entry := range entries {
		srcEntryPath := filepath.Join(srcDirPath, entry.Name())
		dstEntryPath := filepath.Join(dstDirPath, entry.Name())

		if entry.IsDir() {
			err = copyDirectory(srcEntryPath, dstEntryPath)
			if err != nil {
				return errors.Wrapf(err, "failed to recursively copy directory %s to %s", srcEntryPath, dstEntryPath)
			}
		} else {
			err = copyFile(srcEntryPath, dstEntryPath)
			if err != nil {
				return errors.Wrapf(err, "failed to copy file %s to %s", srcEntryPath, dstEntryPath)
			}
		}
	}

	return nil
}

func copyFile(srcPath, dstPath string) (err error) {
	srcFileInfo, err := srcFilesystem.Stat(srcPath)
	if err != nil {
		return errors.Wrapf(err, "failed to get stat for file %s", srcPath)
	}

	// Opening source file for reading. This file can be stored in memory-based
	// file system, or on local disk.
	srcFile, err := srcFilesystem.Open(srcPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", srcPath)
	}
	defer func() {
		closeErr := srcFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// Get src file contents
	srcFileContents := make([]byte, srcFileInfo.Size())
	_, err = srcFile.Read(srcFileContents)
	if err != nil {
		return errors.Wrapf(err, "failed to read src file %s", srcPath)
	}

	// Now let's check import statements in this Go file and fix them if necessary
	// Now let's fix a bit our generated Go file, by correcting import statements.
	srcFileContents, err = fixGoImports(srcPath, srcFileContents)
	if err != nil {
		return errors.Wrapf(err, "failed to adjust go imports in file %s", dstPath)
	}

	err = os.WriteFile(dstPath, srcFileContents, 0666)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file %s to %s", srcPath, dstPath)
	}

	return nil
}
