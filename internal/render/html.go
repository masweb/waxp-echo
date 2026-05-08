package render

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

func buildHTML(sections []sectionRender, opts SiteOptions, data PageData) string {
	var b strings.Builder
	var lb strings.Builder

	b.WriteString("<!DOCTYPE html>")
	fmt.Fprintf(&b, "<html lang=\"%s\" data-theme=\"light\">", data.Locale)
	b.WriteString("<head>")
	b.WriteString("<meta charset=\"utf-8\">")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">")

	if data.SEO != nil {
		fmt.Fprintf(&b, "<title>%s</title>", html.EscapeString(data.SEO.Title))
		if data.SEO.Description != "" {
			fmt.Fprintf(&b, "<meta name=\"description\" content=\"%s\">", html.EscapeString(data.SEO.Description))
		}
		fmt.Fprintf(&b, "<meta property=\"og:title\" content=\"%s\">", html.EscapeString(data.SEO.Title))
		if data.SEO.Description != "" {
			fmt.Fprintf(&b, "<meta property=\"og:description\" content=\"%s\">", html.EscapeString(data.SEO.Description))
		}
		b.WriteString("<meta property=\"og:type\" content=\"website\">")
	}

	if data.Domain != "" {
		for _, sl := range data.PageSlugs {
			path := buildRoutePath(sl.LocaleCode, sl.IsDefault, sl.Slug)
			href := "https://" + data.Domain + path
			fmt.Fprintf(&b, "<link rel=\"alternate\" hreflang=\"%s\" href=\"%s\">", sl.LocaleCode, href)
		}
		defaultSlug := ""
		for _, sl := range data.PageSlugs {
			if sl.IsDefault {
				defaultSlug = sl.Slug
				break
			}
		}
		if defaultSlug != "" || data.Locale != "" {
			for _, sl := range data.PageSlugs {
				if sl.IsDefault {
					path := buildRoutePath(sl.LocaleCode, true, sl.Slug)
					fmt.Fprintf(&b, "<link rel=\"canonical\" href=\"https://%s%s\">", data.Domain, path)
					break
				}
			}
		}
	}

	fontURL := buildGoogleFontsURL(opts.Fonts)
	if fontURL != "" {
		b.WriteString("<link rel=\"preconnect\" href=\"https://fonts.googleapis.com\">")
		b.WriteString("<link rel=\"preconnect\" href=\"https://fonts.gstatic.com\" crossorigin>")
		fmt.Fprintf(&b, "<link href=\"%s\" rel=\"stylesheet\">", fontURL)
	}

	b.WriteString("<style>")
	b.WriteString(buildCSS(sections, opts))
	b.WriteString("</style>")

	b.WriteString(`<script>(function(){var s=localStorage.getItem('waxp-theme');if(s==='dark'||(!s&&window.matchMedia('(prefers-color-scheme:dark)').matches))document.documentElement.setAttribute('data-theme','dark');})();</script>`)

	b.WriteString("</head>")
	b.WriteString("<body>")
	b.WriteString("<div class=\"waxp\">")

	for _, sr := range sections {
		writeSectionHTML(&b, sr, opts, data, &lb)
	}

	if lb.Len() > 0 {
		b.WriteString(lb.String())
	}

	b.WriteString("</div>")

	hasDarkMode := false
	for _, sr := range sections {
		for _, blk := range sr.section.Blocks {
			if blk.Type == "DarkMode" {
				hasDarkMode = true
				break
			}
		}
		if hasDarkMode {
			break
		}
	}
	if hasDarkMode {
		b.WriteString(`<script>function waxpToggleTheme(){var d=document.documentElement;var c=d.getAttribute('data-theme')==='dark'?'light':'dark';d.setAttribute('data-theme',c);localStorage.setItem('waxp-theme',c);}</script>`)
	}

	hasSticky := false
	for _, sr := range sections {
		if sr.isFixed {
			hasSticky = true
			break
		}
	}
	if hasSticky {
		b.WriteString(`<script>(function(){var s=document.querySelectorAll('.sr');window.addEventListener('scroll',function(){s.forEach(function(e){e.classList.toggle('sr-scrolled',window.scrollY>0)})})})();</script>`)
	}

	if lb.Len() > 0 {
		b.WriteString(`<script>document.addEventListener('keydown',function(e){if(e.key==='Escape')document.querySelectorAll('.lb-chk:checked').forEach(function(c){c.checked=false})})</script>`)
	}

	b.WriteString("</body></html>")
	return b.String()
}

