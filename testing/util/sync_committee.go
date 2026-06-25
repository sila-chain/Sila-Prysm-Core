package util

import (
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// HydrateSyncCommittee hydrates the provided sync committee message.
func HydrateSyncCommittee(s *silapb.SyncCommitteeMessage) *silapb.SyncCommitteeMessage {
	if s.Signature == nil {
		s.Signature = make([]byte, 96)
	}
	if s.BlockRoot == nil {
		s.BlockRoot = make([]byte, fieldparams.RootLength)
	}
	return s
}

// ConvertToCommittee takes a list of pubkeys and returns a SyncCommittee with
// these keys as members. Some keys may appear repeated
func ConvertToCommittee(inputKeys [][]byte) *silapb.SyncCommittee {
	var pubKeys [][]byte
	for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSize; i++ {
		if i < uint64(len(inputKeys)) {
			pubKeys = append(pubKeys, bytesutil.PadTo(inputKeys[i], params.BeaconConfig().BLSPubkeyLength))
		} else {
			pubKeys = append(pubKeys, bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength))
		}
	}

	return &silapb.SyncCommittee{
		Pubkeys:         pubKeys,
		AggregatePubkey: bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength),
	}
}
