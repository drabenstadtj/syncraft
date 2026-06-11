package diff

import (
	"encoding/binary"
	"fmt"
	"io"
)

var magic = [4]byte{'S', 'Y', 'N', 'C'}

const version byte = 0x01

// Encode writes a list of RegionDiffs to w in the SYNC binary format.
func Encode(w io.Writer, diffs []*RegionDiff) error {
	if _, err := w.Write(magic[:]); err != nil {
		return err
	}
	if _, err := w.Write([]byte{version}); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, uint32(len(diffs))); err != nil {
		return err
	}

	for _, d := range diffs {
		name := []byte(d.Filename)
		if err := binary.Write(w, binary.BigEndian, uint16(len(name))); err != nil {
			return err
		}
		if _, err := w.Write(name); err != nil {
			return err
		}

		if err := binary.Write(w, binary.BigEndian, uint32(len(d.Chunks))); err != nil {
			return err
		}

		for _, c := range d.Chunks {
			if err := binary.Write(w, binary.BigEndian, int32(c.X)); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, int32(c.Z)); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, uint32(len(c.Data))); err != nil {
				return err
			}
			if _, err := w.Write(c.Data); err != nil {
				return err
			}
		}
	}

	return nil
}

// Decode reads a SYNC binary format from r and returns the list of RegionDiffs.
func Decode(r io.Reader) ([]*RegionDiff, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if hdr != magic {
		return nil, fmt.Errorf("invalid magic: %x", hdr)
	}

	var ver byte
	if err := binary.Read(r, binary.BigEndian, &ver); err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	if ver != version {
		return nil, fmt.Errorf("unsupported version: %d", ver)
	}

	var regionCount uint32
	if err := binary.Read(r, binary.BigEndian, &regionCount); err != nil {
		return nil, fmt.Errorf("read region count: %w", err)
	}

	diffs := make([]*RegionDiff, regionCount)
	for i := range diffs {
		var nameLen uint16
		if err := binary.Read(r, binary.BigEndian, &nameLen); err != nil {
			return nil, fmt.Errorf("read filename length: %w", err)
		}
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, fmt.Errorf("read filename: %w", err)
		}

		var chunkCount uint32
		if err := binary.Read(r, binary.BigEndian, &chunkCount); err != nil {
			return nil, fmt.Errorf("read chunk count: %w", err)
		}

		chunks := make([]ChunkDiff, chunkCount)
		for j := range chunks {
			if err := binary.Read(r, binary.BigEndian, &chunks[j].X); err != nil {
				return nil, fmt.Errorf("read chunk X: %w", err)
			}
			if err := binary.Read(r, binary.BigEndian, &chunks[j].Z); err != nil {
				return nil, fmt.Errorf("read chunk Z: %w", err)
			}
			var dataLen uint32
			if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
				return nil, fmt.Errorf("read data length: %w", err)
			}
			chunks[j].Data = make([]byte, dataLen)
			if _, err := io.ReadFull(r, chunks[j].Data); err != nil {
				return nil, fmt.Errorf("read chunk data: %w", err)
			}
		}

		diffs[i] = &RegionDiff{Filename: string(name), Chunks: chunks}
	}

	return diffs, nil
}