func writeSectionHTML(b *strings.Builder, sr sectionRender, opts SiteOptions, data PageData, lb *strings.Builder) {
	s := sr.section
	p := sr.cssPrefix

	fmt.Fprintf(b, "<div class=\"%s-row sr\">", p)

	overlay := writeOverlayCSS(s.Style.SectionBackground, "light", data.MediaBase)
	if overlay != "" {
		fmt.Fprintf(b, "<div class=\"s-overlay\" style=\"%s\"></div>", overlay)
		darkOverlay := writeOverlayCSS(s.Style.SectionBackground, "dark", data.MediaBase)
		if darkOverlay != "" && darkOverlay != overlay {
			fmt.Fprintf(b, "<style>:root[data-theme=\"dark\"] .%s-row .s-overlay{%s}</style>", p, darkOverlay)
		}
	}

	fmt.Fprintf(b, "<div class=\"%s s\">", p)

	sOverlay := writeOverlayCSS(s.Style.Background, "light", data.MediaBase)
	if sOverlay != "" {
		fmt.Fprintf(b, "<div class=\"s-overlay\" style=\"%s\"></div>", sOverlay)
		sDarkOverlay := writeOverlayCSS(s.Style.Background, "dark", data.MediaBase)
		if sDarkOverlay != "" && sDarkOverlay != sOverlay {
			fmt.Fprintf(b, "<style>:root[data-theme=\"dark\"] .%s .s-overlay{%s}</style>", p, sDarkOverlay)
		}
	}

	for i := range s.Blocks {
		writeBlockHTML(b, &s.Blocks[i], p, opts, data, lb)
	}

	b.WriteString("</div>")
	b.WriteString("</div>")
}

func writeBlockHTML(b *strings.Builder, blk *Block, prefix string, opts SiteOptions, data PageData, lb *strings.Builder) {
	fmt.Fprintf(b, "<div class=\"%s-b%d b\">", prefix, blk.ID)

	bgOverlay := writeOverlayCSS(blk.Style.Background, "light", data.MediaBase)
	if bgOverlay != "" {
		fmt.Fprintf(b, "<div class=\"b-overlay\" style=\"%s\"></div>", bgOverlay)
		darkBgOverlay := writeOverlayCSS(blk.Style.Background, "dark", data.MediaBase)
		if darkBgOverlay != "" && darkBgOverlay != bgOverlay {
			fmt.Fprintf(b, "<style>:root[data-theme=\"dark\"] .%s-b%d .b-overlay{%s}</style>", prefix, blk.ID, darkBgOverlay)
		}
	}

	center := ""
	switch blk.Type {
	case "DarkMode", "LanguageSwitcher", "Button", "Icon":
		center = " b-center"
	}

	fmt.Fprintf(b, "<div class=\"b-inner%s\">", center)

	switch blk.Type {
	case "Text":
		writeTextBlock(b, blk)
	case "Image":
		writeImageBlock(b, blk, data, lb)
	case "Button":
		writeButtonBlock(b, blk, prefix, opts)
	case "Space":
		writeSpaceBlock(b, blk)
	case "DarkMode":
		writeDarkModeBlock(b, blk, prefix, opts)
	case "LanguageSwitcher":
		writeLangBlock(b, blk, data)
	case "Icon":
		writeIconBlock(b, blk, prefix, opts)
	case "Menu":
		writeMenuBlock(b, blk, prefix, blk.ID, opts)
	}

	b.WriteString("</div>")
	b.WriteString("</div>")
}

var (
	reAnchorInternal = regexp.MustCompile(`<a\b[^>]*href="(/[^"]*)"[^>]*>`)
	reTargetBlank    = regexp.MustCompile(` target="_blank"`)
	reRelAttr        = regexp.MustCompile(` rel="[^"]*"`)
)

func cleanInternalLinks(s string) string {
	return reAnchorInternal.ReplaceAllStringFunc(s, func(tag string) string {
		tag = reTargetBlank.ReplaceAllString(tag, "")
		tag = reRelAttr.ReplaceAllString(tag, "")
		return tag
	})
}

