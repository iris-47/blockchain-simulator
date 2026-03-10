# Blockchain Simulation Platform

[Chinese version](README_zh.md)

---

## Project Overview

I am a graduate student, and this is a blockchain simulation platform I built for my thesis experiments. It aims to provide a modular and easily extensible experimental environment for blockchain research. The platform supports the following transaction models:

- **UTXO model**
- **Account model**
- **Custom smart contract transactions**

With its modular design, the platform facilitates the replacement and integration of different consensus protocols, flexibly supporting various blockchain experiments and tests.

## Usage

### Adding Custom Protocols
1. Implement your node behavior module plugins in the `node/plugins` directory.
2. Each plugin should be a folder and must include an exported struct `PluginPackage`.
3. Register new protocol combinations in `protocols.json`.
4. For detailed instructions, refer to [node/README_zh.md](node/README_zh.md).

### Configuration
All configuration parameters can be set via command line or the `config.json` configuration file.

## Quick Start

### 1. Build the Project
```bash
make
```
Use `make help` for assistance. Of course, you can also compile using `go build` or write your own build script instead of using the Makefile.
> **Note**: This project depends on the `github.com/herumi/bls-go-binary/bls` library. It is recommended to run after compilation for optimal performance.

### 2. Start a Node
```bash
# Start the client
./BlockChainSimulator -c -m "TBB"
```

### 3. System Control
If the `StartSystem` and `StopSystem` plugins are loaded, the client node will automatically initialize the entire network after startup. These are useful common plugins—please use them.

In case of unexpected situations, use the `pkill` command to forcibly terminate all nodes (or similar commands).

Blockchain data is stored in `blockchain_data/`.

Logs are stored in `log/`.

## Project Features

- **Support for Multiple Transaction Models**: Includes UTXO model, Account model, and custom smart contract transactions.
- **Modular Design**: Consensus protocols and transaction processing modules can be easily replaced and extended to meet different experimental needs.
- **Extensibility**: Researchers can easily integrate new functionalities and consensus protocols for experimentation and optimization.

## Current Status

The platform is currently under development, with features continuously being improved. Feedback and discussions from those interested in blockchain technology are welcome!

## Contact

If you have any questions or suggestions, feel free to contact me:
- **Email**: []