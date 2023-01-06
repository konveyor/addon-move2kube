package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/konveyor/addon-move2kube/types"
	"github.com/konveyor/tackle2-addon/command"
	"gopkg.in/yaml.v3"
)

// Build index.html
func runMove2Kube(input, output string, data types.Data) error {
	if err := os.RemoveAll(output); err != nil {
		return fmt.Errorf("failed to remove the output directory. Error: %w", err)
	}
	if data.Config != nil && data.ConfigBase64 != "" {
		return fmt.Errorf("cannot specify both config and config base64 at the same time")
	}
	cmd := command.Command{Path: "/usr/bin/move2kube"}
	cmd.Options.Add("transform")
	cmd.Options.Add("--source", input)
	cmd.Options.Add("--output", output)
	cmd.Options.Add("--log-level", "trace")
	cmd.Options.Add("--qa-skip")
	if data.Config != nil {
		configPath := "config.yaml"
		configYaml, err := yaml.Marshal(data.Config)
		if err != nil {
			return fmt.Errorf("failed to marshal the config to yaml. Error: %w", err)
		}
		if err := os.WriteFile(configPath, configYaml, 0664); err != nil {
			return fmt.Errorf("failed to write the config to a file at path '%s' . Error: %w", configPath, err)
		}
		cmd.Options.Add("--config", configPath)
		addon.Activity("running Move2Kube transform with the config: '%s'", string(configYaml))
	}
	if data.ConfigBase64 != "" {
		config, err := base64.StdEncoding.DecodeString(data.ConfigBase64)
		if err != nil {
			return fmt.Errorf("failed to decode the config as base64. Error: %w", err)
		}
		configPath := "config.yaml"
		if err := os.WriteFile(configPath, config, 0664); err != nil {
			return fmt.Errorf("failed to write the config to a file at path '%s' . Error: %w", configPath, err)
		}
		cmd.Options.Add("--config", configPath)
		addon.Activity("running Move2Kube transform with the config: '%s'", string(config))
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run the Move2Kube transform command. Error: %w", err)
	}
	return nil
}
