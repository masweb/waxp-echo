package render

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

func buildCSS(sections []sectionRender, opts SiteOptions) string {
	var b strings.Builder

	b.WriteString("*{margin:0;padding:0;box-sizing:border-box;}")
	b.WriteString("html{scroll-behavior:smooth;}")
	b.WriteString("body{font-family:var(--waxp-ff);-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale;text-rendering:optimizeLegibility;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%;}")
	b.WriteString("::-webkit-scrollbar{width:4px;height:4px;}")

	lightAccent := opts.LightAccentColor
	if lightAccent == "" {
		lightAccent = opts.LightColor
	}
	darkAccent := opts.DarkAccentColor
	if darkAccent == "" {
		darkAccent = opts.DarkColor
	}

	fmt.Fprintf(&b, ":root{--waxp-text:%s;--waxp-bg:%s;--waxp-accent:%s;--waxp-ff:'%s',sans-serif;--waxp-fw:%d;--waxp-fs:%grem;--waxp-lh:%gem;--waxp-dw:%dpx;}",
		opts.LightColor, opts.LightBackColor, lightAccent,
		opts.GlobalFontFamily.Family, opts.GlobalFontFamily.Weight,
		opts.FontSize, opts.LineHeight, opts.DesktopWidth,
	)

	fmt.Fprintf(&b, ":root[data-theme=\"dark\"]{--waxp-text:%s;--waxp-bg:%s;--waxp-accent:%s;}",
		opts.DarkColor, opts.DarkBackColor, darkAccent,
	)

	fmt.Fprintf(&b, ".waxp{color:var(--waxp-text);background:var(--waxp-bg);font-family:var(--waxp-ff);font-weight:var(--waxp-fw);min-height:100vh;display:flex;flex-direction:column;")
	if opts.DesktopMargin > 0 {
		fmt.Fprintf(&b, "padding:0 %gpx;", opts.DesktopMargin)
	}
	b.WriteString("}")

	writeBaseBlockCSS(&b)

	if opts.Headers != nil {
		writeHeaderCSS(&b, opts.Headers)
	}

	for _, sec := range sections {
		writeSectionCSS(&b, sec, opts)
	}

	mobileBP := opts.MobileBP
	if mobileBP == 0 {
		mobileBP = 767
	}
	tabletBP := opts.TabletBP
	if tabletBP == 0 {
		tabletBP = 1024
	}

	if opts.TabletMargin > 0 {
		fmt.Fprintf(&b, "@media(max-width:%dpx){.waxp{padding:0 %gpx;}}", tabletBP, opts.TabletMargin)
	}
	if opts.MobileMargin > 0 {
		fmt.Fprintf(&b, "@media(max-width:%dpx){.waxp{padding:0 %gpx;}}", mobileBP, opts.MobileMargin)
	}
	fmt.Fprintf(&b, "@media(max-width:%dpx){.lang-select{font-size:16px!important;width:auto!important;min-width:80px;height:auto!important;appearance:auto!important;-webkit-appearance:auto!important;padding:4px 8px!important;text-align:left!important;text-align-last:auto!important;}}", mobileBP)

	return b.String()
}

