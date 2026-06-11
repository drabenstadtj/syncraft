// nbt decoder
package anvil

import (
	"encoding/binary"
	"fmt"
	"io"
)

type TagType byte

const (
	TagEnd       TagType = 0
	TagByte      TagType = 1
	TagShort     TagType = 2
	TagInt       TagType = 3
	TagLong      TagType = 4
	TagFloat     TagType = 5
	TagDouble    TagType = 6
	TagByteArray TagType = 7
	TagString    TagType = 8
	TagList      TagType = 9
	TagCompound  TagType = 10
	TagIntArray  TagType = 11
	TagLongArray TagType = 12
)

// Decode reads an NBT compound tag from r and returns it as a map.
func Decode(r io.Reader) (map[string]any, error) {
	// the root of a chunk NBT is always a TAG_Compound
	var tagType byte
	if err := binary.Read(r, binary.BigEndian, &tagType); err != nil {
		return nil, err
	}
	if TagType(tagType) != TagCompound {
		return nil, fmt.Errorf("expected root TAG_Compound, got %d", tagType)
	}

	// read (and discard) the root name — always empty in chunk NBT
	if _, err := readName(r); err != nil {
		return nil, err
	}

	return readCompound(r)
}

// readPayload reads the payload for a given tag type.
func readPayload(r io.Reader, t TagType) (any, error) {
	switch t {
	case TagByte:
		var v int8
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagShort:
		var v int16
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagInt:
		var v int32
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagLong:
		var v int64
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagFloat:
		var v float32
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagDouble:
		var v float64
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagByteArray:
		var length int32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		v := make([]byte, length)
		_, err := io.ReadFull(r, v)
		return v, err
	case TagString:
		return readString(r)
	case TagList:
		return readList(r)
	case TagCompound:
		return readCompound(r)
	case TagIntArray:
		var length int32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		v := make([]int32, length)
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	case TagLongArray:
		var length int32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		v := make([]int64, length)
		err := binary.Read(r, binary.BigEndian, &v)
		return v, err
	default:
		return nil, fmt.Errorf("unknown tag type: %d", t)
	}
}

// readCompound reads key/value pairs until TAG_End.
func readCompound(r io.Reader) (map[string]any, error) {
	m := make(map[string]any)
	for {
		var tagType byte
		if err := binary.Read(r, binary.BigEndian, &tagType); err != nil {
			return nil, err
		}
		if TagType(tagType) == TagEnd {
			break
		}
		name, err := readName(r)
		if err != nil {
			return nil, err
		}
		payload, err := readPayload(r, TagType(tagType))
		if err != nil {
			return nil, fmt.Errorf("reading %q: %w", name, err)
		}
		m[name] = payload
	}
	return m, nil
}

// readList reads a list of same-typed payloads.
func readList(r io.Reader) ([]any, error) {
	var elemType byte
	if err := binary.Read(r, binary.BigEndian, &elemType); err != nil {
		return nil, err
	}
	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	list := make([]any, length)
	for i := range list {
		v, err := readPayload(r, TagType(elemType))
		if err != nil {
			return nil, err
		}
		list[i] = v
	}
	return list, nil
}

func readName(r io.Reader) (string, error) {
	var length int16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

func readString(r io.Reader) (string, error) {
	var length int16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	b := make([]byte, length)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}
