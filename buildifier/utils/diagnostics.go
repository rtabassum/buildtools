package utils

import (
	"encoding/json"
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/warn"
	"sort"
	"strings"
)

// Diagnostics contains diagnostic information returned by formatter and linter
type Diagnostics struct {
	Success bool               `json:"success"` // overall success (whether all files are formatted properly and have no warnings)
	Files   []*FileDiagnostics `json:"files"`   // diagnostics per file
}

// Format formats a Diagnostics object either as plain text or as json
func (d *Diagnostics) Format(format string, verbose bool) string {
	switch format {
	case "text", "":
		var output strings.Builder
		for _, f := range d.Files {
			for _, w := range f.Warnings {
				formatString := "%s:%d: %s: %s (%s)\n"
				if !w.Actionable {
					formatString = "%s:%d: %s: %s [%s]\n"
				}
				output.WriteString(fmt.Sprintf(formatString,
					f.Filename,
					w.Start.Line,
					w.Category,
					w.Message,
					w.URL))
			}
			if !f.Formatted {
				rewrites := []string{}
				for category, count := range f.Rewrites {
					if count > 0 {
						rewrites = append(rewrites, category)
					}
				}
				log := ""
				if len(rewrites) > 0 {
					sort.Strings(rewrites)
					log = " " + strings.Join(rewrites, " ")
				}
				output.WriteString(fmt.Sprintf("%s # reformat%s\n", f.Filename, log))
			}
		}
		return output.String()
	case "json":
		var result []byte
		if verbose {
			result, _ = json.MarshalIndent(*d, "", "    ")
		} else {
			result, _ = json.Marshal(*d)
		}
		return string(result) + "\n"
	}
	return ""
}

// FileDiagnostics contains diagnostics information for a file
type FileDiagnostics struct {
	Filename  string         `json:"filename"`
	Formatted bool           `json:"formatted"`
	Valid     bool           `json:"valid"`
	Warnings  []*warning     `json:"warnings"`
	Rewrites  map[string]int `json:"rewrites,omitempty"`
}

// SetRewrites adds information about rewrites to the diagnostics
func (fd *FileDiagnostics) SetRewrites(categories map[string]int) {
	for category, count := range categories {
		if count > 0 {
			fd.Rewrites[category] = count
		}
	}
}

type warning struct {
	Start      position `json:"start"`
	End        position `json:"end"`
	Category   string   `json:"category"`
	Actionable bool     `json:"actionable"`
	Message    string   `json:"message"`
	URL        string   `json:"url"`
}

type position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// NewDiagnostics returns a new Diagnostics object
func NewDiagnostics(fileDiagnostics ...*FileDiagnostics) *Diagnostics {
	diagnostics := &Diagnostics{
		Success: true,
		Files:   fileDiagnostics,
	}
	for _, file := range diagnostics.Files {
		if !file.Formatted || len(file.Warnings) > 0 {
			diagnostics.Success = false
			break
		}
	}
	return diagnostics
}

// NewFileDiagnostics returns a new FileDiagnostics object
func NewFileDiagnostics(filename string, warnings []*warn.Finding) *FileDiagnostics {
	fileDiagnostics := FileDiagnostics{
		Filename:  filename,
		Formatted: true,
		Valid:     true,
		Warnings:  []*warning{},
		Rewrites:  map[string]int{},
	}

	for _, w := range warnings {
		fileDiagnostics.Warnings = append(fileDiagnostics.Warnings, &warning{
			Start:      makePosition(w.Start),
			End:        makePosition(w.End),
			Category:   w.Category,
			Actionable: w.Actionable,
			Message:    w.Message,
			URL:        w.URL,
		})
	}

	return &fileDiagnostics
}

// InvalidFileDiagnostics returns a new FileDiagnostics object for an invalid file
func InvalidFileDiagnostics(filename string) *FileDiagnostics {
	fileDiagnostics := &FileDiagnostics{
		Filename:  filename,
		Formatted: false,
		Valid:     false,
		Warnings:  []*warning{},
		Rewrites:  map[string]int{},
	}
	if filename == "" {
		fileDiagnostics.Filename = "<stdin>"
	}
	return fileDiagnostics
}

func makePosition(p build.Position) position {
	return position{
		Line:   p.Line,
		Column: p.LineRune,
	}
}
