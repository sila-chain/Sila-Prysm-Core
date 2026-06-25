package eth

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
)

type copier[T any] interface {
	Copy() T
}

func CopySlice[T any, C copier[T]](original []C) []T {
	// Create a new slice with the same length as the original
	newSlice := make([]T, len(original))
	for i := range newSlice {
		newSlice[i] = original[i].Copy()
	}
	return newSlice
}

// CopyValidator copies the provided validator.
func CopyValidator(val *Validator) *Validator {
	pubKey := make([]byte, len(val.PublicKey))
	copy(pubKey, val.PublicKey)
	withdrawalCreds := make([]byte, len(val.WithdrawalCredentials))
	copy(withdrawalCreds, val.WithdrawalCredentials)
	return &Validator{
		PublicKey:                  pubKey,
		WithdrawalCredentials:      withdrawalCreds,
		EffectiveBalance:           val.EffectiveBalance,
		Slashed:                    val.Slashed,
		ActivationEligibilityEpoch: val.ActivationEligibilityEpoch,
		ActivationEpoch:            val.ActivationEpoch,
		ExitEpoch:                  val.ExitEpoch,
		WithdrawableEpoch:          val.WithdrawableEpoch,
	}
}

// CopyBuilder copies the provided builder.
func CopyBuilder(builder *Builder) *Builder {
	if builder == nil {
		return nil
	}
	return &Builder{
		Pubkey:            bytesutil.SafeCopyBytes(builder.Pubkey),
		Version:           bytesutil.SafeCopyBytes(builder.Version),
		ExecutionAddress:  bytesutil.SafeCopyBytes(builder.ExecutionAddress),
		Balance:           builder.Balance,
		DepositEpoch:      builder.DepositEpoch,
		WithdrawableEpoch: builder.WithdrawableEpoch,
	}
}

// CopySyncCommitteeMessage copies the provided sync committee message object.
func CopySyncCommitteeMessage(s *SyncCommitteeMessage) *SyncCommitteeMessage {
	if s == nil {
		return nil
	}
	return &SyncCommitteeMessage{
		Slot:           s.Slot,
		BlockRoot:      bytesutil.SafeCopyBytes(s.BlockRoot),
		ValidatorIndex: s.ValidatorIndex,
		Signature:      bytesutil.SafeCopyBytes(s.Signature),
	}
}

// CopySyncCommitteeContribution copies the provided sync committee contribution object.
func CopySyncCommitteeContribution(c *SyncCommitteeContribution) *SyncCommitteeContribution {
	if c == nil {
		return nil
	}
	return &SyncCommitteeContribution{
		Slot:              c.Slot,
		BlockRoot:         bytesutil.SafeCopyBytes(c.BlockRoot),
		SubcommitteeIndex: c.SubcommitteeIndex,
		AggregationBits:   bytesutil.SafeCopyBytes(c.AggregationBits),
		Signature:         bytesutil.SafeCopyBytes(c.Signature),
	}
}

// CopySignedBeaconBlockGloas copies the provided signed beacon block Gloas object.
func CopySignedBeaconBlockGloas(sb *SignedBeaconBlockGloas) *SignedBeaconBlockGloas {
	if sb == nil {
		return nil
	}
	return &SignedBeaconBlockGloas{
		Block:     copyBeaconBlockGloas(sb.Block),
		Signature: bytesutil.SafeCopyBytes(sb.Signature),
	}
}

// copyBeaconBlockGloas copies the provided beacon block Gloas object.
func copyBeaconBlockGloas(b *BeaconBlockGloas) *BeaconBlockGloas {
	if b == nil {
		return nil
	}
	return &BeaconBlockGloas{
		Slot:          b.Slot,
		ProposerIndex: b.ProposerIndex,
		ParentRoot:    bytesutil.SafeCopyBytes(b.ParentRoot),
		StateRoot:     bytesutil.SafeCopyBytes(b.StateRoot),
		Body:          copyBeaconBlockBodyGloas(b.Body),
	}
}

