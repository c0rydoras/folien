package preprocessor

type Config struct {
	TOCTitle       string
	TOCDescription string
	EnableHeadings bool
}

func NewConfig() *Config {
	return &Config{
		TOCTitle:       "",
		TOCDescription: "",
		EnableHeadings: false,
	}
}

func (c *Config) WithTOC(title string, description string) *Config {
	c.TOCTitle = title
	c.TOCDescription = description
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
		result = append([]string{GenerateTOC(slides, c.TOCTitle, c.TOCDescription)}, result...)
	}

	return result
}
