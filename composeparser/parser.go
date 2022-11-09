package composeparser

import (
	"errors"
	"fmt"
)

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

func NewParser(cfg *Config) (*Parser, error) {
	if cfg.ComposeFilePath == "" {
		return nil, errors.New("ComposeFilePath must be set")
	}

	return &Parser{
		composeFilePath: cfg.ComposeFilePath,
	}, nil
}

func (p *Parser) GetImageVersions() []*ImageVersion {
	fmt.Println(p.composeFilePath)
	return []*ImageVersion{}
}
