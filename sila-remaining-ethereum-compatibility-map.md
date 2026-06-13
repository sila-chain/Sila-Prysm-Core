# Sila Remaining Ethereum Compatibility Map

This document tracks remaining Ethereum-derived compatibility dependencies in Sila-Prysm.

Rules:
- Do not remove Ethereum-derived dependencies until a 100% equivalent Sila replacement exists.
- Do not change consensus logic.
- Do not change execution compatibility.
- Do not rename protocol/package identifiers that are required for interoperability.

## Compatibility categories

### go-ethereum imports
Used for common types, hex utilities, RPC helpers, execution payload types, KZG helpers, ENR/enode, and compatibility with execution-layer interfaces.

Status: keep until Sila equivalents exist.

### EIPs
Used as normative references for keystores, slashing interchange, deposits, withdrawals, and consensus behavior.

Status: keep until Sila standards exist.

### consensus-specs references
Used as normative consensus behavior references.

Status: keep until Sila Consensus Specification exists.

### execution-apis references
Used as normative Engine API behavior references.

Status: keep until Sila Execution API Specification exists.

### protobuf / gRPC ethereum package names
Some package names are part of external API compatibility.

Status: keep until Sila protobuf/gRPC replacement plan exists.

## Required Sila replacements before removal

1. Sila Consensus Specification
2. Sila Execution API Specification
3. Sila P2P / ENR Specification
4. Sila protobuf/gRPC identity plan
5. Sila go-ethereum compatibility layer
6. Sila KZG / cryptography compatibility strategy
