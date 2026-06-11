// mca file parsing
package anvil

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type RegionFile struct {
	offsets    [1024]uint32
	timestamps [1024]uint32
}

// Open reads the header of an .mca file into a RegionFile.
// Files smaller than 8KB are treated as empty regions.
func Open(path string) (*RegionFile, *os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	if info.Size() < 8192 {
		return &RegionFile{}, f, nil
	}

	var r RegionFile
	if err := binary.Read(f, binary.BigEndian, &r.offsets); err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("reading offsets: %w", err)
	}
	if err := binary.Read(f, binary.BigEndian, &r.timestamps); err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("reading timestamps: %w", err)
	}

	return &r, f, nil
}

// chunkIndex returns the header table index for a chunk at region-local coords.
func chunkIndex(x, z int) int {
	return x + z*32
}

// ReadChunk returns the decompressed NBT bytes for the chunk at (x, z).
// x and z must be in the range [0, 31].
func (r *RegionFile) ReadChunk(f *os.File, x, z int) ([]byte, error) {
	idx := chunkIndex(x, z)
	entry := r.offsets[idx]

	if entry == 0 {
		return nil, fmt.Errorf("chunk (%d, %d) not present in region", x, z)
	}

	// offset is in 4096-byte sectors; lower byte is sector count (unused here)
	sectorOffset := int64((entry >> 8) * 4096)

	if _, err := f.Seek(sectorOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking to chunk (%d, %d): %w", x, z, err)
	}

	// 4-byte length, 1-byte compression type
	var length uint32
	if err := binary.Read(f, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("reading chunk length: %w", err)
	}

	var compression byte
	if err := binary.Read(f, binary.BigEndian, &compression); err != nil {
		return nil, fmt.Errorf("reading compression type: %w", err)
	}

	compressed := make([]byte, length-1)
	if _, err := io.ReadFull(f, compressed); err != nil {
		return nil, fmt.Errorf("reading chunk data: %w", err)
	}

	switch compression {
	case 2: // zlib
		rc, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("zlib open: %w", err)
		}
		defer rc.Close()
		return io.ReadAll(rc)
	default:
		return nil, fmt.Errorf("unsupported compression type: %d", compression)
	}
}

// ReadRawChunk returns the raw compressed bytes for the chunk at (x, z).
func (r *RegionFile) ReadRawChunk(f *os.File, x, z int) ([]byte, error) {
	idx := chunkIndex(x, z)
	entry := r.offsets[idx]
	if entry == 0 {
		return nil, nil // chunk not present
	}

	sectorOffset := int64((entry >> 8) * 4096)
	if _, err := f.Seek(sectorOffset, io.SeekStart); err != nil {
		return nil, err
	}

	var length uint32
	if err := binary.Read(f, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	_, err := io.ReadFull(f, data)
	return data, err
}

// WriteChunk writes raw compressed chunk bytes to the region file and updates the offset table.
func (r *RegionFile) WriteChunk(f *os.File, x, z int, data []byte) error {
	// find end of file to append the chunk
	end, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// pad start to sector boundary
	if end%4096 != 0 {
		padding := 4096 - (end % 4096)
		f.Write(make([]byte, padding))
		end += padding
	}

	sectorStart := uint32(end / 4096)

	// write: 4-byte length, then data (data[0] is compression type, rest is compressed bytes)
	if err := binary.Write(f, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}

	// pad to sector boundary
	written := 4 + len(data)
	if written%4096 != 0 {
		f.Write(make([]byte, 4096-(written%4096)))
	}

	sectorCount := uint32((written + 4095) / 4096)

	// update offset table entry
	idx := chunkIndex(x, z)
	r.offsets[idx] = (sectorStart << 8) | (sectorCount & 0xFF)

	// write updated offset table back to file header
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return binary.Write(f, binary.BigEndian, &r.offsets)
}

// WalkRegionDir calls fn for every .mca file found in dir.
func WalkRegionDir(dir string, fn func(path string) error) error {
	matches, err := filepath.Glob(filepath.Join(dir, "r.*.*.mca"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := fn(path); err != nil {
			return err
		}
	}
	return nil
}
