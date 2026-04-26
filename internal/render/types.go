package render

import "encoding/json"

type Layout []Section

type Section struct {
	ID      int64          `json:"id"`
	Blocks  []Block        `json:"blocks"`
	Mobile  BreakpointSize `json:"mobile"`
	Tablet  BreakpointSize `json:"tablet"`
	Desktop BreakpointSize `json:"desktop"`
	Style   SectionStyle   `json:"style"`
}

type BreakpointSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
	Gap  int `json:"gap"`
}

type BlockCoords struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type SectionStyle struct {
	MaxWidth          *int64     `json:"maxWidth"`
	FullWidth         bool       `json:"fullWidth"`
	HideOn            []string   `json:"hideOn"`
	SectionBackground Background `json:"section_background"`
	Background        Background `json:"background"`
	Padding           Sides      `json:"padding"`
	Margin            Sides      `json:"margin"`
}

type BlockStyle struct {
	HideOn     []string   `json:"hideOn"`
	Background Background `json:"background"`
	Border     Border     `json:"border"`
	Padding    Sides      `json:"padding"`
}

type Background struct {
	Mode       string `json:"mode"`
	LightColor string `json:"lightColor"`
	DarkColor  string `json:"darkcolorColor"`
	URLDesk    string `json:"url_desk"`
	URLTab     string `json:"url_tab"`
	URLMob     string `json:"url_mob"`
	Opacity    string `json:"opacity"`
	FixImgBack bool   `json:"fix_img_back"`
	Pos        string `json:"pos"`
	Size       string `json:"size"`
	Repeat     bool   `json:"repeat"`
	LightGradA string `json:"lightGradA"`
	LightGradB string `json:"lightGradB"`
	DarkGradA  string `json:"darkGradA"`
	DarkGradB  string `json:"darkGradB"`
	GradDeg    string `json:"gradDeg"`
	FocalX     string `json:"focalX"`
	FocalY     string `json:"focalY"`
	Zoom       string `json:"zoom"`
}

type Sides struct {
	T string `json:"t"`
	R string `json:"r"`
	B string `json:"b"`
	L string `json:"l"`
}

type Border struct {
	Radius       Corners       `json:"radius"`
	AllBorders   SideBorder    `json:"allBorders"`
	SidesBorders *SidesBorders `json:"sidesBorders"`
}

type Corners struct {
	TL string `json:"tl"`
	TR string `json:"tr"`
	BR string `json:"br"`
	BL string `json:"bl"`
}

type SidesBorders struct {
	L SideBorder `json:"l"`
	T SideBorder `json:"t"`
	R SideBorder `json:"r"`
	B SideBorder `json:"b"`
}

type SideBorder struct {
	Active bool   `json:"active"`
	Color  string `json:"color"`
	Thick  string `json:"thick"`
	Mode   string `json:"mode"`
}

type Block struct {
	ID                int64                  `json:"id"`
	Type              string                 `json:"type"`
	Locales           map[string]interface{} `json:"locales"`
	D                 BlockCoords            `json:"d"`
	M                 BlockCoords            `json:"m"`
	T                 BlockCoords            `json:"t"`
	Style             BlockStyle             `json:"style"`
	Color             *string                `json:"color"`
	DarkColor         *string                `json:"darkColor"`
	FontSize          *float64               `json:"fontSize"`
	LineHeight        *float64               `json:"lineHeight"`
	Divider           *SideBorder            `json:"divider"`
	Image             *BlockImage            `json:"image"`
	Link              *BlockLink             `json:"link"`
	Button            *BlockButton           `json:"button"`
	Icon              *BlockIcon             `json:"icon"`
	Menu              []MenuItem             `json:"menu"`
	MenuColors        *MenuColors            `json:"menuColors"`
	MenuFont          *Font                  `json:"menuFont"`
	MenuFontSize      *float64               `json:"menuFontSize"`
	MenuLineHeight    *float64               `json:"menuLineHeight"`
	MenuSubFont       *Font                  `json:"menuSubFont"`
	MenuSubFontSize   *float64               `json:"menuSubFontSize"`
	MenuSubLineHeight *float64               `json:"menuSubLineHeight"`
	IsMobileMenu      *bool                  `json:"isMobileMenu"`
}

