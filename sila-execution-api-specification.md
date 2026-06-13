# Sila Execution API Specification

This document defines the Sila execution API identity used by Sila-Prysm when communicating with SilaChain.

## Scope

This specification documents the currently implemented Sila execution RPC surface.

It does not change consensus logic.
It does not remove Ethereum-derived compatibility dependencies.
It does not rename protocol/package identifiers required for interoperability.

## Authenticated Engine API

Sila-Prysm uses the Sila authenticated engine namespace:

- silaEngine_newPayloadV1
- silaEngine_newPayloadV2
- silaEngine_newPayloadV3
- silaEngine_newPayloadV4
- silaEngine_forkchoiceUpdatedV1
- silaEngine_forkchoiceUpdatedV2
- silaEngine_forkchoiceUpdatedV3
- silaEngine_getPayloadV1
- silaEngine_getPayloadV2
- silaEngine_getPayloadV3
- silaEngine_getPayloadV4
- silaEngine_exchangeTransitionConfigurationV1

Legacy `engine_*` methods are treated as compatibility-derived behavior and must not be reintroduced as the primary Sila-Prysm path.

## Public execution RPC used by Sila-Prysm

Sila-Prysm uses Sila public execution RPC methods for deposit-contract reads:

- sila_call
- sila_getCode

These replace practical dependency on:

- eth_call
- eth_getCode

## Compatibility status

The Sila execution API currently remains compatible with Ethereum Engine API semantics where required by consensus-layer behavior.

This compatibility must remain until a complete Sila-native replacement is implemented and tested.

## Do not remove yet

The following dependencies remain compatibility substrate and must not be removed yet:

- go-ethereum RPC/common/types imports
- ethereum.NotFound
- ethereum.CallMsg
- ethereum.FilterQuery
- ethereum.Subscription
- execution-apis references
- Engine API semantic compatibility

## Required before removing compatibility substrate

1. Sila-native execution RPC type layer
2. Sila-native call/filter/query abstractions
3. Sila-native Engine API test vectors
4. SilaChain + Sila-Prysm end-to-end tests covering all Sila engine methods
5. Migration plan for compatibility-derived go-ethereum types
