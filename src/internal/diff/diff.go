package diff

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drabenstadtj/syncraft/src/internal/anvil"
)

// ChunkDiff represents a changed chunk — raw compressed bytes ready to write.
type ChunkDiff struct {
	X, Z int
	Data []byte // compression type byte + compressed NBT
}

// RegionDiff holds all changed chunks for one .mca file.
type RegionDiff struct {
	Filename string
	Chunks   []ChunkDiff
}

// DiffRegion compares two .mca files and returns the chunks that differ.
func DiffRegion(pathA, pathB string) (*RegionDiff, error) {
	rA, fA, err := anvil.Open(pathA)
	if err != nil {
		return nil, fmt.Errorf("open A: %w", err)
	}
	defer fA.Close()

	rB, fB, err := anvil.Open(pathB)
	if err != nil {
		return nil, fmt.Errorf("open B: %w", err)
	}
	defer fB.Close()

	diff := &RegionDiff{Filename: pathB}

	for z := 0; z < 32; z++ {
		for x := 0; x < 32; x++ {
			a, err := rA.ReadRawChunk(fA, x, z)
			if err != nil {
				return nil, fmt.Errorf("read chunk A (%d,%d): %w", x, z, err)
			}
			b, err := rB.ReadRawChunk(fB, x, z)
			if err != nil {
				return nil, fmt.Errorf("read chunk B (%d,%d): %w", x, z, err)
			}

			if !bytes.Equal(a, b) && b != nil {
				diff.Chunks = append(diff.Chunks, ChunkDiff{X: x, Z: z, Data: b})
			}
		}
	}

	return diff, nil
}

// DiffWorld compares two world region directories and returns one RegionDiff per changed file.
func DiffWorld(dirA, dirB string) ([]*RegionDiff, error) {
	matches, err := filepath.Glob(filepath.Join(dirB, "r.*.*.mca"))
	if err != nil {
		return nil, err
	}

	var diffs []*RegionDiff
	for _, pathB := range matches {
		pathA := filepath.Join(dirA, filepath.Base(pathB))

		// if the file doesn't exist in A, treat every chunk in B as new
		if _, err := os.Stat(pathA); os.IsNotExist(err) {
			pathA = ""
		}

		var d *RegionDiff
		if pathA == "" {
			d, err = diffAgainstEmpty(pathB)
		} else {
			d, err = DiffRegion(pathA, pathB)
		}
		if err != nil {
			return nil, fmt.Errorf("diff %s: %w", filepath.Base(pathB), err)
		}
		if len(d.Chunks) > 0 {
			diffs = append(diffs, d)
		}
	}

	return diffs, nil
}

// diffAgainstEmpty returns all present chunks in a region file as a diff.
func diffAgainstEmpty(path string) (*RegionDiff, error) {
	r, f, err := anvil.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := &RegionDiff{Filename: filepath.Base(path)}
	for z := 0; z < 32; z++ {
		for x := 0; x < 32; x++ {
			data, err := r.ReadRawChunk(f, x, z)
			if err != nil {
				return nil, err
			}
			if data != nil {
				d.Chunks = append(d.Chunks, ChunkDiff{X: x, Z: z, Data: data})
			}
		}
	}
	return d, nil
}

// OpenForWrite opens an .mca file for reading and writing, creating it if needed.
func OpenForWrite(path string) (*anvil.RegionFile, *os.File, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, err
	}

	// ensure file has at least an 8KB header
	info, _ := f.Stat()
	if info.Size() < 8192 {
		f.Write(make([]byte, 8192))
	}

	r, _, err := anvil.Open(path)
	if err != nil {
		f.Close()
		return nil, nil, err
	}

	return r, f, nil
}