func writeTextBlock(b *strings.Builder, blk *Block) {
	text := getLocale(blk.Locales, "text")
	text = cleanInternalLinks(text)
	fmt.Fprintf(b, "<div class=\"b-tiptap\">%s</div>", text)
}

func writeImageBlock(b *strings.Builder, blk *Block, data PageData, lb *strings.Builder) {
	if blk.Image == nil {
		return
	}
	img := blk.Image
	urlLight := img.URLDesk
	if urlLight == "" {
		return
	}

	alt := getLocale(blk.Locales, "alt")

	var style string
	switch img.Fit {
	case "cover":
		style = "width:100%;height:100%;object-fit:cover;"
	case "height":
		style = "height:100%;width:auto;max-width:none;"
	default:
		style = "width:100%;height:auto;"
	}

	urlDark := img.URLDeskDark

	if img.Lightbox {
		fmt.Fprintf(b, "<label for=\"lb-%d\" class=\"lb-trigger\">", blk.ID)
		if urlDark != "" {
			fmt.Fprintf(b, "<img src=\"%s%s\" alt=\"%s\" style=\"%s\" loading=\"lazy\" class=\"img-light\">",
				data.MediaBase, urlLight, html.EscapeString(alt), style,
			)
			fmt.Fprintf(b, "<img src=\"%s%s\" alt=\"%s\" style=\"%s\" loading=\"lazy\" class=\"img-dark\">",
				data.MediaBase, urlDark, html.EscapeString(alt), style,
			)
		} else {
			fmt.Fprintf(b, "<img src=\"%s%s\" alt=\"%s\" style=\"%s\" loading=\"lazy\">",
				data.MediaBase, urlLight, html.EscapeString(alt), style,
			)
		}
		b.WriteString("</label>")

		fmt.Fprintf(lb, "<style>")
		fmt.Fprintf(lb, "#lb-%d:checked~#lb-ov-%d{display:flex;}", blk.ID, blk.ID)
		if urlDark != "" {
			fmt.Fprintf(lb, "#lb-ov-%d .lb-d{display:none;}", blk.ID)
			fmt.Fprintf(lb, ":root[data-theme=\"dark\"] #lb-ov-%d .lb-l{display:none;}", blk.ID)
			fmt.Fprintf(lb, ":root[data-theme=\"dark\"] #lb-ov-%d .lb-d{display:block;}", blk.ID)
		}
		fmt.Fprintf(lb, "</style>")
		fmt.Fprintf(lb, "<input type=\"checkbox\" id=\"lb-%d\" class=\"lb-chk\" hidden>", blk.ID)
		fmt.Fprintf(lb, "<label for=\"lb-%d\" class=\"lb-ov\" id=\"lb-ov-%d\">", blk.ID, blk.ID)
		if urlDark != "" {
			fmt.Fprintf(lb, "<img src=\"%s%s\" alt=\"%s\" style=\"max-width:90vw;max-height:90vh;object-fit:contain;border-radius:4px\" class=\"lb-l\">",
				data.MediaBase, urlLight, html.EscapeString(alt),
			)
			fmt.Fprintf(lb, "<img src=\"%s%s\" alt=\"%s\" style=\"max-width:90vw;max-height:90vh;object-fit:contain;border-radius:4px\" class=\"lb-d\">",
				data.MediaBase, urlDark, html.EscapeString(alt),
			)
		} else {
			fmt.Fprintf(lb, "<img src=\"%s%s\" alt=\"%s\" style=\"max-width:90vw;max-height:90vh;object-fit:contain;border-radius:4px\">",
				data.MediaBase, urlLight, html.EscapeString(alt),
			)
		}
		lb.WriteString("<span class=\"lb-x\">&times;</span>")
		lb.WriteString("</label>")
		return
	}

	href := ""
	target := ""
	rel := ""
	if blk.Link != nil && blk.Link.URL != "" {
		href = blk.Link.URL
		if blk.Link.Type == "external" {
			target = " target=\"_blank\""
			rel = " rel=\"noopener noreferrer\""
		}
	}

	if href != "" {
		fmt.Fprintf(b, "<a class=\"b-link\" href=\"%s\"%s%s>", html.EscapeString(href), target, rel)
	}

	if urlDark != "" {
		fmt.Fprintf(b, "<img src=\"%s%s\" alt=\"%s\" style=\"%s\" loading=\"lazy\" class=\"img-light\">",
			data.MediaBase, urlLight, html.EscapeString(alt), style,
		)
		fmt.Fprintf(b, "<img src=\"%s%s\" alt=\"%s\" style=\"%s\" loading=\"lazy\" class=\"img-dark\">",
			data.MediaBase, urlDark, html.EscapeString(alt), style,
		)
	} else {
		fmt.Fprintf(b, "<img src=\"%s%s\" alt=\"%s\" style=\"%s\" loading=\"lazy\">",
			data.MediaBase, urlLight, html.EscapeString(alt), style,
		)
	}

	if href != "" {
		b.WriteString("</a>")
	}
}

