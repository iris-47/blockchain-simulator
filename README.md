# Blockchain Simulation Platform

[中文版README](README_zh.md)

---

## Project Overview

I am a graduate student, and this is a blockchain simulation platform built for the experiments in my thesis. The platform is designed to provide a modular and extensible environment for blockchain research and development. It currently supports the following transaction models:

- **UTXO Model**
- **Account Model**
- **Custom Smart Contract Transactions**

The platform's modular design allows easy swapping and integration of various consensus protocols, making it flexible for blockchain experiments and testing.

## Usage

### Adding Custom Protocols
1. Implement node modules in `node/runningMod`
2. Register new protocol combinations in `protocols.go`
3. See [node/runningMod/README.md](node/runningMod/README_zh.md) for development guide

## Quick Start

### 1. Build Project
```bash
go build
```
> **Note**: Requires `github.com/herumi/bls-go-binary/bls`, recommended to build before running
### 2. 启动节点
```bash
# Start client node (controller)
./BlockChainSimulator -c -m "TBB"
```
### 3. System Control
Client node will bootstrap the network

Use `Ctrl+C` for shutdown client and automatically shutdown the network

In case of unexpected failures, forcefully terminate all nodes using `./kill.sh`

Blockchain data: `blockchain_data/`

Log files: `log/`

## Key Features

- **Support for Multiple Transaction Models**: Includes UTXO, Account, and custom smart contract transactions.
- **Modular Design**: Consensus protocols and transaction processing modules are easily replaceable and extensible, catering to various experimental needs.
- **Extensibility**: Researchers can effortlessly integrate new functionalities and consensus protocols for testing and optimization.

## Current Status

The platform is still under development, with features being actively improved. I welcome collaboration and discussions with anyone interested in blockchain technology!

## Contact Information

If you have any questions or suggestions, feel free to reach out:
- **Email**: []
