package project

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
)

type Project struct {
	Metadata      Metadata `json:"metadata"`
	Domain        string   `json:"wikipedia-domain"`
	OutputFile    string   `json:"output-file"`
	CacheDir      string   `json:"cache-dir"`
	Cover         string   `json:"cover"`
	Style         string   `json:"style"`
	PandocDataDir string   `json:"pandoc-data-dir"`
	Articles      []string `json:"articles"`
}

type Metadata struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Author   string `json:"author"`
	License  string `json:"license"`
	Date     string `json:"date"`
}

func LoadProject(file string) (*Project, error) {
	projectString, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error reading project file %s", file))
	}

	project := &Project{
		CacheDir: ".wiki2book",
	}
	err = json.Unmarshal(projectString, project)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing project file content")
	}

	return project, nil
}
