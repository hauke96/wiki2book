package project

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"wiki2book/config"
)

type Project struct {
	Metadata      Metadata `json:"metadata"`
	WikipediaUrl  string   `json:"wikipedia-url"`
	OutputFile    string   `json:"output-file"`
	CacheDir      string   `json:"cache-dir"`
	Cover         string   `json:"cover"`
	Style         string   `json:"style"`
	PandocDataDir string   `json:"pandoc-data-dir"`
	Articles      []string `json:"articles"`
	FontFiles     []string `json:"font-files"`
}

type Metadata struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Author   string `json:"author"`
	License  string `json:"license"`
	Date     string `json:"date"`
}

// LoadProject reads the given file and creates a corresponding Project instance. It also alters the config.Current
// object to override default configurations since project specific configs have a higher precedence.
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

	// Override default configs with project specific ones
	if project.WikipediaUrl != "" {
		config.Current.WikipediaUrl = project.WikipediaUrl
	}

	return project, nil
}
