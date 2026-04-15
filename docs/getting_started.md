# Getting Started

This page will help you get up and running with **yams**

### Prerequisites

- Go >= `1.24`

### Installation

**yams** can be used as a Go library via:
```shell
go get -u github.com/nsiow/yams
```

Similarly, the CLI for **yams** can be installed via:
```shell
go install github.com/nsiow/yams/cmd/yams@latest
```

!!! note

    If installing via `go install`, ensure that you have `go env GOPATH` somewhere in your `$PATH`

Alternatively, you can clone the [source](https://github.com/nsiow/yams.git) and run:
```
make && make install
```

By default, the `make` commands will install **yams** to `/usr/local/bin/`. If you wish to install
elsewhere or do not have sufficient permissions, you may need to either:

- Set the `YAMS_INSTALL_DIR` environment variable to an alternative location
- Run `make install` as root

### Running a Server

A local or remote instance of the **yams** server can be started via:
```shell
yams server \
  -source testdata/real-world/awsconfig.jsonl \
  -source testdata/real-world/org.jsonl
```

Server options:

- `-a/-addr`: Address and port to listen on (default: `:8888`)
- `-s/-source`: Data source(s) to load (supports multiple)
- `-r/-refresh`: Refresh interval in seconds for reloading sources (default: no refresh)
- `-e/-env`: Environment variables to report in the `/status` endpoint
- `-overlay`: Overlay store backend: `memory` (default) or `ddb://<table-name>` for DynamoDB

- For information about configuring sources, see [Data Sources](./data_sources.md)
- For information about generating data, see [Generating Data](./generating_data.md)

### Configuring the CLI

There are multiple ways to configure the **yams** CLI (in order of priority):

1. **Environment variable**: `YAMS_SERVER_ADDRESS`
2. **Config file**: `~/.config/yams/config.json`
3. **Command-line flag**: `-s/--server` for individual invocations

#### Config File

Create a config file at `~/.config/yams/config.json`:
```json
{
  "server": "localhost:8888",
  "format": "table"
}
```

Supported config options:

- `server`: Default server address
- `format`: Default output format (`json` or `table`)

#### Verifying Connectivity

Once configured, confirm connectivity using the `status` subcommand:
```shell
yams status
```
```json
{
  "accounts": 4,
  "entities": 1448,
  "groups": 0,
  "policies": 1371,
  "principals": 20,
  "resources": 53,
  "sources": [
    {
      "source": "testdata/real-world/awsconfig.jsonl",
      "updated": "2025-03-15T:04:35.173468943-07:00"
    },
    {
      "source": "testdata/real-world/org.jsonl",
      "updated": "2025-03-35T:04:35.173687682-07:00"
    }
  ]
}
```

### Shell Completion

**yams** supports shell completion for bash and zsh. To enable:

**Bash**:
```shell
# Add to ~/.bashrc
eval "$(yams completion bash)"
```

**Zsh**:
```shell
# Add to ~/.zshrc
eval "$(yams completion zsh)"
```

### Global Flags

The following flags can be used with any command:

- `-h, --help`: Show help
- `-v, --version`: Show version information
- `-V, --verbose`: Enable debug logging
