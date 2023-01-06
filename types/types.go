package types

type Data struct {
	DontCopyConfigToOutput bool                   `json:"dont-copy-config-to-output,omitempty"`
	CommitMessage          string                 `json:"commit-message,omitempty"`
	OutputBranch           string                 `json:"output-branch,omitempty"`
	OutputDir              string                 `json:"output-dir,omitempty"`
	ConfigBase64           string                 `json:"config-base64,omitempty"`
	Config                 map[string]interface{} `json:"config,omitempty"`
}
