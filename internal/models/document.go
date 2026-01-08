package models

type Document struct {
	Title    string
	Author   string
	Date     string
	Chapters []Chapter
	Config   Config
}

type Chapter struct {
	Title   string
	Content string // Markdown content
	Path    string
	Order   string // e.g., "1.1"
}

type Config struct {
	HeaderTitle     string `yaml:"header_title"`
	FooterText      string `yaml:"footer_text"`
	HeaderLogo      string `yaml:"header_logo"`
	FooterLogo      string `yaml:"footer_logo"`
	TitleImage      string `yaml:"title_image"`
	ShowTOC         bool   `yaml:"show_toc"`
	Theme           string `yaml:"theme"`
	FontSize        string `yaml:"font_size"`
	FontZipPath     string `yaml:"font_zip_path"`
	PageNumberStart int    `yaml:"page_number_start"`
	TitleAlign      string `yaml:"title_align"` // center, left, right, justify
	CodeTheme       string `yaml:"code_theme"`
	TitleColor      string `yaml:"title_color"`
	HeaderColor     string `yaml:"header_color"`
}
