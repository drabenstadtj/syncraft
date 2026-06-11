package anvil

import (
	"os"
	"path/filepath"
	"testing"
)

const regionDir = `C:\Users\jack\Github\syncraft\Parsing Test\dimensions\minecraft\overworld\region`

// openRW opens an existing .mca file for reading and writing.
func openRW(path string) (*RegionFile, *os.File, error) {
	r, fRO, err := Open(path)
	if err != nil {
		return nil, nil, err
	}
	fRO.Close()
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, nil, err
	}
	return r, f, nil
}

const testRegion = `C:\Users\jack\Github\syncraft\Parsing Test\dimensions\minecraft\overworld\region\r.0.0.mca`

func TestParseChunk(t *testing.T) {
	region, f, err := Open(testRegion)
	if err != nil {
		t.Fatalf("open region: %v", err)
	}
	defer f.Close()

	data, err := region.ReadChunk(f, 0, 0)
	if err != nil {
		t.Fatalf("read chunk: %v", err)
	}

	chunk, err := ParseChunk(data)
	if err != nil {
		t.Fatalf("parse chunk: %v", err)
	}

	t.Logf("chunk (%d, %d) — %d sections", chunk.X, chunk.Z, len(chunk.Sections))
	for _, s := range chunk.Sections {
		t.Logf("  section Y=%d palette=%d blocks", s.Y, len(s.Palette))
	}
}

// TestDuplicateChunk stamps a 3x3 grid of copied chunks centered on block (-153, -224).
// Block (-153,-224) → chunk (-10,-14) → r.-1.-1.mca local (22, 18).
// The 3x3 spans local x=21..23, z=17..19, all within r.-1.-1.mca.
func TestDuplicateChunk(t *testing.T) {
	// Read source chunk (0,0) from r.0.0.mca to use as stamp data.
	src, fSrc, err := Open(filepath.Join(regionDir, "r.0.0.mca"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := src.ReadRawChunk(fSrc, 0, 0)
	fSrc.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Open existing r.-1.-1.mca for read+write (do not truncate).
	dst, fDst, err := openRW(filepath.Join(regionDir, "r.-1.-1.mca"))
	if err != nil {
		t.Fatal(err)
	}
	defer fDst.Close()

	for lz := 17; lz <= 19; lz++ {
		for lx := 21; lx <= 23; lx++ {
			if err := dst.WriteChunk(fDst, lx, lz, data); err != nil {
				t.Fatalf("write chunk (%d,%d): %v", lx, lz, err)
			}
		}
	}

	t.Logf("wrote 3x3 grid to r.-1.-1.mca — fly to X=-153, Z=-224")
}