// copyPayloadAttestation copies the provided payload attestation object.
func copyPayloadAttestation(pa *PayloadAttestation) *PayloadAttestation {
	if pa == nil {
		return nil
	}
	copied := &PayloadAttestation{
		AggregationBits: pa.AggregationBits,
		Signature:       bytesutil.SafeCopyBytes(pa.Signature),
	}
	if pa.Data != nil {
		copied.Data = &PayloadAttestationData{
			BeaconBlockRoot:   bytesutil.SafeCopyBytes(pa.Data.BeaconBlockRoot),
			Slot:              pa.Data.Slot,
			PayloadPresent:    pa.Data.PayloadPresent,
			BlobDataAvailable: pa.Data.BlobDataAvailable,
		}
	}
	return copied
}

// copyPayloadAttestations copies a slice of payload attestations.
func copyPayloadAttestations(pas []*PayloadAttestation) []*PayloadAttestation {
	if len(pas) == 0 {
		return nil
	}
	copied := make([]*PayloadAttestation, len(pas))
	for i, pa := range pas {
		copied[i] = copyPayloadAttestation(pa)
	}
	return copied
}

// copySignedSilaPayloadBid copies the provided signed sila payload header.
func copySignedSilaPayloadBid(header *SignedSilaPayloadBid) *SignedSilaPayloadBid {
	if header == nil {
		return nil
	}
	copied := &SignedSilaPayloadBid{
		Signature: bytesutil.SafeCopyBytes(header.Signature),
	}
	if header.Message != nil {
		copied.Message = &SilaPayloadBid{
			ParentBlockHash:       bytesutil.SafeCopyBytes(header.Message.ParentBlockHash),
			ParentBlockRoot:       bytesutil.SafeCopyBytes(header.Message.ParentBlockRoot),
			BlockHash:             bytesutil.SafeCopyBytes(header.Message.BlockHash),
			PrevRandao:            bytesutil.SafeCopyBytes(header.Message.PrevRandao),
			FeeRecipient:          bytesutil.SafeCopyBytes(header.Message.FeeRecipient),
			GasLimit:              header.Message.GasLimit,
			BuilderIndex:          header.Message.BuilderIndex,
			Slot:                  header.Message.Slot,
			Value:                 header.Message.Value,
			ExecutionPayment:      header.Message.ExecutionPayment,
			BlobKzgCommitments:    bytesutil.SafeCopy2dBytes(header.Message.BlobKzgCommitments),
			SilaRequestsRoot: bytesutil.SafeCopyBytes(header.Message.SilaRequestsRoot),
		}
	}
	return copied
}

// copyBeaconBlockBodyGloas copies the provided beacon block body Gloas object.
func copyBeaconBlockBodyGloas(body *BeaconBlockBodyGloas) *BeaconBlockBodyGloas {
	if body == nil {
		return nil
	}

	copied := &BeaconBlockBodyGloas{
		RandaoReveal: bytesutil.SafeCopyBytes(body.RandaoReveal),
		Graffiti:     bytesutil.SafeCopyBytes(body.Graffiti),
	}

	if body.SilaExecutionData != nil {
		copied.SilaExecutionData = body.SilaExecutionData.Copy()
	}

	if body.SyncAggregate != nil {
		copied.SyncAggregate = body.SyncAggregate.Copy()
	}

	copied.ProposerSlashings = CopySlice(body.ProposerSlashings)
	copied.AttesterSlashings = CopySlice(body.AttesterSlashings)
	copied.Attestations = CopySlice(body.Attestations)
	copied.Deposits = CopySlice(body.Deposits)
	copied.VoluntaryExits = CopySlice(body.VoluntaryExits)
	copied.BlsToSilaChanges = CopySlice(body.BlsToSilaChanges)

	copied.SignedSilaPayloadBid = copySignedSilaPayloadBid(body.SignedSilaPayloadBid)
	copied.PayloadAttestations = copyPayloadAttestations(body.PayloadAttestations)
	copied.ParentSilaRequests = CopySilaRequests(body.ParentSilaRequests)

	return copied
}

