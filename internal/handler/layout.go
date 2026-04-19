package handler

import (
	_ "embed"
	"encoding/json"
)

//go:embed default_section.json
var defaultSectionJSON []byte

type sectionTemplate struct {
	Blocks  []interface{}  `json:"blocks"`
	Mobile  map[string]int `json:"mobile"`
	Tablet  map[string]int `json:"tablet"`
	Desktop map[string]int `json:"desktop"`
	Style   interface{}    `json:"style"`
}

var parsedSectionTemplate sectionTemplate

func init() {
	if err := json.Unmarshal(defaultSectionJSON, &parsedSectionTemplate); err != nil {
		panic("failed to parse default_section.json: " + err.Error())
	}
}

func makeDefaultSection(id int64) map[string]interface{} {
	section := map[string]interface{}{
		"id":      id,
		"blocks":  parsedSectionTemplate.Blocks,
		"mobile":  parsedSectionTemplate.Mobile,
		"tablet":  parsedSectionTemplate.Tablet,
		"desktop": parsedSectionTemplate.Desktop,
	}
	if parsedSectionTemplate.Style != nil {
		section["style"] = parsedSectionTemplate.Style
	}
	return section
}

type sectionIDGenerator func() (int64, error)

func makeDefaultLayout(genID sectionIDGenerator, count int) ([]byte, error) {
	layout := make([]map[string]interface{}, count)
	for i := range layout {
		id, err := genID()
		if err != nil {
			return nil, err
		}
		layout[i] = makeDefaultSection(id)
	}
	return json.Marshal(layout)
}