func writeBaseBlockCSS(b *strings.Builder) {
	b.WriteString(".b{position:relative;display:flex;}")
	b.WriteString(".b-clip{overflow:hidden;}")
	b.WriteString(".b-overlay{position:absolute;inset:0;pointer-events:none;z-index:0;}")
	b.WriteString(".b-inner{position:relative;z-index:1;width:100%;height:100%;display:flex;align-items:start;}")
	b.WriteString(".b-center{align-items:center!important;justify-content:center!important;}")
	b.WriteString(".b-link{color:inherit;text-decoration:none;}")
	b.WriteString(".b-link:hover{opacity:0.85;}")
	b.WriteString(".sr{display:flex;justify-content:center;}")
	b.WriteString(".s{display:grid;margin:0 auto;width:100%;}")
	b.WriteString(".s-overlay{position:absolute;inset:0;pointer-events:none;z-index:0;}")

	b.WriteString(".b-tiptap>*+*{margin-top:.65em;}")
	b.WriteString(".b-tiptap :last-child{margin-bottom:0;}")
	b.WriteString(".b-tiptap img{max-width:100%;height:auto;}")
	b.WriteString(".b-tiptap a{color:var(--waxp-accent);text-decoration:underline;}")
	b.WriteString(".b-tiptap ul,.b-tiptap ol{margin-left:1.5em;}")
	b.WriteString(".b-tiptap blockquote{padding-left:1em;border-left:3px solid currentColor;opacity:.7;}")

	b.WriteString(".btn-wrap{display:flex;align-items:center;justify-content:center;width:100%;height:100%;position:relative;z-index:1;}")
	b.WriteString(".btn-b{font:inherit;border:none;cursor:pointer;text-align:center;transition:background-color .15s ease,color .15s ease;}")
	b.WriteString(".dm-toggle{font:inherit;background:none;border:none;cursor:pointer;padding:0;line-height:1;display:flex;align-items:center;justify-content:center;width:100%;height:100%;user-select:none;-webkit-user-select:none;}")
	b.WriteString(":root[data-theme=\"dark\"] .dm-toggle{color:var(--waxp-text);}")
	b.WriteString(".dm-icon-sun{display:none;}")
	b.WriteString(":root[data-theme=\"dark\"] .dm-icon-moon{display:none;}")
	b.WriteString(":root[data-theme=\"dark\"] .dm-icon-sun{display:inline-block;}")
	b.WriteString(".icon-wrap{display:flex;align-items:center;justify-content:center;width:100%;height:100%;position:relative;z-index:1;line-height:1;}")
	b.WriteString(".lang-select{font:inherit;appearance:none;-webkit-appearance:none;width:100%;height:100%;border:none;background:transparent;padding:0 .1rem;outline:none;box-shadow:none;cursor:pointer;text-align:center;text-align-last:center;color:var(--waxp-text);}")
	b.WriteString(":root[data-theme=\"dark\"] .lang-select{color:var(--waxp-text);}")
	b.WriteString(".menu-nav{width:100%;height:100%;display:flex;align-items:center;position:relative;z-index:1;}")
	b.WriteString(".menu-list{display:flex;align-items:center;gap:1.5rem;list-style:none;margin:0;padding:0;width:100%;}")
	b.WriteString(".menu-item{position:relative;white-space:nowrap;}")
	b.WriteString(".menu-link{text-decoration:none;cursor:pointer;transition:color .15s ease;}")
	b.WriteString(".menu-link:hover{color:var(--m-hover)!important;}")
	b.WriteString(".menu-sub{display:none;position:absolute;top:100%;left:0;list-style:none;margin:0;padding:.35rem 0;border-radius:4px;box-shadow:0 4px 12px rgba(0,0,0,.12);z-index:99999;min-width:180px;}")
	b.WriteString(".menu-item:hover>.menu-sub{display:block;}")
	b.WriteString(".menu-sublink{display:block;padding:.35rem 1rem;text-decoration:none;cursor:pointer;transition:color .15s ease;white-space:nowrap;}")
	b.WriteString(".menu-sublink:hover{color:var(--m-hover)!important;}")
	b.WriteString(".space-divider{position:absolute;top:50%;left:0;right:0;transform:translateY(-50%);pointer-events:none;}")
	b.WriteString(".img-block{display:flex;align-items:center;justify-content:center;}")
	b.WriteString(".img-block img{display:block;}")
	b.WriteString(".img-dark{display:none;}")
	b.WriteString(":root[data-theme=\"dark\"] .img-light{display:none;}")
	b.WriteString(":root[data-theme=\"dark\"] .img-dark{display:block;}")

	b.WriteString(".lb-chk{display:none;}")
	b.WriteString(".lb-trigger{cursor:zoom-in;display:flex;width:100%;height:100%;}")
	b.WriteString(".lb-ov{display:none;position:fixed;inset:0;z-index:9999;align-items:center;justify-content:center;background:rgba(0,0,0,.85);cursor:zoom-out;}")
	b.WriteString(":root[data-theme=\"dark\"] .lb-ov{background:rgba(0,0,0,.92);}")
	b.WriteString(".lb-x{position:absolute;top:16px;right:16px;font-size:32px;color:#fff;cursor:pointer;line-height:1;user-select:none;-webkit-user-select:none;}")
}

