<h1 align="center">Nuon Extension: render</h1>

<p align="center">
  <a href="https://github.com/nuonco/nuon-ext-render/releases"><img src="https://img.shields.io/github/v/release/nuonco/nuon-ext-render?display_name=tag&amp;sort=semver" alt="Release"></a>
  <a href="https://go.dev/doc/devel/release"><img src="https://img.shields.io/badge/Go-1.25.0-00ADD8?logo=go&amp;logoColor=white" alt="Go Version"></a>
  <a href="https://pkg.go.dev/github.com/nuonco/nuon-ext-render"><img src="https://img.shields.io/badge/module-github.com%2Fnuonco%2Fnuon--ext--render-2C6BED" alt="Go Module"></a>
</p>

<p align="center">
  <a href="https://docs.nuon.co/guides">Nuon Docs</a>
  |
  <a href="https://docs.nuon.co/guides/cli-extensions">Nuon Extension Docs</a>
</p>

Nuon Extension: Utility to render app config files using an install's details from the ctl-api.

## Usage

```bash
nuon-ext-render --file <template-path> [--install-id <id>]
```

Output is written to stdout:

```bash
nuon-ext-render --file config.tpl > config.yaml
```

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `NUON_API_URL` | no | `https://ctl.prod.nuon.co` | API base URL |
| `NUON_API_TOKEN` | yes | | API auth token |
| `NUON_ORG_ID` | yes | | Organization ID |
| `NUON_INSTALL_ID` | no | | Install ID (can also use `--install-id` flag) |

An install context is required to render templates. You must provide an install ID via `NUON_INSTALL_ID` or `--install-id`.

## Template Variables

Templates use Go's [text/template](https://pkg.go.dev/text/template) syntax. The install state (from `GET /v1/installs/{install_id}/state`) is loaded under a root `.nuon` key, matching the convention used across nuon config files.

Example template:

```yaml
region: "{{ .nuon.install_stack.outputs.region }}"
db_host: "{{ .nuon.components.rds_cluster.outputs.address }}"
image: "{{ .nuon.components.img_app.outputs.image.repository }}:{{ .nuon.components.img_app.outputs.image.tag }}"
auth_url: '{{ or .nuon.inputs.inputs.auth_issuer_url "https://default.auth.com" }}'
```

All Go template features are supported (`or`, `eq`, `if`/`else`, pipelines, etc.).

## Development

### Run locally

The `run-local.sh` script reads credentials from `~/.nuon` (YAML-style `key: value` pairs) and exports them as `NUON_` env vars:

```bash
./scripts/run-local.sh --file example.tpl --install-id <id>
```

Override the config file:

```bash
NUON_CONFIG_FILE=~/.nuon-staging ./scripts/run-local.sh --file example.tpl --install-id <id>
```

### Build

```bash
./scripts/build.sh
```

### Manual testing with env vars

```bash
export NUON_API_URL=https://ctl.prod.nuon.co
export NUON_API_TOKEN=<your-token>
export NUON_ORG_ID=<your-org-id>
export NUON_INSTALL_ID=<your-install-id>

GOWORK=off go run . --file example.tpl
```
