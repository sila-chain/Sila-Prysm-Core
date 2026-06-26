package helpers

import (
	"bytes"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func ForksEqual(s, t *silapb.Fork) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	if s.Epoch != t.Epoch {
		return false
	}
	if !bytes.Equal(s.PreviousVersion, t.PreviousVersion) {
		return false
	}
	return bytes.Equal(s.CurrentVersion, t.CurrentVersion)
}

func BlockHeadersEqual(s, t *silapb.BeaconBlockHeader) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	if s.Slot != t.Slot {
		return false
	}
	if s.ProposerIndex != t.ProposerIndex {
		return false
	}
	if !bytes.Equal(s.ParentRoot, t.ParentRoot) {
		return false
	}
	if !bytes.Equal(s.StateRoot, t.StateRoot) {
		return false
	}
	return bytes.Equal(s.BodyRoot, t.BodyRoot)
}

func SilaDataEqual(s, t *silapb.SilaData) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	if !bytes.Equal(s.DepositRoot, t.DepositRoot) {
		return false
	}
	if s.DepositCount != t.DepositCount {
		return false
	}
	return bytes.Equal(s.BlockHash, t.BlockHash)
}

func PendingDepositsEqual(s, t *silapb.PendingDeposit) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	if !bytes.Equal(s.PublicKey, t.PublicKey) {
		return false
	}
	if !bytes.Equal(s.WithdrawalCredentials, t.WithdrawalCredentials) {
		return false
	}
	if s.Amount != t.Amount {
		return false
	}
	if !bytes.Equal(s.Signature, t.Signature) {
		return false
	}
	return s.Slot == t.Slot
}

func PendingPartialWithdrawalsEqual(s, t *silapb.PendingPartialWithdrawal) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	if s.Index != t.Index {
		return false
	}
	if s.Amount != t.Amount {
		return false
	}
	return s.WithdrawableEpoch == t.WithdrawableEpoch
}

func PendingConsolidationsEqual(s, t *silapb.PendingConsolidation) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	return s.SourceIndex == t.SourceIndex && s.TargetIndex == t.TargetIndex
}

func BuilderPendingWithdrawalsEqual(s, t *silapb.BuilderPendingWithdrawal) bool {
	if s == nil && t == nil {
		return true
	}
	if s == nil || t == nil {
		return false
	}
	if s.Amount != t.Amount {
		return false
	}
	if s.BuilderIndex != t.BuilderIndex {
		return false
	}
	return bytes.Equal(s.FeeRecipient, t.FeeRecipient)
}
