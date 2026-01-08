package config

type Config struct {
	Title       string      `yaml:"title" validate:"required"`
	Subtitle    string      `yaml:"subtitle"`
	Header      Header      `yaml:"header"`
	Footer      Footer      `yaml:"footer"`
	Colors      Colors      `yaml:"colors"`
	Theme       string      `yaml:"theme"`
	Fonts       Fonts       `yaml:"fonts" validate:"required"`
	FontSize    float64     `yaml:"font_size" validate:"required,gt=0"`
	PageNumbers PageNumbers `yaml:"page_numbers"`
	Layout      Layout      `yaml:"layout"`
	CodeTheme   string      `yaml:"code_theme"`
	Mermaid     Mermaid     `yaml:"mermaid"`
}

type Header struct {
	Text  string `yaml:"text"`
	Image string `yaml:"image"`
}

type Footer struct {
	Text  string `yaml:"text"`
	Image string `yaml:"image"`
}

type Colors struct {
	Title      string `yaml:"title"`  // Default #C00000
	Header     string `yaml:"header"` // Default #C00000
	Background string `yaml:"background"`
	Text       string `yaml:"text"`
	Accent     string `yaml:"accent"`
}

type Fonts struct {
	Zip     string `yaml:"zip" validate:"required"`
	Regular string `yaml:"regular" validate:"required"`
	Bold    string `yaml:"bold"`
	Italic  string `yaml:"italic"`
	Mono    string `yaml:"mono"`
}

type PageNumbers struct {
	StartPage int `yaml:"start_page"`
}

type Layout struct {
	StartPage string  `yaml:"startpage" validate:"oneof=left center right justify"`
	Body      string  `yaml:"body" validate:"oneof=left center right justify"`
	Margins   Margins `yaml:"margins"`
}

type Margins struct {
	Left   float64 `yaml:"left"`
	Right  float64 `yaml:"right"`
	Top    float64 `yaml:"top"`
	Bottom float64 `yaml:"bottom"`
}

type Mermaid struct {
	Renderer string `yaml:"renderer"` // Default mmdc
}
