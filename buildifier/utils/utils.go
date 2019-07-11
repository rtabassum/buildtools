// Package utils contains shared methods that can be used by different implementations of
// buildifier binary
package utils

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/warn"
)

func isStarlarkFile(name string) bool {
	ext := filepath.Ext(name)
	switch ext {
	case ".bzl", ".sky":
		return true
	}
	base := name[:len(name)-len(ext)]

	switch ext {
	case ".bazel", ".oss":
		// The extension can be ignored
		name = base
	}

	return name == "BUILD" || name == "WORKSPACE"
}

func skip(info os.FileInfo) bool {
	return info.IsDir() && info.Name() == ".git"
}

// ExpandDirectories takes a list of file/directory names and returns a list with file names
// by traversing each directory recursively and searching for relevant Starlark files.
func ExpandDirectories(args *[]string) ([]string, error) {
	files := []string{}
	for _, arg := range *args {
		info, err := os.Stat(arg)
		if err != nil {
			return []string{}, err
		}
		if !info.IsDir() {
			files = append(files, arg)
			continue
		}
		err = filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if skip(info) {
				return filepath.SkipDir
			}
			if !info.IsDir() && isStarlarkFile(info.Name()) {
				files = append(files, path)
			}
			return err
		})
		if err != nil {
			return []string{}, err
		}
	}
	return files, nil
}

// GetParser returns a parser for a given file type
func GetParser(inputType string) func(filename string, data []byte) (*build.File, error) {
	switch inputType {
	case "build":
		return build.ParseBuild
	case "bzl":
		return build.ParseBzl
	case "auto":
		return build.Parse
	case "workspace":
		return build.ParseWorkspace
	default:
		return build.ParseDefault
	}
}

// GetPackageName returns the package name of a file by searching for a WORKSPACE file
func GetPackageName(filename string) string {
	dirs := filepath.SplitList(path.Dir(filename))
	parent := ""
	index := len(dirs) - 1
	for i, chunk := range dirs {
		parent = path.Join(parent, chunk)
		metadata := path.Join(parent, "METADATA")
		if _, err := os.Stat(metadata); !os.IsNotExist(err) {
			index = i
		}
	}
	return strings.Join(dirs[index+1:], "/")
}

// Lint calls the linter and returns a list of unresolved findings
func Lint(f *build.File, pkg, lint string, warningsList *[]string, verbose bool) []*warn.Finding {
	switch lint {
	case "warn":
		return warn.FileWarnings(f, pkg, *warningsList, nil, warn.ModeWarn)
	case "fix":
		warn.FixWarnings(f, pkg, *warningsList, verbose)
	}
	return nil
}
