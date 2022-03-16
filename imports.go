package apigen

import (
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func fixGoImports(goFilePath string, srcFileContents []byte) ([]byte, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "", srcFileContents, parser.ImportsOnly)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse imports from go file %s", goFilePath)
	}

	fileContents := srcFileContents

	if len(astFile.Imports) > 0 {
		if config.DebugMode {
			log.Printf("Fixing imports in file %s", goFilePath)
		}
	} else {
		if config.DebugMode {
			log.Printf("Imports not found in file %s", goFilePath)
		}
		return fileContents, nil
	}

	var fixedFileContents []byte
	currentPackage := srcModFile.Module.Mod.Path
	newPackage := dstModFile.Module.Mod.Path
	var lastPosition int

	for _, importSpec := range astFile.Imports {
		requiredPackage := strings.Trim(importSpec.Path.Value, "\"")

		if isSystemPackage(requiredPackage) {
			// noop
		} else if isLocalPackage(requiredPackage) {
			// rewrite import in Go file
			currentPackageStart := int(importSpec.Path.Pos())
			currentPackageEnd := currentPackageStart + len(currentPackage)
			importEnd := int(importSpec.Path.End())

			// Append Go code from last saved position until this import
			// copy fixedFileContents[lastPosition:currentPackageStart] to new array
			fixedFileContents = append(fixedFileContents, srcFileContents[lastPosition:currentPackageStart]...)

			// Now then append the new package name.
			fixedFileContents = append(fixedFileContents, []byte(newPackage)...)

			// Now append the remaining part of the original src Go file.
			fixedFileContents = append(fixedFileContents, srcFileContents[currentPackageEnd:importEnd]...)

			// remember where we stopped
			lastPosition = importEnd

			if !isLocalApiPackage(requiredPackage) {
				// localImports.Add(requiredPackage)
				err = copyLocalPackage(requiredPackage)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to copy required local package %s", requiredPackage)
				}
			}
			// fileContents = rewritePackageImport(fileContents, importSpec, srcModFile.Module.Mod.Path, dstModFile.Module.Mod.Path)
		} else {
			externalImports.Add(requiredPackage)
		}
	}

	// append rest of the file
	// append Go code from last position until this import
	// copy fixedFileContents[lastPosition:importStart] to new array
	fixedFileContents = append(fixedFileContents, srcFileContents[lastPosition:]...)

	return fixedFileContents, nil
}

func isSystemPackage(p string) bool {
	return !strings.Contains(p, ".")
}

func isLocalPackage(p string) bool {
	packageName := strings.Trim(p, "\"")
	return strings.HasPrefix(packageName, srcModFile.Module.Mod.Path)
}

func isLocalApiPackage(p string) bool {
	packageName := strings.Trim(p, "\"")
	apiPackageName := filepath.Join(srcModFile.Module.Mod.Path, "api")
	return strings.HasPrefix(packageName, apiPackageName)
}

func copyLocalPackage(localPackage string) error {
	if copiedLocalImports.Contains(localPackage) {
		return nil
	}
	copiedLocalImports.Add(localPackage)
	if config.DebugMode {
		log.Printf("copying local package %s", localPackage)
	}

	srcModule := srcModFile.Module.Mod.Path
	relativePackagePath := strings.TrimPrefix(localPackage, srcModule)
	relativePackagePath = strings.Trim(relativePackagePath, "/")
	srcPackagePath := filepath.Join(srcRoot, relativePackagePath)
	dstPackagePath := filepath.Join(config.TargetDir, relativePackagePath)

	err := copyDirectory(srcPackagePath, dstPackagePath, false)
	if err != nil {
		return errors.Wrapf(err, "failed to copy package from %s to %s", srcPackagePath, dstPackagePath)
	}

	return nil
}
