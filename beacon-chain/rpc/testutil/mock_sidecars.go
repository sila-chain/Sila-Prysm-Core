package testutil

import silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"

// CreateDataColumnSidecar generates a filled dummy data column sidecar
func CreateDataColumnSidecar(index uint64, data []byte) *silapb.DataColumnSidecar {
	return &silapb.DataColumnSidecar{
		Index:  index,
		Column: [][]byte{data},
		SignedBlockHeader: &silapb.SignedBeaconBlockHeader{
			Header: &silapb.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 1,
				ParentRoot:    make([]byte, 32),
				StateRoot:     make([]byte, 32),
				BodyRoot:      make([]byte, 32),
			},
			Signature: make([]byte, 96),
		},
		KzgCommitments:               [][]byte{make([]byte, 48)},
		KzgProofs:                    [][]byte{make([]byte, 48)},
		KzgCommitmentsInclusionProof: [][]byte{make([]byte, 32)},
	}
}

// CreateBlobSidecar generates a filled dummy data blob sidecar
func CreateBlobSidecar(index uint64, blob []byte) *silapb.BlobSidecar {
	return &silapb.BlobSidecar{
		Index: index,
		Blob:  blob,
		SignedBlockHeader: &silapb.SignedBeaconBlockHeader{
			Header: &silapb.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 1,
				ParentRoot:    make([]byte, 32),
				StateRoot:     make([]byte, 32),
				BodyRoot:      make([]byte, 32),
			},
			Signature: make([]byte, 96),
		},
		KzgCommitment: make([]byte, 48),
		KzgProof:      make([]byte, 48),
	}
}
