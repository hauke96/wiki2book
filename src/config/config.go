package config

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
)

var Current = &Configuraution{}

type Configuraution struct {
	IgnoredTemplates []string `json:"ignored-templates"`
}

func LoadConfig(file string) error {
	projectString, err := os.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error reading config file %s", file))
	}

	err = json.Unmarshal(projectString, Current)
	if err != nil {
		return errors.Wrap(err, "Error parsing config file content")
	}

	return nil
}
