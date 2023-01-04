package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/konveyor/addon-move2kube/types"
	"github.com/konveyor/tackle2-addon/command"
)

// Build index.html
func runMove2Kube(input, output string, data types.Data) error {
	if err := os.RemoveAll(output); err != nil {
		return fmt.Errorf("failed to remove the output directory. Error: %w", err)
	}
	cmd := command.Command{Path: "/usr/bin/move2kube"}
	cmd.Options.Add("transform")
	cmd.Options.Add("--source", input)
	cmd.Options.Add("--output", output)
	cmd.Options.Add("--log-level", "trace")
	cmd.Options.Add("--qa-skip")
	if data.ConfigBase64 != "" {
		config, err := base64.StdEncoding.DecodeString(data.ConfigBase64)
		if err != nil {
			return fmt.Errorf("failed to decode the config as base64. Error: %w", err)
		}
		configPath := "config.yaml"
		if err := os.WriteFile(configPath, config, 0660); err != nil {
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
