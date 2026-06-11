package diff

import (
	"fmt"
	"os"
	"path/filepath"
)

// Patch applies a WorldDiff to the target world directory.
func Patch(worldDir string, wd *WorldDiff) error {
	for _, d := range wd.Regions {
		target := filepath.Join(worldDir, d.Filename)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", d.Filename, err)
		}
		r, f, err := OpenForWrite(target)
		if err != nil {
			return fmt.Errorf("open %s: %w", d.Filename, err)
		}
		for _, chunk := range d.Chunks {
			if err := r.WriteChunk(f, chunk.X, chunk.Z, chunk.Data); err != nil {
				f.Close()
				return fmt.Errorf("write chunk (%d,%d) in %s: %w", chunk.X, chunk.Z, d.Filename, err)
			}
		}
		f.Close()
	}

	for _, fd := range wd.Files {
		target := filepath.Join(worldDir, fd.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", fd.Path, err)
		}
		if err := os.WriteFile(target, fd.Data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", fd.Path, err)
		}
	}

	return nil
}
