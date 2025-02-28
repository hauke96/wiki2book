package project

import (
	"encoding/json"
	"fmt"
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
	"os"
	"strings"
	"wiki2book/config"
)

type Project struct {
	config.Configuration
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

// LoadProject reads the given file and creates a corresponding Project instance. It also alters the config.Current
// object to override default configurations since project specific configs have a higher precedence.
func LoadProject(file string) (*Project, error) {
	projectString, err := os.ReadFile(file)
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
