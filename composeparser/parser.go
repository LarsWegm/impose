package composeparser

type Config struct {
	ComposeFilePath string
}

type Parser struct {
	composeFilePath string
}

type ImageVersion struct {
	ComposeKey string
	ImageName  string
	Version    string
}

func NewParser(cfg *Config) *Parser {
	return &Parser{
		composeFilePath: cfg.ComposeFilePath,
	}
}

func (p *Parser) GetImageVersions() []*ImageVersion {
	return []*ImageVersion{}
}