// CopySignedSilaPayloadEnvelope copies the provided signed sila payload envelope.
func CopySignedSilaPayloadEnvelope(env *SignedSilaPayloadEnvelope) *SignedSilaPayloadEnvelope {
	if env == nil {
		return nil
	}
	return &SignedSilaPayloadEnvelope{
		Message:   copySilaPayloadEnvelope(env.Message),
		Signature: bytesutil.SafeCopyBytes(env.Signature),
	}
}

// copySilaPayloadEnvelope copies the provided sila payload envelope.
func copySilaPayloadEnvelope(env *SilaPayloadEnvelope) *SilaPayloadEnvelope {
	if env == nil {
		return nil
	}
	return &SilaPayloadEnvelope{
		Payload:               env.Payload, // engine proto, not deep copied here
		SilaRequests:     env.SilaRequests,
		BuilderIndex:          env.BuilderIndex,
		BeaconBlockRoot:       bytesutil.SafeCopyBytes(env.BeaconBlockRoot),
		ParentBeaconBlockRoot: bytesutil.SafeCopyBytes(env.ParentBeaconBlockRoot),
	}
}

// CopySignedBlindedSilaPayloadEnvelope copies the provided signed blinded sila payload envelope.
func CopySignedBlindedSilaPayloadEnvelope(env *SignedBlindedSilaPayloadEnvelope) *SignedBlindedSilaPayloadEnvelope {
	if env == nil {
		return nil
	}
	return &SignedBlindedSilaPayloadEnvelope{
		Message:   copyBlindedSilaPayloadEnvelope(env.Message),
		Signature: bytesutil.SafeCopyBytes(env.Signature),
	}
}

// copyBlindedSilaPayloadEnvelope copies the provided blinded sila payload envelope.
func copyBlindedSilaPayloadEnvelope(env *BlindedSilaPayloadEnvelope) *BlindedSilaPayloadEnvelope {
	if env == nil {
		return nil
	}
	return &BlindedSilaPayloadEnvelope{
		BlockHash:             bytesutil.SafeCopyBytes(env.BlockHash),
		SilaRequests:     env.SilaRequests,
		BuilderIndex:          env.BuilderIndex,
		BeaconBlockRoot:       bytesutil.SafeCopyBytes(env.BeaconBlockRoot),
		Slot:                  env.Slot,
		ParentBlockHash:       bytesutil.SafeCopyBytes(env.ParentBlockHash),
		ParentBeaconBlockRoot: bytesutil.SafeCopyBytes(env.ParentBeaconBlockRoot),
	}
}

// CopyPTCs creates a deep copy of a PTC slot.
func CopyPTCs(slot *PTCs) *PTCs {
	if slot == nil {
		return nil
	}
	indices := make([]primitives.ValidatorIndex, len(slot.ValidatorIndices))
	copy(indices, slot.ValidatorIndices)
	return &PTCs{ValidatorIndices: indices}
}

// CopyPTCWindow creates a deep copy of a PTC window.
func CopyPTCWindow(window []*PTCs) []*PTCs {
	if window == nil {
		return nil
	}
	copied := make([]*PTCs, len(window))
	for i, slot := range window {
		copied[i] = CopyPTCs(slot)
	}
	return copied
}

// CopyBuilderPendingPayment creates a deep copy of a builder pending payment.
func CopyBuilderPendingPayment(original *BuilderPendingPayment) *BuilderPendingPayment {
	if original == nil {
		return nil
	}

	return &BuilderPendingPayment{
		Weight:     original.Weight,
		Withdrawal: CopyBuilderPendingWithdrawal(original.Withdrawal),
	}
}

// CopyBuilderPendingWithdrawal creates a deep copy of a builder pending withdrawal.
func CopyBuilderPendingWithdrawal(original *BuilderPendingWithdrawal) *BuilderPendingWithdrawal {
	if original == nil {
		return nil
	}

	return &BuilderPendingWithdrawal{
		FeeRecipient: bytesutil.SafeCopyBytes(original.FeeRecipient),
		Amount:       original.Amount,
		BuilderIndex: original.BuilderIndex,
	}
}
