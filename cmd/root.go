package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"
)

var (
	// BuildVersion is injected at build time via ldflags.
	BuildVersion = "dev"
)

type config struct {
	APIURL    string
	APIToken  string
	OrgID     string
	InstallID string
}

func loadConfig() *config {
	apiURL := os.Getenv("NUON_API_URL")
	if apiURL == "" {
		apiURL = "https://ctl.prod.nuon.co"
	}

	return &config{
		APIURL:    apiURL,
		APIToken:  os.Getenv("NUON_API_TOKEN"),
		OrgID:     os.Getenv("NUON_ORG_ID"),
		InstallID: os.Getenv("NUON_INSTALL_ID"),
	}
}

func Execute() {
	cfg := loadConfig()

	root := &cobra.Command{
		Use:   "nuon-ext-render --file <path> [--install-id <id>]",
		Short: "Render config files using install details from the Nuon API",
		Long: fmt.Sprintf(`Render template files using an install's details from the ctl-api.

Version: %s

The install state is loaded under a root .nuon key, so templates
use {{.nuon.xxx}} to access state data. For example:

  {{.nuon.install_stack.outputs.region}}
  {{.nuon.components.rds_cluster.outputs.address}}
  {{.nuon.inputs.inputs.my_input_key}}

Output is written to stdout so it can be piped to a file:
  nuon-ext-render --file config.tpl > config.yaml`, BuildVersion),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, cfg)
		},
		SilenceUsage: true,
	}
	root.Version = BuildVersion

	root.Flags().String("file", "", "Path to the template file to render (required)")
	root.Flags().String("install-id", "", "Install ID (overrides NUON_INSTALL_ID env var)")
	root.MarkFlagRequired("file")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, cfg *config) error {
	filePath, _ := cmd.Flags().GetString("file")
	installID, _ := cmd.Flags().GetString("install-id")

	if installID == "" {
		installID = cfg.InstallID
	}
	if installID == "" {
		return fmt.Errorf("install ID is required: use --install-id flag or set NUON_INSTALL_ID")
	}

	tmplBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %w", filePath, err)
	}

	ctx := context.Background()
	client := &apiClient{
		baseURL:  cfg.APIURL,
		apiToken: cfg.APIToken,
		orgID:    cfg.OrgID,
	}

	state, err := client.getInstallState(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to get install state: %w", err)
	}

	data := map[string]any{
		"nuon": state,
	}

	tmpl, err := template.New("render").Option("missingkey=error").Funcs(sprig.FuncMap()).Parse(string(tmplBytes))
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	return tmpl.Execute(os.Stdout, data)
}

type apiClient struct {
	baseURL  string
	apiToken string
	orgID    string
}

func (c *apiClient) getInstallState(ctx context.Context, installID string) (map[string]any, error) {
	url := fmt.Sprintf("%s/v1/installs/%s/state", c.baseURL, installID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("X-Nuon-Org-ID", c.orgID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var state map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return state, nil
}
