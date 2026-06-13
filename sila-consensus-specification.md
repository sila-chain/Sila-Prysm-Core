# Sila Consensus Specification

This document defines the current Sila consensus identity used by Sila-Prysm.

## Scope

This specification documents Sila consensus identity and compatibility boundaries.

It does not change consensus logic.
It does not remove Ethereum-derived compatibility dependencies.
It does not rename protocol/package identifiers required for interoperability.
It does not replace consensus-specs references until Sila-native test vectors exist.

## Current Sila consensus identity

Sila-Prysm runs with the Sila network flag:

- --sila

Sila-Prysm identifies the network as:

- Sila Beacon Chain Mainnet

The current Sila consensus runtime uses Sila-specific configuration while preserving consensus behavior compatibility where required.

## Preserved compatibility

The following remain compatibility substrate:

- Beacon chain state transition semantics
- Fork choice behavior
- Slashing protection interchange behavior
- SSZ compatibility
- BLS and KZG compatibility
- consensus-specs-derived behavior
- external validator and beacon API behavior where required

## Sila consensus naming

User-facing and documentation-facing identity should use Sila terminology:

- Sila Beacon Chain
- Sila consensus
- Sila validator
- Sila execution chain
- Sila execution API
- Sila P2P

## Do not remove yet

The following must remain until Sila-native replacements exist:

- consensus-specs references
- EIP references
- SSZ compatibility terminology when protocol-critical
- protobuf/gRPC external package identifiers
- Ethereum-derived test vectors
- go-ethereum compatibility types used by execution integration

## Required before removing compatibility substrate

1. Sila consensus test vectors
2. Sila fork schedule specification
3. Sila state transition reference document
4. Sila SSZ compatibility statement
5. Sila slashing protection interchange statement
6. Sila validator API compatibility statement
7. Sila end-to-end consensus testnet validation
