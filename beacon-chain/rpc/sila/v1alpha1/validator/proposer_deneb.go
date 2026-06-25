package validator

import (
	"errors"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// BuildBlobSidecars given a block, builds the blob sidecars for the block.
func BuildBlobSidecars(blk interfaces.ReadOnlySignedBeaconBlock, blobs [][]byte, kzgProofs [][]byte) ([]*silapb.BlobSidecar, error) {
	if blk.Version() < version.Deneb {
		return nil, nil // No blobs before deneb.
	}
	commits, err := blk.Block().Body().BlobKzgCommitments()
	if err != nil {
		return nil, err
	}
	cLen := len(commits)
	if cLen != len(blobs) || cLen != len(kzgProofs) {
		return nil, errors.New("blob KZG commitments don't match number of blobs or KZG proofs")
	}
	blobSidecars := make([]*silapb.BlobSidecar, cLen)
	header, err := blk.Header()
	if err != nil {
		return nil, err
	}
	body := blk.Block().Body()
	// Pre-compute subtrees once before the loop to avoid redundant calculations
	proofComponents, err := blocks.PrecomputeMerkleProofComponents(body)
	if err != nil {
		return nil, err
	}

	for i := range blobSidecars {
		proof, err := blocks.MerkleProofKZGCommitmentFromComponents(proofComponents, i)
		if err != nil {
			return nil, err
		}
		blobSidecars[i] = &silapb.BlobSidecar{
			Index:                    uint64(i),
			Blob:                     blobs[i],
			KzgCommitment:            commits[i],
			KzgProof:                 kzgProofs[i],
			SignedBlockHeader:        header,
			CommitmentInclusionProof: proof,
		}
	}
	return blobSidecars, nil
}
