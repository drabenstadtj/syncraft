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

// FileDiff holds the full contents of a changed non-region file.
type FileDiff struct {
	Path string // relative to world root (e.g. "level.dat", "playerdata/abc.dat")
	Data []byte
}

// WorldDiff is the top-level diff produced by comparing two world states.
type WorldDiff struct {
	Regions []*RegionDiff
	Files   []FileDiff
}

// mcaPatterns lists subdirs that contain region files to diff.
var mcaPatterns = []string{"region", "entities", "poi"}

// filePatterns lists glob patterns (relative to world root) for whole-file diffs.
var filePatterns = []string{
	"level.dat",
	"playerdata/*.dat",
}

// DiffWorld compares two world directories and returns a WorldDiff.
func DiffWorld(dirA, dirB string) (*WorldDiff, error) {
	wd := &WorldDiff{}

	for _, sub := range mcaPatterns {
		matches, err := filepath.Glob(filepath.Join(dirB, sub, "r.*.*.mca"))
		if err != nil {
			return nil, err
		}
		for _, pathB := range matches {
			rel, _ := filepath.Rel(dirB, pathB)
			pathA := filepath.Join(dirA, rel)

			var d *RegionDiff
			if _, err := os.Stat(pathA); os.IsNotExist(err) {
				d, err = diffAgainstEmpty(pathB, rel)
				if err != nil {
					return nil, fmt.Errorf("diff %s: %w", rel, err)
				}
			} else {
				d, err = DiffRegion(pathA, pathB)
				if err != nil {
					return nil, fmt.Errorf("diff %s: %w", rel, err)
				}
			}
			if len(d.Chunks) > 0 {
				wd.Regions = append(wd.Regions, d)
			}
		}
	}

	for _, pattern := range filePatterns {
		matches, err := filepath.Glob(filepath.Join(dirB, pattern))
		if err != nil {
			return nil, err
		}
		for _, pathB := range matches {
			rel, _ := filepath.Rel(dirB, pathB)
			pathA := filepath.Join(dirA, rel)

			dataB, err := os.ReadFile(pathB)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", rel, err)
			}

			dataA, err := os.ReadFile(pathA)
			if os.IsNotExist(err) || !bytes.Equal(dataA, dataB) {
				wd.Files = append(wd.Files, FileDiff{Path: rel, Data: dataB})
			} else if err != nil {
				return nil, fmt.Errorf("read snapshot %s: %w", rel, err)
			}
		}
	}

	return wd, nil
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

	rel, _ := filepath.Rel(filepath.Dir(pathB), pathB)
	d := &RegionDiff{Filename: rel}

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
				d.Chunks = append(d.Chunks, ChunkDiff{X: x, Z: z, Data: b})
			}
		}
	}

	return d, nil
}

// diffAgainstEmpty returns all present chunks in a region file as a diff.
func diffAgainstEmpty(path, rel string) (*RegionDiff, error) {
	r, f, err := anvil.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := &RegionDiff{Filename: rel}
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
