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
	"dimensions/minecraft/overworld/region/r.*.*.mca",
	"dimensions/minecraft/overworld/entities/r.*.*.mca",
	"dimensions/minecraft/overworld/poi/r.*.*.mca",
	"dimensions/minecraft/the_nether/region/r.*.*.mca",
	"dimensions/minecraft/the_end/region/r.*.*.mca",
}

// findWorldDir resolves a world name or path to an absolute world directory.
// If name looks like an existing path it is used directly; otherwise it is
// looked up by folder name under the default Minecraft saves directory.
func findWorldDir(name string) (string, error) {
	if info, err := os.Stat(name); err == nil && info.IsDir() {
		return filepath.Abs(name)
	}

	savesDir, err := config.DefaultSavesDir()
	if err != nil {
		return "", err
	}
	worldDir := filepath.Join(savesDir, name)
	if _, err := os.Stat(filepath.Join(worldDir, "level.dat")); err != nil {
		return "", fmt.Errorf("world %q not found in %s", name, savesDir)
	}
	return worldDir, nil
}

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

func formatSize(n int) string {
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/1024/1024)
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d B", n)
	}
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
