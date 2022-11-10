package composeparser

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type parser struct {
	file          string
	yamlContent   yaml.Node
	imageVerisons []*ImageVersion
}

type ImageVersion struct {
	Service   string
	ImageName string
	Version   string
	yamlNode  *yaml.Node
}

func NewParser(file string) (*parser, error) {
	if file == "" {
		return nil, errors.New("file must be set")
	}

	p := &parser{}
	err := p.loadFromFile(file)

	return p, err
}

func (p *parser) GetImageVersions() ([]*ImageVersion, error) {
	/*
		for _, v := range p.imageVerisons {
			fmt.Println(v)
		}
	*/
	return p.imageVerisons, nil
}

func (p *parser) loadFromFile(file string) error {
	p.file = file
	yamlFile, err := os.ReadFile(p.file)
	normLineEndings := strings.Replace(string(yamlFile), "\r\n", "\n", -1)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(normLineEndings), &p.yamlContent)
	if err != nil {
		return err
	}

	if len(p.yamlContent.Content) < 1 {
		return errors.New("invalid YAML content")
	}

	servicesNode, err := getNodeByKey(p.yamlContent.Content[0], "services")
	if err != nil {
		return err
	}

	servicesNodeContent := servicesNode.Content
	servicesNodeContentLen := len(servicesNodeContent)
	for i := 0; i < servicesNodeContentLen; i = i + 2 {
		if servicesNodeContentLen <= i+1 {
			return errors.New("could not parese YAML: invalid services node content length")
		}
		imgVersion, err := getImageVersionFromNode(servicesNodeContent[i].Value, servicesNodeContent[i+1])
		if err != nil {
			return err
		}
		p.imageVerisons = append(p.imageVerisons, imgVersion)
	}
	return nil
}

func (p *parser) writeToFile() error {
	b, err := yaml.Marshal(&p.yamlContent)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	//os.WriteFile("out.yml", b, 0644)
	//os.WriteFile(p.file, b, 0644)
	return nil
}

func getImageVersionFromNode(serviceName string, node *yaml.Node) (*ImageVersion, error) {
	imgVersion := &ImageVersion{
		Service: serviceName,
	}

	imageNode, err := getNodeByKey(node, "image")
	if err != nil {
		return nil, err
	}
	imgVersion.yamlNode = imageNode

	imgVal := imageNode.Value
	imgSlice := strings.Split(imgVal, ":")
	imgSliceLen := len(imgSlice)

	if imgSliceLen != 1 && imgSliceLen != 2 {
		return nil, fmt.Errorf("image has invalid format: '%v'", imgVal)
	}

	imgVersion.ImageName = imgSlice[0]

	if imgSliceLen > 1 {
		imgVersion.Version = imgSlice[1]
	}

	return imgVersion, nil
}

func getNodeByKey(node *yaml.Node, key string) (*yaml.Node, error) {
	nodeContent := node.Content
	for i, n := range nodeContent {
		if n.Value == key {
			if len(nodeContent) <= i+1 {
				return nil, errors.New("could not parese YAML: invalid node content length")
			}
			return nodeContent[i+1], nil
		}
	}
	return nil, fmt.Errorf("no key '%v' found in YAML", key)
}
