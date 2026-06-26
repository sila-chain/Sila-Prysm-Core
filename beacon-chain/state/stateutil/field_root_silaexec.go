package stateutil

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// SilaExecutionRoot computes the HashTreeRoot Merkleization of
// a BeaconBlockHeader struct according to the Sila
// Simple Serialize specification.
func SilaExecutionRoot(silaexecData *silapb.SilaData) ([32]byte, error) {
	if silaexecData == nil {
		return [32]byte{}, errors.New("nil silaexec data")
	}
	return SilaDataRootWithHasher(silaexecData)
}

// SilaDataVotesRoot computes the HashTreeRoot Merkleization of
// a list of SilaData structs according to the Sila
// Simple Serialize specification.
func SilaDataVotesRoot(silaDataVotes []*silapb.SilaData) ([32]byte, error) {
	return SilaDatasRoot(silaDataVotes)
}
