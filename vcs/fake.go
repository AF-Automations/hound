package vcs

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

func init() {
	Register(newFake, "fake")
}

type FakeDriver struct {
	DetectChanges bool     `json:"detect-changes"`
	IgnoredFiles  []string `json:"ignored-files"`
}

func newFake(b []byte) (Driver, error) {
	d := FakeDriver{}

	if b != nil {
		if err := json.Unmarshal(b, &d); err != nil {
			return nil, err
		}
	}

	return &d, nil
}

// Adapted from: https://github.com/golang/mod/blob/ad6fd61f94f8fdf6926f5dee6e45bdd13add2f9f/sumdb/dirhash/hash.go#L44
func myHash(files []string, open func(string) (io.ReadCloser, error)) (string, error) {
	h := sha1.New()
	files = append([]string(nil), files...)
	sort.Strings(files)
	for _, file := range files {
		if strings.Contains(file, "\n") {
			return "", errors.New("dirhash: filenames with newlines are not supported")
		}
		r, err := open(file)
		if err != nil {
			return "", err
		}
		hf := sha1.New()
		_, err = io.Copy(hf, r)
		r.Close()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%x  %s\n", hf.Sum(nil), file)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (g *FakeDriver) HeadRev(dir string) (string, error) {
	if !g.DetectChanges {
		idx := strings.LastIndex(dir, "vcs-")
		if idx == -1 {
			return "", fmt.Errorf("could not find 'vcs-' in path: %s", dir)
		}
		return dir[idx+4:], nil
	}

	// Resolve the symbolic link
	if fi, err := os.Stat(dir); err == nil && fi.Mode()|os.ModeSymlink != 0 {
		if s, err := os.Readlink(dir); err == nil {
			dir = s
		}
	}

	// Hash all files in the directory
	// TODO: implement logic from indexAllFiles to skip ignored files?
	hash, err := dirhash.HashDir(dir, "", myHash)
	return hash, err
}

func (g *FakeDriver) Pull(dir string) (string, error) {
	return g.HeadRev(dir)
}

func (g *FakeDriver) Clone(dir, url string) (string, error) {
	src := strings.TrimPrefix(url, "file://")
	if !strings.HasPrefix(url, "file://") {
		return "", fmt.Errorf("expected 'file://' prefix in url: %s", url)
	}
	if err := os.Symlink(src, dir); err != nil {
		return "", err
	}
	return g.Pull(dir)
}

func (g *FakeDriver) SpecialFiles() []string {
	return g.IgnoredFiles
}

func (g *FakeDriver) AutoGeneratedFiles(dir string) []string {
	return []string{}
}
