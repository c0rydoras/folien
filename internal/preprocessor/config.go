package preprocessor

type Config struct {
	TOCTitle       string
	EnableHeadings bool
}

func NewConfig() *Config {
	return &Config{
		TOCTitle:       "",
		EnableHeadings: false,
	}
}

func (c *Config) WithTOC(title string) *Config {
	c.TOCTitle = title
	return c
}

func (c *Config) WithHeadings() *Config {
	c.EnableHeadings = true
	return c
}

func (c *Config) Process(slides []string) []string {
	result := slides

	if c.EnableHeadings {
		result = AddHeadings(result, 2)
	}

	if c.TOCTitle != "" {
		result = append([]string{GenerateTOC(slides, c.TOCTitle)}, result...)
	}

	return result
}