func writeHeaderCSS(b *strings.Builder, headers json.RawMessage) {
	var h map[string]struct {
		Size       float64 `json:"size"`
		Family     string  `json:"family"`
		Weight     int     `json:"weight"`
		LineHeight float64 `json:"lineHeight"`
		Italic     *bool   `json:"italic"`
	}
	if err := json.Unmarshal(headers, &h); err != nil {
		return
	}
	for tag, cfg := range h {
		fs := "normal"
		if cfg.Italic != nil && *cfg.Italic {
			fs = "italic"
		}
		lh := cfg.LineHeight
		if lh == 0 {
			lh = cfg.Size
		}
		fmt.Fprintf(b, ".b-tiptap %s{font-size:%gem;line-height:%gem;font-family:'%s',sans-serif;font-weight:%d;font-style:%s;}",
			tag, cfg.Size, lh, cfg.Family, cfg.Weight, fs,
		)
	}
}

type sectionRender struct {
	section   *Section
	isFixed   bool
	cssPrefix string
}

func writeSectionCSS(b *strings.Builder, sr sectionRender, opts SiteOptions) {
	s := sr.section
	p := sr.cssPrefix
	desktopWidth := opts.DesktopWidth

	fmt.Fprintf(b, ".%s-row{", p)
	writeBackgroundCSS(b, s.Style.SectionBackground, "light")
	writeSidesCSS(b, "margin", s.Style.Margin)
	if sr.isFixed {
		b.WriteString("position:sticky;top:0;z-index:100;transition:background-color .2s,backdrop-filter .2s;")
	}
	b.WriteString("}")
	if sr.isFixed {
		fmt.Fprintf(b, ".%s-row.sr-scrolled{background-color:%scc;backdrop-filter:blur(12px);-webkit-backdrop-filter:blur(12px);}", p, opts.LightBackColor)
		fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s-row.sr-scrolled{background-color:%scc;}", p, opts.DarkBackColor)
	}
	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s-row{", p)
	writeBackgroundCSS(b, s.Style.SectionBackground, "dark")
	if sr.isFixed {
		b.WriteString("position:sticky;top:0;z-index:100;")
	}
	b.WriteString("}")

	maxW := desktopWidth
	if s.Style.MaxWidth != nil && *s.Style.MaxWidth > 0 {
		maxW = int(*s.Style.MaxWidth)
	}

	fmt.Fprintf(b, ".%s{position:relative;", p)
	fmt.Fprintf(b, "grid-template-columns:repeat(%d,1fr);grid-template-rows:repeat(%d,minmax(20px,1fr));gap:%dpx;",
		s.Desktop.Cols, s.Desktop.Rows, s.Desktop.Gap,
	)
	if s.Desktop.Rows > 0 {
		fmt.Fprintf(b, "aspect-ratio:%d/%d;", s.Desktop.Cols, s.Desktop.Rows)
	}
	if !s.Style.FullWidth {
		fmt.Fprintf(b, "max-width:%dpx;", maxW)
	}
	writeBackgroundCSS(b, s.Style.Background, "light")
	writeSidesCSS(b, "padding", s.Style.Padding)
	if bgNeedsClip(s.Style.Background) {
		b.WriteString("overflow:hidden;")
	}
	b.WriteString("}")

	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s{", p)
	writeBackgroundCSS(b, s.Style.Background, "dark")
	b.WriteString("}")

	writeFluidFontCSS(b, p, opts.FontSize, opts.LineHeight, maxW, s.Style.FullWidth, opts)

	for _, blk := range s.Blocks {
		writeBlockCSS(b, blk, p, opts, maxW, s.Style.FullWidth)
	}

	mobileBP := opts.MobileBP
	if mobileBP == 0 {
		mobileBP = 767
	}
	tabletBP := opts.TabletBP
	if tabletBP == 0 {
		tabletBP = 1024
	}

	fmt.Fprintf(b, "@media(max-width:%dpx){", tabletBP)
	fmt.Fprintf(b, ".%s{grid-template-columns:repeat(%d,1fr);gap:%dpx;}", p, s.Tablet.Cols, s.Tablet.Gap)
	if s.Tablet.Rows > 0 {
		fmt.Fprintf(b, ".%s{grid-template-rows:repeat(%d,minmax(20px,1fr));aspect-ratio:%d/%d;}", p, s.Tablet.Rows, s.Tablet.Cols, s.Tablet.Rows)
	}
	for _, blk := range s.Blocks {
		c := blk.T
		fmt.Fprintf(b, ".%s-b%d{grid-column:%d/span %d;grid-row:%d/span %d;}", p, blk.ID, c.X, c.W, c.Y, c.H)
	}
	writeFluidFontTabletCSS(b, p, opts.FontSize, opts.LineHeight, maxW, s.Style.FullWidth, opts)
	for _, blk := range s.Blocks {
		writeBlockFontCSS(b, blk, p, "t", opts, maxW, s.Style.FullWidth)
		writeMenuBlockFontCSS(b, blk, p, "t", opts, maxW, s.Style.FullWidth)
	}
	b.WriteString("}")

	fmt.Fprintf(b, "@media(max-width:%dpx){", mobileBP)
	fmt.Fprintf(b, ".%s{grid-template-columns:repeat(%d,1fr);gap:%dpx;}", p, s.Mobile.Cols, s.Mobile.Gap)
	if s.Mobile.Rows > 0 {
		fmt.Fprintf(b, ".%s{grid-template-rows:repeat(%d,minmax(20px,1fr));aspect-ratio:%d/%d;}", p, s.Mobile.Rows, s.Mobile.Cols, s.Mobile.Rows)
	}
	for _, blk := range s.Blocks {
		c := blk.M
		fmt.Fprintf(b, ".%s-b%d{grid-column:%d/span %d;grid-row:%d/span %d;}", p, blk.ID, c.X, c.W, c.Y, c.H)
	}
	writeFluidFontMobileCSS(b, p, opts.FontSize, opts.LineHeight, maxW, s.Style.FullWidth, opts)
	for _, blk := range s.Blocks {
		writeBlockFontCSS(b, blk, p, "m", opts, maxW, s.Style.FullWidth)
		writeMenuBlockFontCSS(b, blk, p, "m", opts, maxW, s.Style.FullWidth)
	}
	b.WriteString("}")

	for _, hideOn := range s.Style.HideOn {
		switch hideOn {
		case "mobile":
			fmt.Fprintf(b, "@media(max-width:%dpx){.%s-row{display:none!important;}}", mobileBP, p)
		case "tablet":
			fmt.Fprintf(b, "@media(min-width:%dpx)and(max-width:%dpx){.%s-row{display:none!important;}}", mobileBP+1, tabletBP, p)
		case "desktop":
			fmt.Fprintf(b, "@media(min-width:%dpx){.%s-row{display:none!important;}}", tabletBP+1, p)
		}
	}
}

func writeBlockCSS(b *strings.Builder, blk Block, prefix string, opts SiteOptions, targetWidth int, fullWidth bool) {
	c := blk.D
	fmt.Fprintf(b, ".%s-b%d{grid-column:%d/span %d;grid-row:%d/span %d;",
		prefix, blk.ID, c.X, c.W, c.Y, c.H,
	)

	writeBackgroundCSS(b, blk.Style.Background, "light")
	writeBorderCSS(b, blk.Style.Border)
	writeSidesCSS(b, "padding", blk.Style.Padding)
	writeSidesCSS(b, "margin", blk.Style.Margin)
	if bgNeedsClip(blk.Style.Background) {
		b.WriteString("overflow:hidden;")
	}
	if blk.Type == "Menu" {
		b.WriteString("overflow:visible;")
	}
	b.WriteString("}")

	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s-b%d{", prefix, blk.ID)
	writeBackgroundCSS(b, blk.Style.Background, "dark")
	b.WriteString("}")

	writeBlockFontCSS(b, blk, prefix, "d", opts, targetWidth, fullWidth)
	writeBlockColorCSS(b, blk, prefix)
	writeMenuBlockFontCSS(b, blk, prefix, "d", opts, targetWidth, fullWidth)

	for _, hideOn := range blk.Style.HideOn {
		mobileBP := opts.MobileBP
		tabletBP := opts.TabletBP
		if mobileBP == 0 {
			mobileBP = 767
		}
		if tabletBP == 0 {
			tabletBP = 1024
		}
		switch hideOn {
		case "mobile":
			fmt.Fprintf(b, "@media(max-width:%dpx){.%s-b%d{display:none!important;}}", mobileBP, prefix, blk.ID)
		case "tablet":
			fmt.Fprintf(b, "@media(min-width:%dpx)and(max-width:%dpx){.%s-b%d{display:none!important;}}", mobileBP+1, tabletBP, prefix, blk.ID)
		case "desktop":
			fmt.Fprintf(b, "@media(min-width:%dpx){.%s-b%d{display:none!important;}}", tabletBP+1, prefix, blk.ID)
		}
	}
}

func writeBlockFontCSS(b *strings.Builder, blk Block, prefix string, bp string, opts SiteOptions, targetWidth int, fullWidth bool) {
	if blk.FontSize == nil && blk.LineHeight == nil {
		return
	}
	fs := opts.FontSize
	lh := opts.LineHeight
	if blk.FontSize != nil {
		fs = *blk.FontSize
	}
	if blk.LineHeight != nil {
		lh = *blk.LineHeight
	}

	switch bp {
	case "d":
		writeFluidFontCSS(b, fmt.Sprintf("%s-b%d", prefix, blk.ID), fs, lh, targetWidth, fullWidth, opts)
	case "t":
		writeFluidFontTabletCSS(b, fmt.Sprintf("%s-b%d", prefix, blk.ID), fs, lh, targetWidth, fullWidth, opts)
	case "m":
		writeFluidFontMobileCSS(b, fmt.Sprintf("%s-b%d", prefix, blk.ID), fs, lh, targetWidth, fullWidth, opts)
	}
}

func writeBlockColorCSS(b *strings.Builder, blk Block, prefix string) {
	if blk.Color == nil && blk.DarkColor == nil {
		return
	}
	if blk.Color != nil && *blk.Color != "" {
		fmt.Fprintf(b, ".%s-b%d .b-inner{color:%s;}", prefix, blk.ID, *blk.Color)
	}
	if blk.DarkColor != nil && *blk.DarkColor != "" {
		fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s-b%d .b-inner{color:%s;}", prefix, blk.ID, *blk.DarkColor)
	}
}

func writeMenuBlockFontCSS(b *strings.Builder, blk Block, prefix string, bp string, opts SiteOptions, targetWidth int, fullWidth bool) {
	if blk.Type != "Menu" {
		return
	}
	blockSel := fmt.Sprintf("%s-b%d", prefix, blk.ID)

	if blk.MenuFontSize != nil || blk.MenuLineHeight != nil {
		fs := opts.FontSize
		lh := opts.LineHeight
		if blk.MenuFontSize != nil {
			fs = *blk.MenuFontSize
		}
		if blk.MenuLineHeight != nil {
			lh = *blk.MenuLineHeight
		}
		linkSel := blockSel + " .menu-link"
		switch bp {
		case "d":
			writeFluidFontCSS(b, linkSel, fs, lh, targetWidth, fullWidth, opts)
		case "t":
			writeFluidFontTabletCSS(b, linkSel, fs, lh, targetWidth, fullWidth, opts)
		case "m":
			writeFluidFontMobileCSS(b, linkSel, fs, lh, targetWidth, fullWidth, opts)
		}
	}

	if blk.MenuSubFontSize != nil || blk.MenuSubLineHeight != nil {
		fs := opts.FontSize
		lh := opts.LineHeight
		if blk.MenuSubFontSize != nil {
			fs = *blk.MenuSubFontSize
		}
		if blk.MenuSubLineHeight != nil {
			lh = *blk.MenuSubLineHeight
		}
		subSel := blockSel + " .menu-sublink"
		switch bp {
		case "d":
			writeFluidFontCSS(b, subSel, fs, lh, targetWidth, fullWidth, opts)
		case "t":
			writeFluidFontTabletCSS(b, subSel, fs, lh, targetWidth, fullWidth, opts)
		case "m":
			writeFluidFontMobileCSS(b, subSel, fs, lh, targetWidth, fullWidth, opts)
		}
	}
}

func writeFluidFontCSS(b *strings.Builder, selector string, fontSize, lineHeight float64, targetWidth int, fullWidth bool, opts SiteOptions) {
	if fullWidth {
		zoomFactor := 1.491 - 0.000965*float64(targetWidth)
		fsVW := fontSize + zoomFactor
		lhVW := fsVW * lineHeight
		fmt.Fprintf(b, ".%s{font-size:%gvw;line-height:%gvw;}", selector, fsVW, lhVW)
		return
	}

	fixedPx := fontSize * 16
	textZoom := opts.DesktopTextZoom
	if textZoom == 0 {
		textZoom = 1
	}

	a := fixedPx * (1 - textZoom)
	slope := fixedPx * textZoom * 100 / float64(targetWidth)

	fmt.Fprintf(b, "@media(min-width:%dpx){.%s{font-size:%gpx;line-height:%gpx;}}", targetWidth, selector, fixedPx, fixedPx*lineHeight)

	margin := opts.DesktopMargin
	if margin > 0 {
		fsCorr := slope * 2 * margin / 100
		lhCorr := fsCorr * lineHeight
		fmt.Fprintf(b, "@media(max-width:%dpx){.%s{font-size:calc(%gpx + %gvw - %gpx);line-height:calc(%gpx + %gvw - %gpx);}}",
			targetWidth-1, selector, a, slope, fsCorr, a*lineHeight, slope*lineHeight, lhCorr,
		)
	} else {
		fmt.Fprintf(b, "@media(max-width:%dpx){.%s{font-size:calc(%gpx + %gvw);line-height:calc(%gpx + %gvw);}}",
			targetWidth-1, selector, a, slope, a*lineHeight, slope*lineHeight,
		)
	}
}

func writeFluidFontTabletCSS(b *strings.Builder, selector string, fontSize, lineHeight float64, targetWidth int, fullWidth bool, opts SiteOptions) {
	zoom := opts.TabletTextZoom
	if zoom == 0 {
		zoom = 1
	}
	fsVW := fontSize + zoom
	lhVW := fsVW * lineHeight
	margin := opts.TabletMargin
	if margin <= 0 {
		fmt.Fprintf(b, ".%s{font-size:%gvw;line-height:%gvw;}", selector, fsVW, lhVW)
		return
	}
	fsCorr := fsVW * 2 * margin / 100
	lhCorr := lhVW * 2 * margin / 100
	fmt.Fprintf(b, ".%s{font-size:calc(%gvw - %gpx);line-height:calc(%gvw - %gpx);}", selector, fsVW, fsCorr, lhVW, lhCorr)
}

func writeFluidFontMobileCSS(b *strings.Builder, selector string, fontSize, lineHeight float64, targetWidth int, fullWidth bool, opts SiteOptions) {
	zoom := opts.MobileTextZoom
	if zoom == 0 {
		zoom = 2.6
	}
	fsVW := fontSize + zoom
	lhVW := fsVW * lineHeight
	margin := opts.MobileMargin
	if margin <= 0 {
		fmt.Fprintf(b, ".%s{font-size:%gvw;line-height:%gvw;}", selector, fsVW, lhVW)
		return
	}
	fsCorr := fsVW * 2 * margin / 100
	lhCorr := lhVW * 2 * margin / 100
	fmt.Fprintf(b, ".%s{font-size:calc(%gvw - %gpx);line-height:calc(%gvw - %gpx);}", selector, fsVW, fsCorr, lhVW, lhCorr)
}

func writeBackgroundCSS(b *strings.Builder, bg Background, theme string) {
	if bg.Mode == "none" {
		return
	}
	switch bg.Mode {
	case "color":
		c := bg.LightColor
		if theme == "dark" {
			c = bg.DarkColor
		}
		if c != "" {
			fmt.Fprintf(b, "background-color:%s;", c)
		}
	case "gradient":
		a := bg.LightGradA
		gradB := bg.LightGradB
		if theme == "dark" {
			a = bg.DarkGradA
			gradB = bg.DarkGradB
		}
		deg := bg.GradDeg
		if deg == "" {
			deg = "0"
		}
		if a != "" && gradB != "" {
			fmt.Fprintf(b, "background:linear-gradient(%sdeg,%s,%s);", deg, a, gradB)
		} else if a != "" {
			fmt.Fprintf(b, "background-color:%s;", a)
		}
	}
}

func writeOverlayCSS(bg Background, theme string, mediaBase string) string {
	if bg.Mode != "image" {
		return ""
	}
	url := bg.URLDesk
	if theme == "dark" && bg.URLDeskDark != "" {
		url = bg.URLDeskDark
	}
	if url == "" {
		return ""
	}
	var rules []string
	rules = append(rules, fmt.Sprintf("background-image:url('%s%s');", mediaBase, url))

	hasOpacity := bg.Opacity != "" && bg.Opacity != "1"
	if hasOpacity {
		rules = append(rules, fmt.Sprintf("opacity:%s;", bg.Opacity))
	}

	zoomVal, _ := parseFloat(bg.Zoom, 100)
	if bg.Pos == "cover" && zoomVal > 100 {
		fx := bg.FocalX
		fy := bg.FocalY
		if fx == "" {
			fx = "50"
		}
		if fy == "" {
			fy = "50"
		}
		rules = append(rules, fmt.Sprintf("transform:scale(%g);", zoomVal/100))
		rules = append(rules, fmt.Sprintf("transform-origin:%s%% %s%%;", fx, fy))
	}

	if bg.FixImgBack {
		rules = append(rules, "background-attachment:fixed;")
	}

	posMap := map[string]string{
		"img":     "center center",
		"cover":   bg.FocalX + "% " + bg.FocalY + "%",
		"contain": "center center",
		"top":     "center top",
		"bottom":  "center bottom",
		"left":    "left center",
		"right":   "right center",
	}
	pos := posMap[bg.Pos]
	if pos == "" || (bg.Pos == "cover" && (bg.FocalX == "" || bg.FocalY == "")) {
		pos = "center center"
	}
	rules = append(rules, fmt.Sprintf("background-position:%s;", pos))

	switch bg.Pos {
	case "cover":
		rules = append(rules, "background-size:cover;")
	case "contain":
		rules = append(rules, "background-size:contain;")
	default:
		rules = append(rules, "background-size:auto;")
	}

	if bg.Repeat {
		rules = append(rules, "background-repeat:repeat;")
	} else {
		rules = append(rules, "background-repeat:no-repeat;")
	}

	return strings.Join(rules, "")
}

func bgNeedsClip(bg Background) bool {
	if bg.Mode != "image" {
		return false
	}
	zoomVal, _ := parseFloat(bg.Zoom, 100)
	return bg.Pos == "cover" && zoomVal > 100
}

func writeBorderCSS(b *strings.Builder, border Border) {
	r := border.Radius
	if r.TL != "" || r.TR != "" || r.BR != "" || r.BL != "" {
		fmt.Fprintf(b, "border-radius:%spx %spx %spx %spx;",
			cssVal(r.TL), cssVal(r.TR), cssVal(r.BR), cssVal(r.BL),
		)
	}
	ab := border.AllBorders
	if ab.Active {
		fmt.Fprintf(b, "border:%spx %s %s;", ab.Thick, ab.Mode, ab.Color)
	}
}

func writeSidesCSS(b *strings.Builder, prop string, s Sides) {
	if s.T == "" && s.R == "" && s.B == "" && s.L == "" {
		return
	}
	fmt.Fprintf(b, "%s:%s %s %s %s;", prop, s.T, s.R, s.B, s.L)
}

func parseFloat(s string, def float64) (float64, bool) {
	if s == "" {
		return def, false
	}
	v, err := fmt.Sscanf(s, "%f", &def)
	_ = v
	if err != nil {
		return def, false
	}
	f, e := parseFloatInline(s)
	if e != nil {
		return def, false
	}
	return f, true
}

func parseFloatInline(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func buildGoogleFontsURL(fonts []SiteFont) string {
	if len(fonts) == 0 {
		return ""
	}
	var families []string
	for _, f := range fonts {
		var weights []string
		for _, w := range f.Weights {
			weights = append(weights, fmt.Sprintf("%d", w))
		}
		family := strings.ReplaceAll(f.Family, " ", "+")
		if len(weights) > 0 {
			families = append(families, fmt.Sprintf("family=%s:wght@%s", family, strings.Join(weights, ";")))
		} else {
			families = append(families, fmt.Sprintf("family=%s", family))
		}
	}
	return "https://fonts.googleapis.com/css2?" + strings.Join(families, "&") + "&display=swap"
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

func cssVal(s string) string {
	if s == "" || s == "0" {
		return "0"
	}
	if strings.ContainsAny(s, "em%pxvhvw") {
		return s
	}
	return s
}
