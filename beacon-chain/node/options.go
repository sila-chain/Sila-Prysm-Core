package node

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/builder"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/filesystem"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
)

// Option for beacon node configuration.
type Option func(bn *BeaconNode) error

// WithBlockchainFlagOptions includes functional options for the blockchain service related to CLI flags.
func WithBlockchainFlagOptions(opts []blockchain.Option) Option {
	return func(bn *BeaconNode) error {
		bn.serviceFlagOpts.blockchainFlagOpts = opts
		return nil
	}
}

// WithSilaChainOptions includes functional options for the Sila chain service related to CLI flags.
func WithSilaChainOptions(opts []silaexec.Option) Option {
	return func(bn *BeaconNode) error {
		bn.serviceFlagOpts.silaChainFlagOpts = opts
		return nil
	}
}

// WithBuilderFlagOptions includes functional options for the builder service related to CLI flags.
func WithBuilderFlagOptions(opts []builder.Option) Option {
	return func(bn *BeaconNode) error {
		bn.serviceFlagOpts.builderOpts = opts
		return nil
	}
}

// WithBlobStorage sets the BlobStorage backend for the BeaconNode
func WithBlobStorage(bs *filesystem.BlobStorage) Option {
	return func(bn *BeaconNode) error {
		bn.BlobStorage = bs
		return nil
	}
}

// WithBlobStorageOptions appends 1 or more filesystem.BlobStorageOption on the beacon node,
// to be used when initializing blob storage.
func WithBlobStorageOptions(opt ...filesystem.BlobStorageOption) Option {
	return func(bn *BeaconNode) error {
		bn.BlobStorageOptions = append(bn.BlobStorageOptions, opt...)
		return nil
	}
}

func WithConfigOptions(opt ...params.Option) Option {
	return func(bn *BeaconNode) error {
		bn.ConfigOptions = append(bn.ConfigOptions, opt...)
		return nil
	}
}

// WithDataColumnStorage sets the DataColumnStorage backend for the BeaconNode
func WithDataColumnStorage(bs *filesystem.DataColumnStorage) Option {
	return func(bn *BeaconNode) error {
		bn.DataColumnStorage = bs
		return nil
	}
}

// WithDataColumnStorageOptions appends 1 or more filesystem.DataColumnStorageOption on the beacon node,
// to be used when initializing data column storage.
func WithDataColumnStorageOptions(opt ...filesystem.DataColumnStorageOption) Option {
	return func(bn *BeaconNode) error {
		bn.DataColumnStorageOptions = append(bn.DataColumnStorageOptions, opt...)
		return nil
	}
}
