# Getting Started

This page will help you get up and running with **yams**

### Prerequisites

- Go >= `1.24`

### Installation

**yams** can be used as a Go library via:
```shell
go get -u github.com/nsiow/yams@latest
```

Similarly, the CLI for **yams** can be installed via:
```shell
go install github.com/nsiow/yams/cmd/yams@latest
```

Alternatively, you can clone the [source](https://github.com/nsiow/yams.git) and run:
```
make && make install
```

By default, this will install **yams** to `/usr/local/bin/`. If you wish to install elsewhere or
do not have sufficient permissions, you may need to either:

* Set the `YAMS_INSTALL_DIR` environment variable to an alternative location
* Run `make install` as root

### Running a Server

A local or remote instance of the **yams** server can be started via:
```shell
yams server \
  -source testdata/real-world/awsconfig.jsonl \
  -source testdata/real-world/org.jsonl
```

- For information about configuring sources, see [Data Sources](./data_sources.md)
- For information about generating data, see [Generating Data](./generating_data.md)

### Configuring the CLI

There are two options for pointing the **yams** CLI at the desired server:

- via the `-s/--server` flag for individual invocations
- via the `YAMS_SERVER_ADDRESS` environment variable

Once set, you should be able to confirm connectivity using the `status` subcommand:
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