func writeButtonBlock(b *strings.Builder, blk *Block, prefix string, opts SiteOptions) {
	if blk.Button == nil {
		return
	}
	btn := blk.Button
	label := getLocale(blk.Locales, "label")
	selector := fmt.Sprintf("%s-b%d", prefix, blk.ID)

	tag := "button"
	href := ""
	target := ""
	rel := ""
	if blk.Link != nil && blk.Link.URL != "" {
		if blk.Link.Type == "external" {
			tag = "a"
			href = fmt.Sprintf(" href=\"%s\"", html.EscapeString(blk.Link.URL))
			target = " target=\"_blank\""
			rel = " rel=\"noopener noreferrer\""
		} else if blk.Link.Type == "internal" {
			tag = "a"
			href = fmt.Sprintf(" href=\"%s\"", html.EscapeString(blk.Link.URL))
		}
	}

	var inlineStyle []string
	if btn.Width != "" && btn.Width != "0" {
		inlineStyle = append(inlineStyle, fmt.Sprintf("width:%s%%", btn.Width))
	}
	if btn.Padding.T != "" || btn.Padding.R != "" || btn.Padding.B != "" || btn.Padding.L != "" {
		inlineStyle = append(inlineStyle, fmt.Sprintf("padding:%spx %spx %spx %spx",
			szVal(btn.Padding.T, "0"), szVal(btn.Padding.R, "0"), szVal(btn.Padding.B, "0"), szVal(btn.Padding.L, "0"),
		))
	}
	r := btn.Border.Radius
	if r.TL != "" || r.TR != "" || r.BR != "" || r.BL != "" {
		inlineStyle = append(inlineStyle, fmt.Sprintf("border-radius:%spx %spx %spx %spx",
			szVal(r.TL, "0"), szVal(r.TR, "0"), szVal(r.BR, "0"), szVal(r.BL, "0"),
		))
	}

	inlineStyleStr := strings.Join(inlineStyle, ";")

	var lightRules []string
	if btn.Bg.Light != "" {
		lightRules = append(lightRules, fmt.Sprintf("background-color:%s", btn.Bg.Light))
	}
	textColor := btn.TextColor.Light
	if textColor == "" && blk.Color != nil {
		textColor = *blk.Color
	}
	if textColor != "" {
		lightRules = append(lightRules, fmt.Sprintf("color:%s", textColor))
	}
	if btn.Border.AllBorders.Active {
		lightRules = append(lightRules, fmt.Sprintf("border:%spx %s %s",
			btn.Border.AllBorders.Thick, btn.Border.AllBorders.Mode, btn.BorderColor.Light,
		))
	}
	lightRules = append(lightRules,
		fmt.Sprintf("--btn-hover-bg:%s", btn.Hover.Light),
		fmt.Sprintf("--btn-active-bg:%s", btn.Active.Light),
		fmt.Sprintf("--btn-hover-color:%s", btn.HoverTextColor.Light),
		fmt.Sprintf("--btn-active-color:%s", btn.ActiveTextColor.Light),
		fmt.Sprintf("--btn-focus-ring:0 0 0 3px %s", btn.Focus.Light),
	)

	var darkRules []string
	if btn.Bg.Dark != "" {
		darkRules = append(darkRules, fmt.Sprintf("background-color:%s", btn.Bg.Dark))
	}
	darkTextColor := btn.TextColor.Dark
	if darkTextColor == "" && blk.DarkColor != nil {
		darkTextColor = *blk.DarkColor
	}
	if darkTextColor != "" {
		darkRules = append(darkRules, fmt.Sprintf("color:%s", darkTextColor))
	}
	if btn.Border.AllBorders.Active {
		darkRules = append(darkRules, fmt.Sprintf("border:%spx %s %s",
			btn.Border.AllBorders.Thick, btn.Border.AllBorders.Mode, btn.BorderColor.Dark,
		))
	}
	darkRules = append(darkRules,
		fmt.Sprintf("--btn-hover-bg:%s", btn.Hover.Dark),
		fmt.Sprintf("--btn-active-bg:%s", btn.Active.Dark),
		fmt.Sprintf("--btn-hover-color:%s", btn.HoverTextColor.Dark),
		fmt.Sprintf("--btn-active-color:%s", btn.ActiveTextColor.Dark),
		fmt.Sprintf("--btn-focus-ring:0 0 0 3px %s", btn.Focus.Dark),
	)

	fmt.Fprintf(b, "<div class=\"btn-wrap\">")

	fmt.Fprintf(b, "<%s class=\"btn-b\"%s%s%s", tag, href, target, rel)
	if inlineStyleStr != "" {
		fmt.Fprintf(b, " style=\"%s\"", inlineStyleStr)
	}
	fmt.Fprintf(b, ">%s</%s>", html.EscapeString(label), tag)

	fmt.Fprintf(b, "<style>")
	fmt.Fprintf(b, ".%s .btn-b{%s}", selector, strings.Join(lightRules, ";"))
	fmt.Fprintf(b, ".%s .btn-b:hover{background-color:var(--btn-hover-bg)!important;color:var(--btn-hover-color)!important;}", selector)
	fmt.Fprintf(b, ".%s .btn-b:active{background-color:var(--btn-active-bg)!important;color:var(--btn-active-color)!important;}", selector)
	fmt.Fprintf(b, ".%s .btn-b:focus{box-shadow:var(--btn-focus-ring);outline:none;}", selector)

	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s .btn-b{%s}", selector, strings.Join(darkRules, ";"))
	fmt.Fprintf(b, "</style>")

	b.WriteString("</div>")
}

