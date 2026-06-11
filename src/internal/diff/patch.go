package diff

import "fmt"

// Patch applies a RegionDiff to the target .mca file at path.
func Patch(path string, d *RegionDiff) error {
	r, f, err := OpenForWrite(path)
	if err != nil {
		return fmt.Errorf("open target: %w", err)
	}
	defer f.Close()

	for _, chunk := range d.Chunks {
		if err := r.WriteChunk(f, chunk.X, chunk.Z, chunk.Data); err != nil {
			return fmt.Errorf("write chunk (%d,%d): %w", chunk.X, chunk.Z, err)
		}
	}

	return nil
}
