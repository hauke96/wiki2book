package project

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Project struct {
	Metadata Metadata `json:"metadata"`
	Domain   string   `json:"wikipedia-domain"`
	Output   string   `json:"output"`
	Cover    string   `json:"cover"`
	Style    string   `json:"style"`
	Articles []string `json:"articles"`
}

type Metadata struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Author   string `json:"author"`
	License  string `json:"license"`
}

func LoadProject(file string) (*Project, error) {
	projectString, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error reading project file %s", file))
	}

	project := &Project{}
	err = json.Unmarshal(projectString, project)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing project file content")
	}

	return project, nil
}
