// Package scan performs safe filesystem traversal and file selection for
// analysis (spec §9.3). It honors ignore rules, skips binaries and oversized
// files, and prioritizes documentation, manifests, CI, and entry points.
package scan

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// tldg-eca

// Category classifies a selected file by role.
type Category string

const (
	CatReadme     Category = "readme"
	CatLicense    Category = "license"
	CatChangelog  Category = "changelog"
	CatContrib    Category = "contributing"
	CatSecurity   Category = "security"
	CatManifest   Category = "manifest"
	CatLockfile   Category = "lockfile"
	CatCI         Category = "ci"
	CatDocker     Category = "docker"
	CatBuild      Category = "build"
	CatDoc        Category = "doc"
	CatADR        Category = "adr"
	CatEntrypoint Category = "entrypoint"
	CatTest       Category = "test"
	CatSource     Category = "source"
	CatConfig     Category = "config"
	CatOther      Category = "other"
)

// File is a selected file with classification metadata.
type File struct {
	RelPath  string
	AbsPath  string
	Size     int64
	Category Category
	Priority int // higher = more important for summaries
}

// Result is the outcome of a scan.
type Result struct {
	Root     string
	Total    int
	Selected []File
	Skipped  []string // human-readable skip reasons
}

// Options controls selection.
type Options struct {
	IgnoreNames   []string // path segments to skip (e.g. .git, node_modules)
	MaxFileSizeKB int
	FollowSymlinks bool
}

// binaryExts are treated as binary regardless of content sniffing.
var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
	".pdf": true, ".zip": true, ".gz": true, ".tar": true, ".bz2": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".a": true,
	".o": true, ".class": true, ".jar": true, ".wasm": true, ".bin": true,
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".mp3": true,
	".mp4": true, ".mov": true, ".webp": true, ".svg": false,
}

// WalkDir scans root by walking the filesystem (used for non-Git directories).
func WalkDir(root string, opts Options) (*Result, error) {
	res := &Result{Root: root}
	ignore := toSet(opts.IgnoreNames)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // tolerate unreadable entries
		}
		if path == root {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if ignore[name] {
				return filepath.SkipDir
			}
			return nil
		}
		if !opts.FollowSymlinks && d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		consider(res, root, rel, opts)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortSelected(res)
	return res, nil
}

// SelectFiles classifies an explicit candidate list (e.g. from git ls-files),
// applying size and binary filters. rels are relative to root.
func SelectFiles(root string, rels []string, opts Options) *Result {
	res := &Result{Root: root}
	ignore := toSet(opts.IgnoreNames)
	for _, rel := range rels {
		if hasIgnoredSegment(rel, ignore) {
			continue
		}
		consider(res, root, rel, opts)
	}
	sortSelected(res)
	return res
}

// consider evaluates one candidate file and appends it if selectable.
func consider(res *Result, root, rel string, opts Options) {
	res.Total++
	abs := filepath.Join(root, rel)
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return
	}
	maxBytes := int64(opts.MaxFileSizeKB) * 1024
	if maxBytes > 0 && info.Size() > maxBytes {
		res.Skipped = append(res.Skipped, rel+" (oversized)")
		return
	}
	if isBinary(abs) {
		res.Skipped = append(res.Skipped, rel+" (binary)")
		return
	}
	cat, prio := Classify(rel)
	res.Selected = append(res.Selected, File{
		RelPath:  rel,
		AbsPath:  abs,
		Size:     info.Size(),
		Category: cat,
		Priority: prio,
	})
}

// isBinary sniffs the first bytes for a NUL byte.
func isBinary(abs string) bool {
	ext := strings.ToLower(filepath.Ext(abs))
	if b, ok := binaryExts[ext]; ok {
		return b
	}
	f, err := os.Open(abs)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	return bytes.IndexByte(buf[:n], 0) >= 0
}

func toSet(names []string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	m[".git"] = true
	return m
}

func hasIgnoredSegment(rel string, ignore map[string]bool) bool {
	for _, seg := range strings.Split(rel, string(filepath.Separator)) {
		if ignore[seg] {
			return true
		}
	}
	return false
}
