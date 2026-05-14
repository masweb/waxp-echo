package i18n

import (
	"encoding/json"
	"testing"
)

func TestMergePreservesLink(t *testing.T) {
	existing := json.RawMessage(`[{
		"id": 1,
		"desktop": {"cols": 12, "rows": 4, "gap": 10},
		"tablet": {"cols": 8, "rows": 4, "gap": 10},
		"mobile": {"cols": 4, "rows": 4, "gap": 10},
		"blocks": [
			{
				"id": 10,
				"type": "Icon",
				"icon": {"name": "external-link", "strokeWidth": 1.5},
				"link": {"type": "external", "url": "https://example.com"},
				"locales": {"alt": {"es": "Mi icono", "en": "My icon"}},
				"d": {"x": 1, "y": 1, "w": 1, "h": 1},
				"m": {"x": 1, "y": 1, "w": 1, "h": 1},
				"t": {"x": 1, "y": 1, "w": 1, "h": 1},
				"style": {"hideOn": [], "background": {"mode": "none"}}
			},
			{
				"id": 11,
				"type": "Image",
				"image": {"url_desk": "/photo.jpg"},
				"link": {"type": "external", "url": "https://example.com/photo"},
				"locales": {"alt": {"es": "Foto", "en": "Photo"}},
				"d": {"x": 2, "y": 1, "w": 2, "h": 2},
				"m": {"x": 2, "y": 1, "w": 2, "h": 2},
				"t": {"x": 2, "y": 1, "w": 2, "h": 2},
				"style": {"hideOn": [], "background": {"mode": "none"}}
			},
			{
				"id": 12,
				"type": "Button",
				"button": {"bg": {"light": "#000"}},
				"link": {"type": "external", "url": "https://example.com/click"},
				"locales": {"label": {"es": "Pulsar", "en": "Click"}},
				"d": {"x": 4, "y": 1, "w": 2, "h": 1},
				"m": {"x": 4, "y": 1, "w": 2, "h": 1},
				"t": {"x": 4, "y": 1, "w": 2, "h": 1},
				"style": {"hideOn": [], "background": {"mode": "none"}}
			}
		]
	}]`)

	t.Run("incoming without link preserves existing link", func(t *testing.T) {
		incoming := json.RawMessage(`[{
			"id": 1,
			"desktop": {"cols": 12, "rows": 4, "gap": 10},
			"tablet": {"cols": 8, "rows": 4, "gap": 10},
			"mobile": {"cols": 4, "rows": 4, "gap": 10},
			"blocks": [
				{
					"id": 10,
					"type": "Icon",
					"icon": {"name": "external-link", "strokeWidth": 1.5},
					"locales": {"alt": "Icono movido"},
					"d": {"x": 3, "y": 2, "w": 1, "h": 1},
					"m": {"x": 3, "y": 2, "w": 1, "h": 1},
					"t": {"x": 3, "y": 2, "w": 1, "h": 1},
					"style": {"hideOn": [], "background": {"mode": "none"}}
				},
				{
					"id": 11,
					"type": "Image",
					"image": {"url_desk": "/photo.jpg"},
					"locales": {"alt": "Foto"},
					"d": {"x": 2, "y": 1, "w": 2, "h": 2},
					"m": {"x": 2, "y": 1, "w": 2, "h": 2},
					"t": {"x": 2, "y": 1, "w": 2, "h": 2},
					"style": {"hideOn": [], "background": {"mode": "none"}}
				},
				{
					"id": 12,
					"type": "Button",
					"button": {"bg": {"light": "#000"}},
					"locales": {"label": "Pulsar"},
					"d": {"x": 4, "y": 1, "w": 2, "h": 1},
					"m": {"x": 4, "y": 1, "w": 2, "h": 1},
					"t": {"x": 4, "y": 1, "w": 2, "h": 1},
					"style": {"hideOn": [], "background": {"mode": "none"}}
				}
			]
		}]`)

		merged, err := Merge(incoming, existing, "es")
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		t.Logf("Merged: %s", string(merged))

		if !jsonContains(merged, `"url":"https://example.com"`) {
			t.Error("Icon link URL was lost during merge")
		}
		if !jsonContains(merged, `"url":"https://example.com/photo"`) {
			t.Error("Image link URL was lost during merge")
		}
		if !jsonContains(merged, `"url":"https://example.com/click"`) {
			t.Error("Button link URL was lost during merge")
		}
	})

	t.Run("reordered blocks match by id", func(t *testing.T) {
		incoming := json.RawMessage(`[{
			"id": 1,
			"desktop": {"cols": 12, "rows": 4, "gap": 10},
			"tablet": {"cols": 8, "rows": 4, "gap": 10},
			"mobile": {"cols": 4, "rows": 4, "gap": 10},
			"blocks": [
				{
					"id": 12,
					"type": "Button",
					"button": {"bg": {"light": "#000"}},
					"locales": {"label": "Click"},
					"d": {"x": 1, "y": 1, "w": 2, "h": 1},
					"style": {"hideOn": [], "background": {"mode": "none"}}
				},
				{
					"id": 10,
					"type": "Icon",
					"icon": {"name": "external-link", "strokeWidth": 1.5},
					"locales": {"alt": "Icon"},
					"d": {"x": 3, "y": 1, "w": 1, "h": 1},
					"style": {"hideOn": [], "background": {"mode": "none"}}
				}
			]
		}]`)

		merged, err := Merge(incoming, existing, "es")
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		t.Logf("Merged: %s", string(merged))

		if !jsonContains(merged, `"url":"https://example.com"`) {
			t.Error("Icon link URL lost: blocks reordered but id=10 should keep its link")
		}
		if !jsonContains(merged, `"url":"https://example.com/click"`) {
			t.Error("Button link URL lost: blocks reordered but id=12 should keep its link")
		}
	})

	t.Run("new locale preserves other locales", func(t *testing.T) {
		merged, err := Merge(existing, existing, "fr")
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		if !jsonContains(merged, `"en":"My icon"`) {
			t.Error("English locale was lost")
		}
		if !jsonContains(merged, `"es":"Mi icono"`) {
			t.Error("Spanish locale was lost")
		}
	})
}

func jsonContains(data json.RawMessage, substr string) bool {
	return jsonContains_(string(data), substr)
}

func jsonContains_(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
