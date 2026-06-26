package blocks

import (
	"fmt"

	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// SetSignature sets the signature of the signed beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSignature(sig []byte) {
	copy(b.signature[:], sig)
}

// SetSlot sets the respective slot of the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSlot(slot primitives.Slot) {
	b.block.slot = slot
}

// SetProposerIndex sets the proposer index of the beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetProposerIndex(proposerIndex primitives.ValidatorIndex) {
	b.block.proposerIndex = proposerIndex
}

// SetParentRoot sets the parent root of beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetParentRoot(parentRoot []byte) {
	copy(b.block.parentRoot[:], parentRoot)
}

// SetStateRoot sets the state root of the underlying beacon block
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetStateRoot(root []byte) {
	copy(b.block.stateRoot[:], root)
}

// SetRandaoReveal sets the randao reveal in the block body.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetRandaoReveal(r []byte) {
	copy(b.block.body.randaoReveal[:], r)
}

// SetGraffiti sets the graffiti in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetGraffiti(g []byte) {
	copy(b.block.body.graffiti[:], g)
}

// SetSilaData sets the silaexec data in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSilaData(e *eth.SilaData) {
	b.block.body.silaexecData = e
}

// SetProposerSlashings sets the proposer slashings in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetProposerSlashings(p []*eth.ProposerSlashing) {
	b.block.body.proposerSlashings = p
}

// SetAttesterSlashings sets the attester slashings in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetAttesterSlashings(slashings []eth.AttSlashing) error {
	if b.version < version.Electra {
		blockSlashings := make([]*eth.AttesterSlashing, 0, len(slashings))
		for _, slashing := range slashings {
			s, ok := slashing.(*eth.AttesterSlashing)
			if !ok {
				return fmt.Errorf("slashing of type %T is not *eth.AttesterSlashing", slashing)
			}
			blockSlashings = append(blockSlashings, s)
		}
		b.block.body.attesterSlashings = blockSlashings
	} else {
		blockSlashings := make([]*eth.AttesterSlashingElectra, 0, len(slashings))
		for _, slashing := range slashings {
			s, ok := slashing.(*eth.AttesterSlashingElectra)
			if !ok {
				return fmt.Errorf("slashing of type %T is not *eth.AttesterSlashingElectra", slashing)
			}
			blockSlashings = append(blockSlashings, s)
		}
		b.block.body.attesterSlashingsElectra = blockSlashings
	}
	return nil
}

// SetAttestations sets the attestations in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetAttestations(atts []eth.Att) error {
	if b.version < version.Electra {
		blockAtts := make([]*eth.Attestation, 0, len(atts))
		for _, att := range atts {
			a, ok := att.(*eth.Attestation)
			if !ok {
				return fmt.Errorf("attestation of type %T is not *eth.Attestation", att)
			}
			blockAtts = append(blockAtts, a)
		}
		b.block.body.attestations = blockAtts
	} else {
		blockAtts := make([]*eth.AttestationElectra, 0, len(atts))
		for _, att := range atts {
			a, ok := att.(*eth.AttestationElectra)
			if !ok {
				return fmt.Errorf("attestation of type %T is not *eth.AttestationElectra", att)
			}
			blockAtts = append(blockAtts, a)
		}
		b.block.body.attestationsElectra = blockAtts
	}
	return nil
}

// SetDeposits sets the deposits in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetDeposits(d []*eth.Deposit) {
	b.block.body.deposits = d
}

// SetVoluntaryExits sets the voluntary exits in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetVoluntaryExits(v []*eth.SignedVoluntaryExit) {
	b.block.body.voluntaryExits = v
}

// SetSyncAggregate sets the sync aggregate in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSyncAggregate(s *eth.SyncAggregate) error {
	if b.version == version.Phase0 {
		return consensus_types.ErrNotSupported("SyncAggregate", b.version)
	}
	b.block.body.syncAggregate = s
	return nil
}

// SetExecution sets the sila payload of the block body.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetExecution(e interfaces.SilaData) error {
	if b.version == version.Phase0 || b.version == version.Altair || b.version >= version.Gloas {
		return consensus_types.ErrNotSupported("Execution", b.version)
	}
	if e.IsBlinded() {
		b.block.body.silaPayloadHeader = e
		return nil
	}
	b.block.body.silaPayload = e
	return nil
}

// SetBLSToSilaChanges sets the BLS to Sila changes in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetBLSToSilaChanges(blsToSilaChanges []*eth.SignedBLSToSilaChange) error {
	if b.version < version.Capella {
		return consensus_types.ErrNotSupported("BLSToSilaChanges", b.version)
	}
	b.block.body.blsToSilaChanges = blsToSilaChanges
	return nil
}

// SetBlobKzgCommitments sets the blob kzg commitments in the block.
func (b *SignedBeaconBlock) SetBlobKzgCommitments(c [][]byte) error {
	if b.version < version.Deneb {
		return consensus_types.ErrNotSupported("SetBlobKzgCommitments", b.version)
	}
	b.block.body.blobKzgCommitments = c
	return nil
}

// SetSilaRequests sets the sila requests in the block.
func (b *SignedBeaconBlock) SetSilaRequests(req *silaenginev1.SilaRequests) error {
	if b.version < version.Electra || b.version >= version.Gloas {
		return consensus_types.ErrNotSupported("SetSilaRequests", b.version)
	}
	b.block.body.silaRequests = req
	return nil
}

// SetPayloadAttestations sets the payload attestations in the block.
func (b *SignedBeaconBlock) SetPayloadAttestations(pa []*eth.PayloadAttestation) error {
	if b.version < version.Gloas {
		return consensus_types.ErrNotSupported("SetPayloadAttestations", b.version)
	}
	b.block.body.payloadAttestations = pa
	return nil
}

// SetParentSilaRequests sets the parent sila requests in the block.
func (b *SignedBeaconBlock) SetParentSilaRequests(r *silaenginev1.SilaRequests) error {
	if b.version < version.Gloas {
		return consensus_types.ErrNotSupported("SetParentSilaRequests", b.version)
	}
	b.block.body.parentSilaRequests = r
	return nil
}

// SetSignedSilaPayloadBid sets the signed sila payload header in the block.
func (b *SignedBeaconBlock) SetSignedSilaPayloadBid(header *eth.SignedSilaPayloadBid) error {
	if b.version < version.Gloas {
		return consensus_types.ErrNotSupported("SetSignedSilaPayloadBid", b.version)
	}
	b.block.body.signedSilaPayloadBid = header
	return nil
}
