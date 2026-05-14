package render

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderBlocksWithExternalLinks(t *testing.T) {
	layout := Layout{
		{
			ID:      1,
			Desktop: BreakpointSize{Cols: 12, Rows: 4, Gap: 10},
			Tablet:  BreakpointSize{Cols: 8, Rows: 4, Gap: 10},
			Mobile:  BreakpointSize{Cols: 4, Rows: 4, Gap: 10},
			Blocks: []Block{
				{
					ID:   1,
					Type: "Icon",
					Icon: &BlockIcon{Name: "external-link", StrokeWidth: 1.5},
					Link: &BlockLink{Type: "external", URL: "https://example.com/icon"},
					D:    BlockCoords{X: 1, Y: 1, W: 1, H: 1},
					M:    BlockCoords{X: 1, Y: 1, W: 1, H: 1},
					T:    BlockCoords{X: 1, Y: 1, W: 1, H: 1},
				},
				{
					ID:    2,
					Type:  "Image",
					Image: &BlockImage{URLDesk: "/images/photo.jpg"},
					Link:  &BlockLink{Type: "external", URL: "https://example.com/image"},
					Locales: map[string]interface{}{
						"alt": "Test image",
					},
					D: BlockCoords{X: 2, Y: 1, W: 2, H: 2},
					M: BlockCoords{X: 2, Y: 1, W: 2, H: 2},
					T: BlockCoords{X: 2, Y: 1, W: 2, H: 2},
				},
				{
					ID:   3,
					Type: "Button",
					Button: &BlockButton{
						Bg:              ButtonColors{Light: "#000", Dark: "#fff"},
						Hover:           ButtonColors{Light: "#333", Dark: "#ccc"},
						Active:          ButtonColors{Light: "#111", Dark: "#ddd"},
						Focus:           ButtonColors{Light: "rgba(0,0,0,.3)", Dark: "rgba(255,255,255,.3)"},
						TextColor:       ButtonColors{Light: "#fff", Dark: "#000"},
						HoverTextColor:  ButtonColors{Light: "#fff", Dark: "#000"},
						ActiveTextColor: ButtonColors{Light: "#fff", Dark: "#000"},
						BorderColor:     ButtonColors{Light: "#000", Dark: "#fff"},
						Width:           "50",
						Padding:         Sides{T: "10", R: "20", B: "10", L: "20"},
					},
					Link: &BlockLink{Type: "external", URL: "https://example.com/button"},
					Locales: map[string]interface{}{
						"label": "Click me",
					},
					D: BlockCoords{X: 4, Y: 1, W: 2, H: 1},
					M: BlockCoords{X: 4, Y: 1, W: 2, H: 1},
					T: BlockCoords{X: 4, Y: 1, W: 2, H: 1},
				},
			},
		},
	}

	layoutJSON, _ := json.Marshal(layout)
	optsJSON, _ := json.Marshal(SiteOptions{
		LightColor:       "#333",
		DarkColor:        "#f8f9fa",
		LightBackColor:   "#fff",
		DarkBackColor:    "#1a1a1a",
		GlobalFontFamily: Font{Family: "Inter", Weight: 400},
		FontSize:         1,
		LineHeight:       1.5,
		DesktopWidth:     1200,
	})

	htmlOut, err := Render(RenderInput{
		LayoutJSON:  layoutJSON,
		OptionsJSON: optsJSON,
		SEO:         &SEOData{Title: "Test"},
		Locale:      "es",
		MediaBase:   "https://cdn.example.com",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	t.Log("=== FULL HTML ===")
	t.Log(htmlOut)

	t.Run("Icon external link", func(t *testing.T) {
		if !strings.Contains(htmlOut, `href="https://example.com/icon"`) {
			t.Error("Icon link href not found")
		}
		if !strings.Contains(htmlOut, `target="_blank"`) {
			t.Error("target=_blank not found for icon")
		}
		if !strings.Contains(htmlOut, `rel="noopener noreferrer"`) {
			t.Error("rel=noopener noreferrer not found for icon")
		}
		if !strings.Contains(htmlOut, `<a class="b-link icon-wrap"`) {
			t.Error("Icon link tag not found")
		}
	})

	t.Run("Image external link", func(t *testing.T) {
		if !strings.Contains(htmlOut, `href="https://example.com/image"`) {
			t.Error("Image link href not found")
		}
		if !strings.Contains(htmlOut, `<a class="b-link" href="https://example.com/image"`) {
			t.Error("Image link tag not found")
		}
	})

	t.Run("Button external link", func(t *testing.T) {
		if !strings.Contains(htmlOut, `href="https://example.com/button"`) {
			t.Error("Button link href not found")
		}
		if !strings.Contains(htmlOut, `<a class="btn-b" href="https://example.com/button"`) {
			t.Error("Button link tag not found")
		}
	})
}

func TestRenderIconNoLink(t *testing.T) {
	layout := Layout{
		{
			ID:      1,
			Desktop: BreakpointSize{Cols: 12, Rows: 4, Gap: 10},
			Tablet:  BreakpointSize{Cols: 8, Rows: 4, Gap: 10},
			Mobile:  BreakpointSize{Cols: 4, Rows: 4, Gap: 10},
			Blocks: []Block{
				{
					ID:   1,
					Type: "Icon",
					Icon: &BlockIcon{Name: "heart", StrokeWidth: 1},
					D:    BlockCoords{X: 1, Y: 1, W: 1, H: 1},
					M:    BlockCoords{X: 1, Y: 1, W: 1, H: 1},
					T:    BlockCoords{X: 1, Y: 1, W: 1, H: 1},
				},
			},
		},
	}

	layoutJSON, _ := json.Marshal(layout)
	optsJSON, _ := json.Marshal(SiteOptions{
		LightColor:       "#333",
		DarkColor:        "#f8f9fa",
		LightBackColor:   "#fff",
		DarkBackColor:    "#1a1a1a",
		GlobalFontFamily: Font{Family: "Inter", Weight: 400},
		FontSize:         1,
		LineHeight:       1.5,
		DesktopWidth:     1200,
	})

	htmlOut, err := Render(RenderInput{
		LayoutJSON:  layoutJSON,
		OptionsJSON: optsJSON,
		SEO:         &SEOData{Title: "Test"},
		Locale:      "es",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(htmlOut, `<a class="b-link icon-wrap"`) {
		t.Error("Icon without link should NOT have <a> tag")
	}
	if !strings.Contains(htmlOut, `<div class="icon-wrap">`) {
		t.Error("Icon without link should have <div> wrapper")
	}
}

func TestIconNameNormalization(t *testing.T) {
	tests := []struct{ in, want string }{
		{"external-link", "ExternalLink"},
		{"heart", "Heart"},
		{"brand-facebook", "BrandFacebook"},
		{"IconExternalLink", "IconExternalLink"},
		{"ExternalLink", "ExternalLink"},
	}
	for _, tt := range tests {
		got := toPascalCase(tt.in)
		if got != tt.want {
			t.Errorf("toPascalCase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGetIconSVG(t *testing.T) {
	tests := []struct {
		name string
		ok   bool
	}{
		{"IconExternalLink", true},
		{"ExternalLink", true},
		{"external-link", true},
		{"heart", true},
		{"Heart", true},
		{"IconHeart", true},
		{"nonexistent-icon", false},
	}
	for _, tt := range tests {
		svg := GetIconSVG(tt.name, 1)
		if tt.ok && svg == "" {
			t.Errorf("GetIconSVG(%q): expected SVG, got empty", tt.name)
		}
		if !tt.ok && svg != "" {
			t.Errorf("GetIconSVG(%q): expected empty, got SVG", tt.name)
		}
	}
}
