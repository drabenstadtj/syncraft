package diff

import (
	"encoding/binary"
	"fmt"
	"io"
)

var magic = [4]byte{'S', 'Y', 'N', 'C'}

const version byte = 0x02

// Encode writes a WorldDiff to w in the SYNC binary format.
func Encode(w io.Writer, wd *WorldDiff) error {
	if _, err := w.Write(magic[:]); err != nil {
		return err
	}
	if _, err := w.Write([]byte{version}); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, uint32(len(wd.Regions))); err != nil {
		return err
	}
	for _, d := range wd.Regions {
		if err := writeString(w, d.Filename); err != nil {
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
			if err := writeBytes(w, c.Data); err != nil {
				return err
			}
		}
	}

	if err := binary.Write(w, binary.BigEndian, uint32(len(wd.Files))); err != nil {
		return err
	}
	for _, f := range wd.Files {
		if err := writeString(w, f.Path); err != nil {
			return err
		}
		if err := writeBytes(w, f.Data); err != nil {
			return err
		}
	}

	return nil
}

// Decode reads a SYNC binary format from r and returns a WorldDiff.
func Decode(r io.Reader) (*WorldDiff, error) {
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

	wd := &WorldDiff{}
	for range regionCount {
		filename, err := readString(r)
		if err != nil {
			return nil, fmt.Errorf("read region filename: %w", err)
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
			chunks[j].Data, err = readBytes(r)
			if err != nil {
				return nil, fmt.Errorf("read chunk data: %w", err)
			}
		}

		wd.Regions = append(wd.Regions, &RegionDiff{Filename: filename, Chunks: chunks})
	}

	var fileCount uint32
	if err := binary.Read(r, binary.BigEndian, &fileCount); err != nil {
		return nil, fmt.Errorf("read file count: %w", err)
	}

	for range fileCount {
		path, err := readString(r)
		if err != nil {
			return nil, fmt.Errorf("read file path: %w", err)
		}
		data, err := readBytes(r)
		if err != nil {
			return nil, fmt.Errorf("read file data: %w", err)
		}
		wd.Files = append(wd.Files, FileDiff{Path: path, Data: data})
	}

	return wd, nil
}

func writeString(w io.Writer, s string) error {
	b := []byte(s)
	if err := binary.Write(w, binary.BigEndian, uint16(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func writeBytes(w io.Writer, b []byte) error {
	if err := binary.Write(w, binary.BigEndian, uint32(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func readString(r io.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

func readBytes(r io.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	return b, nil
}