type BlockImage struct {
	URLDesk string `json:"url_desk"`
	URLTab  string `json:"url_tab"`
	URLMob  string `json:"url_mob"`
	Fit     string `json:"fit"`
}

type BlockLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type BlockIcon struct {
	Name        string  `json:"name"`
	StrokeWidth float64 `json:"strokeWidth"`
}

type BlockButton struct {
	Bg              ButtonColors `json:"bg"`
	Hover           ButtonColors `json:"hover"`
	Active          ButtonColors `json:"active"`
	Focus           ButtonColors `json:"focus"`
	TextColor       ButtonColors `json:"textColor"`
	HoverTextColor  ButtonColors `json:"hoverTextColor"`
	ActiveTextColor ButtonColors `json:"activeTextColor"`
	BorderColor     ButtonColors `json:"borderColor"`
	Border          Border       `json:"border"`
	Width           string       `json:"width"`
	Padding         Sides        `json:"padding"`
}

type ButtonColors struct {
	Light string `json:"light"`
	Dark  string `json:"dark"`
}

type MenuColors struct {
	Color  ButtonColors `json:"color"`
	Hover  ButtonColors `json:"hover"`
	Active ButtonColors `json:"active"`
}

type MenuItem struct {
	Label    string     `json:"label"`
	Link     *BlockLink `json:"link"`
	Children []MenuItem `json:"children"`
}

type Font struct {
	Family string `json:"family"`
	Weight int    `json:"weight"`
	Italic *bool  `json:"italic"`
}

type SiteOptions struct {
	DarkColor        string            `json:"darkColor"`
	LightColor       string            `json:"lightColor"`
	DarkBackColor    string            `json:"darkBackColor"`
	LightBackColor   string            `json:"lightBackColor"`
	DarkAccentColor  string            `json:"darkAccentColor"`
	LightAccentColor string            `json:"lightAccentColor"`
	Fonts           []SiteFont        `json:"fonts"`
	GlobalFontFamily Font             `json:"globalFontFamily"`
	FontSize        float64           `json:"fontSize"`
	LineHeight      float64           `json:"lineHeight"`
	DarkMode        bool              `json:"darkMode"`
	MobileBP        int               `json:"mobileBP"`
	MobileTextZoom  float64           `json:"mobileTextZoom"`
	MobileMargin    float64           `json:"mobileMargin"`
	TabletBP        int               `json:"tabletBP"`
	TabletTextZoom  float64           `json:"tabletTextZoom"`
	TabletMargin    float64           `json:"tabletMargin"`
	DesktopTextZoom float64           `json:"desktopTextZoom"`
	DesktopMargin   float64           `json:"desktopMargin"`
	DesktopWidth    int               `json:"desktopWidth"`
	Headers         json.RawMessage   `json:"headers"`
	Header          *Section          `json:"header"`
	Footer          *Section          `json:"footer"`
}

type SiteFont struct {
	Family  string `json:"family"`
	Weights []int  `json:"weights"`
	Italics []int  `json:"italics"`
}

type PageData struct {
	Layout    Layout
	Options   SiteOptions
	SEO       *SEOData
	Locale    string
	Locales   []LocaleInfo
	PageSlugs []SlugInfo
	Domain    string
	MediaBase string
}

type SEOData struct {
	Title       string
	Description string
}

type LocaleInfo struct {
	Code      string
	IsDefault bool
}

type SlugInfo struct {
	LocaleCode string
	Slug       string
	IsDefault  bool
}

func ParseLayout(data []byte) (Layout, error) {
	var l Layout
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return l, nil
}

func ParseOptions(data []byte) (SiteOptions, error) {
	var o SiteOptions
	if err := json.Unmarshal(data, &o); err != nil {
		return o, err
	}
	return o, nil
}

func strPtr(s string) *string { return &s }
func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func floatPtr(f float64) *float64 { return &f }