func writeSpaceBlock(b *strings.Builder, blk *Block) {
	if blk.Divider != nil && blk.Divider.Active {
		fmt.Fprintf(b, "<div class=\"space-divider\" style=\"border-top:%spx %s %s\"></div>",
			blk.Divider.Thick, blk.Divider.Mode, blk.Divider.Color,
		)
	}
}

func writeDarkModeBlock(b *strings.Builder, blk *Block, prefix string, opts SiteOptions) {
	selector := fmt.Sprintf("%s-b%d", prefix, blk.ID)

	var lightColor string
	if blk.Color != nil && *blk.Color != "" {
		lightColor = *blk.Color
	} else {
		lightColor = opts.LightColor
	}
	var darkColor string
	if blk.DarkColor != nil && *blk.DarkColor != "" {
		darkColor = *blk.DarkColor
	} else {
		darkColor = opts.DarkColor
	}

	fmt.Fprintf(b, "<style>")
	fmt.Fprintf(b, ".%s .dm-toggle{color:%s;}", selector, lightColor)
	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s .dm-toggle{color:%s;}", selector, darkColor)
	b.WriteString("</style>")

	fmt.Fprintf(b, "<button class=\"dm-toggle\" onclick=\"waxpToggleTheme()\">")
	b.WriteString(`<svg class="dm-icon-moon" xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor"><path d="M12 1.992a10 10 0 1 0 9.236 13.838c.341 -.82 -.476 -1.644 -1.298 -1.31a6.5 6.5 0 0 1 -6.864 -10.787l.077 -.08c.551 -.63 .113 -1.653 -.758 -1.653h-.266l-.068 -.006l-.06 -.002z"/></svg>`)
	b.WriteString(`<svg class="dm-icon-sun" xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor"><path d="M12 19a1 1 0 0 1 .993 .883l.007 .117v1a1 1 0 0 1 -1.993 .117l-.007 -.117v-1a1 1 0 0 1 1 -1z"/><path d="M18.313 16.91l.094 .083l.7 .7a1 1 0 0 1 -1.32 1.497l-.094 -.083l-.7 -.7a1 1 0 0 1 1.218 -1.567l.102 .07z"/><path d="M7.007 16.993a1 1 0 0 1 .083 1.32l-.083 .094l-.7 .7a1 1 0 0 1 -1.497 -1.32l.083 -.094l-.7 -.7a1 1 0 0 1 1.414 0z"/><path d="M4 11a1 1 0 0 1 .117 1.993l-.117 .007h-1a1 1 0 0 1 -.117 -1.993l.117 -.007h1z"/><path d="M21 11a1 1 0 0 1 .117 1.993l-.117 .007h-1a1 1 0 0 1 -.117 -1.993l.117 -.007h1z"/><path d="M6.213 4.81l.094 .083l.7 .7a1 1 0 0 1 -1.32 1.497l-.094 -.083l-.7 -.7a1 1 0 0 1 1.217 -1.567l.102 .07z"/><path d="M19.107 4.893a1 1 0 0 1 .083 1.32l-.083 .094l-.7 .7a1 1 0 0 1 -1.497 -1.32l.083 -.094l.7 -.7a1 1 0 0 1 1.414 0z"/><path d="M12 2a1 1 0 0 1 .993 .883l.007 .117v1a1 1 0 0 1 -1.993 .117l-.007 -.117v-1a1 1 0 0 1 1 -1z"/><path d="M12 7a5 5 0 1 1 -4.995 5.217l-.005 -.217l.005 -.217a5 5 0 0 1 4.995 -4.783z"/></svg>`)
	b.WriteString("</button>")
}

