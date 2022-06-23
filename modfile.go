package apigen

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
)

func readSrcModFile(modFilePath string) error {
	srcMod, err := srcFilesystem.Open(modFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open go.mod file %s", modFilePath)
	}
	defer srcMod.Close()

	srcModStat, err := srcFilesystem.Stat(modFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read go.mod file %s stat", modFilePath)
	}
	size := int(srcModStat.Size()) // assuming go.mod is not larger than max int (~2Gb)

	modFileContents := make([]byte, size)
	_, err = srcMod.Read(modFileContents)
	if err != nil {
		return errors.Wrapf(err, "failed to read source go.mod file %s", modFilePath)
	}

	srcModFile, err = modfile.Parse("go.mod", modFileContents, nil)
	if err != nil {
		return errors.Wrap(err, "failed to parse go.mod file")
	}

	if config.DebugMode {
		if len(srcModFile.Require) > 0 {
			log.Printf("Found imports in go.mod file %s (see next lines)", modFilePath)
			for _, required := range srcModFile.Require {
				log.Println(required.Mod)
			}
		} else {
			log.Printf("Imports not found in go.mod file %s", modFilePath)
		}
	}

	return nil
}

func readDstModFile() error {
	modFilePath := filepath.Join(config.TargetDir, "go.mod")
	modFileContents, err := os.ReadFile(modFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to read go.mod file")
	}

	// fixer := func(path string, version string) (string, error) {
	// 	return module.CanonicalVersion(version), nil
	// }

	dstModFile, err = modfile.Parse(modFilePath, modFileContents, nil)
	if err != nil {
		return errors.Wrap(err, "failed to parse go.mod file")
	}

	if config.DebugMode {
		if len(dstModFile.Require) > 0 {
			log.Printf("Found imports in go.mod file %s (see next lines)", modFilePath)
			for _, required := range dstModFile.Require {
				log.Println(required.Mod)
			}
		} else {
			log.Printf("Imports not found in go.mod file %s", modFilePath)
		}
	}

	return nil
}

func fixDstModFile() error {
	if externalImports.Size() == 0 {
		if config.DebugMode {
			log.Println("External imports not found")
		}
		return nil
	}

	if config.DebugMode {
		log.Println("Found external imports (see next lines)")
	}

	for dep, ok := externalImports.TakeOne(); ok; dep, ok = externalImports.TakeOne() {
		srcRequire, ok := modFileFindRequiresForPackage(srcModFile, dep)
		if !ok {
			return errors.Errorf("Cannot determine which module to import for package %s", dep)
		}

		err := addRequireToModFile(dstModFile, srcRequire)
		if err != nil {
			return errors.Wrap(err, "failed to add new require to dst go.mod")
		}
	}

	dstModFile.SortBlocks()
	dstGoModContents, err := dstModFile.Format()
	if err != nil {
		return errors.Wrap(err, "failed to format dst go.mod")
	}

	modFilePath := filepath.Join(config.TargetDir, "go.mod")
	err = os.WriteFile(modFilePath, dstGoModContents, 0664)
	if err != nil {
		return errors.Wrap(err, "failed to update dst go.mod")
	}

	return nil
}

func modFileFindRequiresForPackage(modFile *modfile.File, pkg string) (modfile.Require, bool) {
	for _, require := range modFile.Require {
		if require == nil {
			continue
		}

		if strings.HasPrefix(pkg, require.Mod.Path) {
			return *require, true
		}
	}

	return modfile.Require{}, false
}

func addRequireToModFile(modFile *modfile.File, newRequire modfile.Require) error {
	for _, require := range modFile.Require {
		if require == nil {
			continue
		}

		if require.Mod.Path == newRequire.Mod.Path && require.Mod.Version == require.Mod.Version {
			// dst go.mod is already requesting this module
			if config.DebugMode {
				log.Printf("go.mod already has require for %s", newRequire.Mod.String())
			}
			return nil
		}
	}

	err := modFile.AddRequire(newRequire.Mod.Path, newRequire.Mod.Version)
	if err != nil {
		return errors.Wrapf(err, "failed to add new require line for module %s", newRequire.Mod.String())
	}
	if config.DebugMode {
		log.Printf("Added go.mod require for %s", newRequire.Mod.String())
	}

	return nil
}
