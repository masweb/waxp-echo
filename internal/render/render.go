package render

import (
	"encoding/json"
	"fmt"
)

func RenderPage(data PageData) (string, error) {
	var sections []sectionRender

	if data.Options.Header != nil {
		sections = append(sections, sectionRender{
			section:   data.Options.Header,
			isFixed:   true,
			cssPrefix: "sh",
		})
	}

	for i := range data.Layout {
		sections = append(sections, sectionRender{
			section:   &data.Layout[i],
			cssPrefix: fmt.Sprintf("s%d", data.Layout[i].ID),
		})
	}

	if data.Options.Footer != nil {
		sections = append(sections, sectionRender{
			section:   data.Options.Footer,
			isFixed:   true,
			cssPrefix: "sf",
		})
	}

	return buildHTML(sections, data.Options, data), nil
}

type RenderInput struct {
	LayoutJSON   json.RawMessage
	OptionsJSON  json.RawMessage
	SEO          *SEOData
	Locale       string
	Locales      []LocaleInfo
	PageSlugs    []SlugInfo
	Domain       string
	MediaBase    string
}

func Render(input RenderInput) (string, error) {
	layout, err := ParseLayout(input.LayoutJSON)
	if err != nil {
		return "", err
	}

	opts, err := ParseOptions(input.OptionsJSON)
	if err != nil {
		return "", err
	}

	data := PageData{
		Layout:    layout,
		Options:   opts,
		SEO:       input.SEO,
		Locale:    input.Locale,
		Locales:   input.Locales,
		PageSlugs: input.PageSlugs,
		Domain:    input.Domain,
		MediaBase: input.MediaBase,
	}

	return RenderPage(data)
}
