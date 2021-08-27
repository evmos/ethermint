<!--
order: 1
-->

# Installation

Build and install the Ethermint binaries from source or using Docker. {synopsis}

## Pre-requisites

- [Install Go 1.16+](https://golang.org/dl/) {prereq}
- [Install jq](https://stedolan.github.io/jq/download/) {prereq}

## Install Binaries

### GitHub

Clone and build Ethermint using `git`:

```bash
git clone https://github.com/tharsis/ethermint.git
cd ethermint
make install
```

Check that the binaries have been successfully installed:

```bash
ethermintd version
```

### Docker

You can build Ethermint using Docker by running:

```bash
make docker-build
```

This will install the binaries on the `./build` directory. Now, check that the binaries have been
successfully installed:

```bash
ethermintd version
```

### Releases

You can also download a specific release available on the Ethermint [repository](https://github.com/tharsis/ethermint/releases) or via command line:

```bash
go install github.com/tharsis/ethermint@latest
```
