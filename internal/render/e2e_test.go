package render

import (
	"encoding/json"
	"strings"
	"testing"

	"waxp/echo/internal/i18n"
)

func TestEndToEndLinkFlow(t *testing.T) {
	opts := SiteOptions{
		LightColor:       "#333",
		DarkColor:        "#f8f9fa",
		LightBackColor:   "#fff",
		DarkBackColor:    "#1a1a1a",
		GlobalFontFamily: Font{Family: "Inter", Weight: 400},
		FontSize:         1,
		LineHeight:       1.5,
		DesktopWidth:     1200,
	}
	optsJSON, _ := json.Marshal(opts)

	existingInDB := json.RawMessage(`[{
		"id": 1,
		"desktop": {"cols": 12, "rows": 4, "gap": 10},
		"tablet": {"cols": 8, "rows": 4, "gap": 10},
		"mobile": {"cols": 4, "rows": 4, "gap": 10},
		"blocks": [
			{
				"id": 10,
				"type": "Icon",
				"icon": {"name": "external-link", "strokeWidth": 1.5},
				"link": {"type": "external", "url": "https://example.com/icon"},
				"locales": {"alt": {"es": "Mi icono"}},
				"d": {"x": 1, "y": 1, "w": 1, "h": 1},
				"m": {"x": 1, "y": 1, "w": 1, "h": 1},
				"t": {"x": 1, "y": 1, "w": 1, "h": 1},
				"style": {}
			},
			{
				"id": 11,
				"type": "Image",
				"image": {"url_desk": "/photo.jpg", "url_tab": "", "url_mob": "", "url_desk_dark": "", "url_tab_dark": "", "url_mob_dark": "", "fit": "", "lightbox": false},
				"link": {"type": "internal", "url": "/about"},
				"locales": {"alt": {"es": "Foto"}},
				"d": {"x": 2, "y": 1, "w": 2, "h": 2},
				"m": {"x": 2, "y": 1, "w": 2, "h": 2},
				"t": {"x": 2, "y": 1, "w": 2, "h": 2},
				"style": {}
			},
			{
				"id": 12,
				"type": "Menu",
				"menu": {"es": [{"label": "Google", "link": {"type": "external", "url": "https://google.com"}}]},
				"menuColors": {"color":{"light":"#333","dark":"#f8f9fa"},"hover":{"light":"#0d6efd","dark":"#6ea8fe"},"active":{"light":"#0a58ca","dark":"#5aa4f0"}},
				"locales": {},
				"d": {"x": 6, "y": 1, "w": 4, "h": 1},
				"m": {"x": 1, "y": 1, "w": 4, "h": 1},
				"t": {"x": 6, "y": 1, "w": 4, "h": 1},
				"style": {}
			}
		]
	}]`)

	incomingFromFrontend := json.RawMessage(`[{
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
				"style": {}
			},
			{
				"id": 11,
				"type": "Image",
				"image": {"url_desk": "/photo.jpg", "url_tab": "", "url_mob": "", "url_desk_dark": "", "url_tab_dark": "", "url_mob_dark": "", "fit": "", "lightbox": false},
				"locales": {"alt": "Foto"},
				"d": {"x": 2, "y": 1, "w": 2, "h": 2},
				"m": {"x": 2, "y": 1, "w": 2, "h": 2},
				"t": {"x": 2, "y": 1, "w": 2, "h": 2},
				"style": {}
			},
			{
				"id": 12,
				"type": "Menu",
				"menu": [{"label": "Google", "link": {"type": "external", "url": "https://google.com"}}],
				"menuColors": {"color":{"light":"#333","dark":"#f8f9fa"},"hover":{"light":"#0d6efd","dark":"#6ea8fe"},"active":{"light":"#0a58ca","dark":"#5aa4f0"}},
				"locales": {},
				"d": {"x": 6, "y": 1, "w": 4, "h": 1},
				"m": {"x": 1, "y": 1, "w": 4, "h": 1},
				"t": {"x": 6, "y": 1, "w": 4, "h": 1},
				"style": {}
			}
		]
	}]`)

	t.Log("=== STEP 1: i18n.Merge ===")
	merged, err := i18n.Merge(incomingFromFrontend, existingInDB, "es")
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	t.Logf("Merged:\n%s", prettyJSON(t, merged))

	t.Log("\n=== STEP 2: i18n.Resolve ===")
	resolved, err := i18n.Resolve(merged, "es")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	t.Logf("Resolved:\n%s", prettyJSON(t, resolved))

	t.Log("\n=== STEP 3: ParseLayout - check link fields ===")
	layout, err := ParseLayout(resolved)
	if err != nil {
		t.Fatalf("ParseLayout failed: %v", err)
	}
	for _, sec := range layout {
		for _, blk := range sec.Blocks {
			linkStr := "nil"
			if blk.Link != nil {
				linkStr = "{type=" + blk.Link.Type + ", url=" + blk.Link.URL + "}"
			}
			t.Logf("  Block id=%d type=%-6s link=%s", blk.ID, blk.Type, linkStr)
		}
	}

	t.Log("\n=== STEP 4: Render ===")
	htmlOut, err := Render(RenderInput{
		LayoutJSON:  resolved,
		OptionsJSON: optsJSON,
		SEO:         &SEOData{Title: "Test"},
		Locale:      "es",
		MediaBase:   "https://cdn.example.com",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	t.Run("Icon_link", func(t *testing.T) {
		has := strings.Contains(htmlOut, `href="https://example.com/icon"`)
		t.Logf("Icon link in HTML: %v", has)
		if !has {
			t.Error("Icon link NOT found")
		}
	})
	t.Run("Image_link", func(t *testing.T) {
		has := strings.Contains(htmlOut, `href="/about"`)
		t.Logf("Image link in HTML: %v", has)
		if !has {
			t.Error("Image link NOT found")
		}
	})
	t.Run("Menu_link", func(t *testing.T) {
		has := strings.Contains(htmlOut, `href="https://google.com"`)
		t.Logf("Menu link in HTML: %v", has)
		if !has {
			t.Error("Menu link NOT found")
		}
	})
}

func prettyJSON(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	out, _ := json.MarshalIndent(v, "", "  ")
	return string(out)
}
