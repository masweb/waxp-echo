package render

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"sync"
)

//go:embed icons.json
var iconsJSON []byte

var (
	iconData     map[string]iconEntry
	iconDataOnce sync.Once
)

type iconEntry struct {
	Type  string   `json:"type"`
	Paths []string `json:"paths"`
}

func loadIcons() {
	iconDataOnce.Do(func() {
		var raw map[string]iconEntry
		if err := json.Unmarshal(iconsJSON, &raw); err != nil {
			iconData = make(map[string]iconEntry)
			return
		}
		iconData = raw
	})
}

func GetIconSVG(name string, strokeWidth float64) string {
	loadIcons()
	entry, ok := iconData[name]
	if !ok {
		return ""
	}

	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24"`
	if entry.Type == "filled" {
		svg += ` fill="currentColor">`
	} else {
		if strokeWidth <= 0 {
			strokeWidth = 1
		}
		svg += ` fill="none" stroke="currentColor" stroke-width="` + strconv.FormatFloat(strokeWidth, 'f', -1, 64) + `" stroke-linecap="round" stroke-linejoin="round">`
	}

	for _, p := range entry.Paths {
		svg += `<path d="` + p + `"/>`
	}

	svg += `</svg>`
	return svg
}
