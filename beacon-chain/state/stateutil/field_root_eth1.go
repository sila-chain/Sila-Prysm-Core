package stateutil

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// Eth1Root computes the HashTreeRoot Merkleization of
// a BeaconBlockHeader struct according to the Sila
// Simple Serialize specification.
func Eth1Root(eth1Data *silapb.Eth1Data) ([32]byte, error) {
	if eth1Data == nil {
		return [32]byte{}, errors.New("nil eth1 data")
	}
	return Eth1DataRootWithHasher(eth1Data)
}

// Eth1DataVotesRoot computes the HashTreeRoot Merkleization of
// a list of Eth1Data structs according to the Sila
// Simple Serialize specification.
func Eth1DataVotesRoot(eth1DataVotes []*silapb.Eth1Data) ([32]byte, error) {
	return Eth1DatasRoot(eth1DataVotes)
}
