# Sila Proto Shadow Layer

Goal: introduce Sila-native protobuf identity while preserving Ethereum protobuf compatibility until all callers migrate safely.

## Current compatibility source

The current project still depends on:

- proto/eth/ext
- proto/eth/v1

These packages are used by:

- proto/engine/v1
- proto/prysm/v1alpha1
- proto/migration
- beacon-chain packages
- rpc packages
- p2p packages

## Migration strategy

### Phase 1: Shadow ext package

Create Sila-native equivalent:

- proto/sila/ext

It must be functionally identical to proto/eth/ext at first.

Rules:

- no behavior change
- no generated API breakage
- no deletion of proto/eth/ext
- proto/eth/ext remains compatibility

### Phase 2: Shadow v1 package

Create Sila-native equivalent:

- proto/sila/v1

It must mirror proto/eth/v1 initially.

Rules:

- no message field changes
- no wire format changes
- no field number changes
- no JSON name changes

### Phase 3: Internal migration

Move internal imports gradually:

- proto/eth/v1 -> proto/sila/v1
- proto/eth/ext -> proto/sila/ext

Only after each package passes tests.

### Phase 4: Compatibility removal

Remove proto/eth only after:

- all internal imports are migrated
- generated files compile
- beacon-chain builds
- prysmctl builds
- validator builds
- RPC tests pass
- proto migration tests pass

## Hard rule

No broad replace across the repository. Every migration must be scoped, tested, committed, and pushed.
