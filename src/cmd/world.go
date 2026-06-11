package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/drabenstadtj/syncraft/src/internal/config"
)

var trackedPatterns = []string{
	"level.dat",
	"playerdata/*.dat",
	"region/r.*.*.mca",
	"entities/r.*.*.mca",
	"poi/r.*.*.mca",
}

// resolveWorld looks up a registered world by name and returns its directory and snapshot directory.
func resolveWorld(name string) (worldDir, snapshotDir string, err error) {
	idx, err := config.LoadIndex()
	if err != nil {
		return "", "", fmt.Errorf("load index: %w", err)
	}
	worldDir, ok := idx.Worlds[name]
	if !ok {
		return "", "", fmt.Errorf("world %q not registered — run init first", name)
	}
	snapshotDir, err = config.SnapshotDir(name)
	if err != nil {
		return "", "", err
	}
	return worldDir, snapshotDir, nil
}

// snapshotWorld copies all tracked files from worldDir into snapshotDir.
func snapshotWorld(worldDir, snapshotDir string) (int, error) {
	var count int
	for _, pattern := range trackedPatterns {
		matches, err := filepath.Glob(filepath.Join(worldDir, pattern))
		if err != nil {
			return 0, err
		}
		for _, src := range matches {
			rel, err := filepath.Rel(worldDir, src)
			if err != nil {
				return 0, err
			}
			dst := filepath.Join(snapshotDir, rel)
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				return 0, err
			}
			if err := copyFile(src, dst); err != nil {
				return 0, fmt.Errorf("copy %s: %w", rel, err)
			}
			count++
		}
	}
	return count, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
