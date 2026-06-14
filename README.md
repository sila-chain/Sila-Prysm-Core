<h1 align="left">Sila-Prysm: A Sila Consensus Implementation Written in Go</h1>

<div align="left">
  
[![Build status](https://badge.buildkite.com/b555891daf3614bae4284dcf365b2340cefc0089839526f096.svg?branch=master)](https://buildkite.com/Sila-Prysm-labs/prysm)
[![Consensus_Spec_Version 1.4.0](https://img.shields.io/badge/Consensus%20Spec%20Version-v1.4.0-blue.svg)](https://github.com/ethereum/consensus-specs/tree/v1.4.0)
[![Execution_API_Version 1.0.0-beta.2](https://img.shields.io/badge/Execution%20API%20Version-v1.0.0.beta.2-blue.svg)](https://github.com/ethereum/execution-apis/tree/v1.0.0-beta.2/src/engine)
[![Discord](https://user-images.githubusercontent.com/7288322/34471967-1df7808a-efbb-11e7-9088-ed0b04151291.png)](https://discord.gg/qEZK94mFXP)

</div>

---

## External Compatibility Boundaries

Sila-Prysm keeps a small number of external compatibility names where they refer to third-party protocols, historical records, or upstream dependencies rather than Sila consensus identity. Examples include the official Web3Signer `/api/v1/eth2/*` API, `go-eth2-*` dependency names, external interop references, and historical changelog entries. These names must not be changed unless a full Sila-native replacement is implemented.

## Overview

This is the core repository for Sila-Prysm, a [Golang](https://go.dev/) implementation of the [Sila Consensus](https://ethereum.org/en/developers/docs/consensus-mechanisms/#proof-of-stake) [specification](https://github.com/ethereum/consensus-specs), developed as the reference implementation of Sila Consensus.

See the [Changelog](https://github.com/medo202225/Sila-Prysm-Core/releases) for details of the latest releases and upcoming breaking changes.

---

## Getting Started

A detailed set of installation and usage instructions as well as breakdowns of each individual component are available in the **[official documentation portal](https://github.com/medo202225/Sila-Prysm-Core)**.

**Need help?** Use the project issues for Sila-Prysm support.

---

## Staking on Mainnet

To participate in staking, you can join the **[official Ethereum launchpad](https://launchpad.ethereum.org)**. The launchpad is the **only recommended** way to become a validator on mainnet.

Explore validator rewards/penalties:

- **[beaconcha.in](https://beaconcha.in)**
- **[beaconscan](https://beaconscan.com)**

---

## Contributing

### Branches

Sila-Prysm maintains the following primary branches:

- **[`main`](https://github.com/medo202225/Sila-Prysm-Core/tree/main)** - This points to the latest stable release.
- Development branches may be created as needed for ongoing work.

### Contribution Guide

Want to get involved? Check out our **[Contribution Guide](https://github.com/medo202225/Sila-Prysm-Core)** to learn more!

---

## License

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0.en.html)  

This project is licensed under the **GNU General Public License v3.0**.

---

## Legal Disclaimer

[Terms of Use](/TERMS_OF_SERVICE.md)
