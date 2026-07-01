package v1alpha1

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
)

// Copy creates a deep copy of SilaPayloadBid.
func (header *SilaPayloadBid) Copy() *SilaPayloadBid {
	if header == nil {
		return nil
	}
	return &SilaPayloadBid{
		ParentBlockHash:    bytesutil.SafeCopyBytes(header.ParentBlockHash),
		ParentBlockRoot:    bytesutil.SafeCopyBytes(header.ParentBlockRoot),
		BlockHash:          bytesutil.SafeCopyBytes(header.BlockHash),
		PrevRandao:         bytesutil.SafeCopyBytes(header.PrevRandao),
		FeeRecipient:       bytesutil.SafeCopyBytes(header.FeeRecipient),
		GasLimit:           header.GasLimit,
		BuilderIndex:       header.BuilderIndex,
		Slot:               header.Slot,
		Value:              header.Value,
		SilaPayment:        header.SilaPayment,
		BlobKzgCommitments: bytesutil.SafeCopy2dBytes(header.BlobKzgCommitments),
		SilaRequestsRoot:   bytesutil.SafeCopyBytes(header.SilaRequestsRoot),
	}
}

// Copy creates a deep copy of BuilderPendingWithdrawal.
func (withdrawal *BuilderPendingWithdrawal) Copy() *BuilderPendingWithdrawal {
	if withdrawal == nil {
		return nil
	}
	return &BuilderPendingWithdrawal{
		FeeRecipient: bytesutil.SafeCopyBytes(withdrawal.FeeRecipient),
		Amount:       withdrawal.Amount,
		BuilderIndex: withdrawal.BuilderIndex,
	}
}

// Copy creates a deep copy of BuilderPendingPayment.
func (payment *BuilderPendingPayment) Copy() *BuilderPendingPayment {
	if payment == nil {
		return nil
	}
	return &BuilderPendingPayment{
		Weight:     payment.Weight,
		Withdrawal: payment.Withdrawal.Copy(),
	}
}