func writeIconBlock(b *strings.Builder, blk *Block, prefix string, opts SiteOptions) {
	if blk.Icon == nil || blk.Icon.Name == "" {
		return
	}

	strokeWidth := blk.Icon.StrokeWidth
	if strokeWidth <= 0 {
		strokeWidth = 1
	}

	svgContent := GetIconSVG(blk.Icon.Name, strokeWidth)
	if svgContent == "" {
		return
	}

	selector := fmt.Sprintf("%s-b%d", prefix, blk.ID)

	var lightColor string
	if blk.Color != nil && *blk.Color != "" {
		lightColor = *blk.Color
	} else {
		lightColor = opts.LightColor
	}

	var darkColor string
	if blk.DarkColor != nil && *blk.DarkColor != "" {
		darkColor = *blk.DarkColor
	} else {
		darkColor = opts.DarkColor
	}

	href := ""
	target := ""
	rel := ""
	if blk.Link != nil && blk.Link.URL != "" {
		if blk.Link.Type == "external" {
			href = fmt.Sprintf(" href=\"%s\"", html.EscapeString(blk.Link.URL))
			target = " target=\"_blank\""
			rel = " rel=\"noopener noreferrer\""
		} else if blk.Link.Type == "internal" {
			href = fmt.Sprintf(" href=\"%s\"", html.EscapeString(blk.Link.URL))
		}
	}

	fmt.Fprintf(b, "<style>")
	fmt.Fprintf(b, ".%s .icon-wrap{color:%s;}", selector, lightColor)
	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s .icon-wrap{color:%s;}", selector, darkColor)
	b.WriteString("</style>")

	if href != "" {
		fmt.Fprintf(b, "<a class=\"b-link icon-wrap\"%s%s%s>", href, target, rel)
		b.WriteString(svgContent)
		b.WriteString("</a>")
	} else {
		b.WriteString("<div class=\"icon-wrap\">")
		b.WriteString(svgContent)
		b.WriteString("</div>")
	}
}

func writeLangBlock(b *strings.Builder, blk *Block, data PageData) {
	fmt.Fprintf(b, "<select class=\"lang-select\" onchange=\"location.href=this.value\">")
	for _, loc := range data.Locales {
		url := "/" + loc.Code
		for _, sl := range data.PageSlugs {
			if sl.LocaleCode == loc.Code {
				url = buildRoutePath(sl.LocaleCode, sl.IsDefault, sl.Slug)
				break
			}
		}
		selected := ""
		if loc.Code == data.Locale {
			selected = " selected"
		}
		fmt.Fprintf(b, "<option value=\"%s\"%s>%s</option>", url, selected, strings.ToUpper(loc.Code))
	}
	b.WriteString("</select>")
}

