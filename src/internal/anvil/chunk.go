// chunk nbt parsing
package anvil

import (
	"bytes"
)

type Block struct {
	Name       string
	Properties map[string]string
}

type Section struct {
	Y           int8
	Palette     []Block
	BlockStates []int64
}

type Chunk struct {
	X        int32
	Z        int32
	Sections []Section
}

func ParseChunk(data []byte) (*Chunk, error) {
	tags, err := Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	chunk := &Chunk{
		X: tags["xPos"].(int32),
		Z: tags["zPos"].(int32),
	}

	for _, v := range tags["sections"].([]any) {
		chunk.Sections = append(chunk.Sections, parseSection(v.(map[string]any)))
	}

	return chunk, nil
}

func parseSection(s map[string]any) Section {
	section := Section{Y: s["Y"].(int8)}

	blockStates, ok := s["block_states"].(map[string]any)
	if !ok {
		return section
	}

	if palette, ok := blockStates["palette"]; ok {
		section.Palette = parsePalette(palette.([]any))
	}
	if data, ok := blockStates["data"]; ok {
		section.BlockStates = data.([]int64)
	}

	return section
}

func parsePalette(raw []any) []Block {
	palette := make([]Block, len(raw))
	for i, entry := range raw {
		e := entry.(map[string]any)
		block := Block{Name: e["Name"].(string)}
		if props, ok := e["Properties"]; ok {
			block.Properties = make(map[string]string)
			for k, v := range props.(map[string]any) {
				block.Properties[k] = v.(string)
			}
		}
		palette[i] = block
	}
	return palette
}
