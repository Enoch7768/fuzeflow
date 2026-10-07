package workflow

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const CurrentVersion = 1

func Parse(data []byte) (Definition, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return Definition{}, errors.New("workflow definition is empty")
	}

	var definition Definition
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, fmt.Errorf("parse workflow YAML: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Definition{}, errors.New("workflow definition must contain one YAML document")
		}
		return Definition{}, fmt.Errorf("parse workflow YAML: %w", err)
	}

	if definition.Version != CurrentVersion {
		return Definition{}, fmt.Errorf("unsupported workflow version %d", definition.Version)
	}
	if strings.TrimSpace(definition.Name) == "" {
		return Definition{}, errors.New("workflow name is required")
	}
	if len(definition.Jobs) == 0 {
		return Definition{}, errors.New("workflow must define at least one job")
	}

	return definition, nil
}
