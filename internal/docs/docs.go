// Package docs extracts and classifies documentation and manifest files from a
// scan result (spec §9.5). Documentation-sourced claims remain distinguishable
// from implementation-verified ones downstream (spec §4.1).
package docs

import (
	"github.com/lesliesrussell/tldg/internal/repo"
	"github.com/lesliesrussell/tldg/internal/scan"
)

// tldg-eca

// docCategories are the scan categories treated as documentation/manifest
// evidence for the doc index.
var docCategories = map[scan.Category]bool{
	scan.CatReadme: true, scan.CatLicense: true, scan.CatChangelog: true,
	scan.CatContrib: true, scan.CatSecurity: true, scan.CatDoc: true,
	scan.CatADR: true, scan.CatManifest: true, scan.CatCI: true,
	scan.CatDocker: true, scan.CatBuild: true,
}

// Index builds the classified documentation index from a scan result.
func Index(res *scan.Result) []repo.DocFile {
	var out []repo.DocFile
	for _, f := range res.Selected {
		if docCategories[f.Category] {
			out = append(out, repo.DocFile{
				Path:     f.RelPath,
				Category: string(f.Category),
				Bytes:    int(f.Size),
			})
		}
	}
	return out
}

// Readme returns the highest-priority README file from a scan result, if any.
func Readme(res *scan.Result) (scan.File, bool) {
	for _, f := range res.Selected {
		if f.Category == scan.CatReadme {
			return f, true
		}
	}
	return scan.File{}, false
}
