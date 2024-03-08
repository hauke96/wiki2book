package project

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"wiki2book/config"
	"wiki2book/generator"
)

type Project struct {
	Metadata          Metadata `json:"metadata"`
	WikipediaInstance string   `json:"wikipedia-instance"`
	OutputFile        string   `json:"output-file"`
	OutputType        string   `json:"output-type"`
	OutputDriver      string   `json:"output-driver"`
	CacheDir          string   `json:"cache-dir"`
	CoverImage        string   `json:"cover-image"`
	StyleFile         string   `json:"style-file"`
	PandocDataDir     string   `json:"pandoc-data-dir"`
	Articles          []string `json:"articles"`
	FontFiles         []string `json:"font-files"`
	ImagesToGrayscale bool     `json:"images-to-grayscale"`
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

	project := NewWithDefaults()
	err = json.Unmarshal(projectString, project)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing project file content")
	}

	// Override default configs with project specific ones
	if project.WikipediaInstance != "" {
		config.Current.WikipediaInstance = project.WikipediaInstance
	}

	return project, nil
}

func NewWithDefaults() *Project {
	return &Project{
		CacheDir:     ".wiki2book",
		OutputType:   "epub2",
		OutputDriver: generator.OutputDriverPandoc,
	}
}
