package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

type Project struct {
	Configuration
	Metadata   Metadata `json:"metadata"`
	OutputFile string   `json:"output-file"`
	Articles   []string `json:"articles"`
}

type Metadata struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Author   string `json:"author"`
	License  string `json:"license"`
	Date     string `json:"date"`
}

func (p *Project) Print() {
	jsonBytes, err := json.MarshalIndent(p.Metadata, "  ", "  ")
	sigolo.FatalCheck(err)
	sigolo.Debugf("Project:\n  OutputFile: %s\n  Articles: %v\n  Metadata: %s", p.OutputFile, strings.Join(p.Articles, ", "), string(jsonBytes))
}

// LoadProjectFromFile reads the given file and creates a corresponding Project instance.
func LoadProjectFromFile(file string) (*Project, error) {
	projectString, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error reading project file %s", file))
	}

	return LoadProjectFromBytes(projectString)
}

// LoadProjectFromBytes creates a project instance from the given bytes of a project JSON string.
func LoadProjectFromBytes(projectJsonBytes []byte) (*Project, error) {
	project := &Project{}
	project.Configuration = *NewDefaultConfig()
	err := json.Unmarshal(projectJsonBytes, project)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing project file content")
	}

	return project, nil
}