func writeMenuBlock(b *strings.Builder, blk *Block, prefix string, blkID int64, opts SiteOptions) {
	if len(blk.Menu) == 0 {
		return
	}

	menuColors := blk.MenuColors
	if menuColors == nil || menuColors.Color.Light == "" {
		menuColors = &MenuColors{
			Color:  ButtonColors{Light: "#212529", Dark: "#f8f9fa"},
			Hover:  ButtonColors{Light: "#0d6efd", Dark: "#6ea8fe"},
			Active: ButtonColors{Light: "#0a58ca", Dark: "#5aa4f0"},
		}
	}

	selector := fmt.Sprintf("%s-b%d", prefix, blkID)

	fmt.Fprintf(b, "<style>")
	fmt.Fprintf(b, ".%s{--m-color:%s;--m-hover:%s;--m-active:%s;--m-sub-bg:#fff;}",
		selector, menuColors.Color.Light, menuColors.Hover.Light, menuColors.Active.Light,
	)
	fmt.Fprintf(b, ".%s .menu-link{color:var(--m-color);}",
		selector,
	)
	fmt.Fprintf(b, ".%s .menu-sublink{color:var(--m-color);}",
		selector,
	)
	fmt.Fprintf(b, ".%s .menu-sub{background:var(--m-sub-bg);}",
		selector,
	)
	fmt.Fprintf(b, ":root[data-theme=\"dark\"] .%s{--m-color:%s;--m-hover:%s;--m-active:%s;--m-sub-bg:#2b2b2b;}",
		selector, menuColors.Color.Dark, menuColors.Hover.Dark, menuColors.Active.Dark,
	)
	fmt.Fprintf(b, "</style>")

	l1Style := fontStyles(blk.MenuFont)

	fmt.Fprintf(b, "<nav class=\"menu-nav\">")
	b.WriteString("<ul class=\"menu-list\">")

	for _, item := range blk.Menu {
		b.WriteString("<li class=\"menu-item\">")

		if item.Link != nil && item.Link.URL != "" {
			if item.Link.Type == "external" {
				fmt.Fprintf(b, "<a class=\"menu-link\" href=\"%s\" target=\"_blank\" rel=\"noopener noreferrer\" style=\"%s\">%s</a>",
					html.EscapeString(item.Link.URL), l1Style, html.EscapeString(item.Label),
				)
			} else {
				fmt.Fprintf(b, "<a class=\"menu-link\" href=\"%s\" style=\"%s\">%s</a>",
					html.EscapeString(item.Link.URL), l1Style, html.EscapeString(item.Label),
				)
			}
		} else {
			fmt.Fprintf(b, "<span class=\"menu-link\" style=\"%s\">%s</span>", l1Style, html.EscapeString(item.Label))
		}

		if len(item.Children) > 0 {
			subStyle := fontStyles(blk.MenuSubFont)
			b.WriteString("<ul class=\"menu-sub\">")
			for _, child := range item.Children {
				if child.Link != nil && child.Link.URL != "" {
					if child.Link.Type == "external" {
						fmt.Fprintf(b, "<li><a class=\"menu-sublink\" href=\"%s\" target=\"_blank\" rel=\"noopener noreferrer\" style=\"%s\">%s</a></li>",
							html.EscapeString(child.Link.URL), subStyle, html.EscapeString(child.Label),
						)
					} else {
						fmt.Fprintf(b, "<li><a class=\"menu-sublink\" href=\"%s\" style=\"%s\">%s</a></li>",
							html.EscapeString(child.Link.URL), subStyle, html.EscapeString(child.Label),
						)
					}
				} else {
					fmt.Fprintf(b, "<li><span class=\"menu-sublink\" style=\"%s\">%s</span></li>", subStyle, html.EscapeString(child.Label))
				}
			}
			b.WriteString("</ul>")
		}

		b.WriteString("</li>")
	}

	b.WriteString("</ul>")
	b.WriteString("</nav>")
}

func fontStyles(f *Font) string {
	if f == nil || f.Family == "" {
		return ""
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("font-family:'%s'", f.Family))
	parts = append(parts, fmt.Sprintf("font-weight:%d", f.Weight))
	if f.Italic != nil && *f.Italic {
		parts = append(parts, "font-style:italic")
	}
	return strings.Join(parts, ";")
}

func szVal(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func getLocale(locales map[string]interface{}, key string) string {
	if locales == nil {
		return ""
	}
	v, ok := locales[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func buildRoutePath(localeCode string, isDefault bool, slug string) string {
	if slug == "" {
		if isDefault {
			return "/"
		}
		return "/" + localeCode
	}
	if isDefault {
		return "/" + slug
	}
	return "/" + localeCode + "/" + slug
}
