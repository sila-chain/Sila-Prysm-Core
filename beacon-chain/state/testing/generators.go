package testing

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls/common"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

// GeneratePendingDeposit is used for testing and producing a signed pending deposit
func GeneratePendingDeposit(t *testing.T, key common.SecretKey, amount uint64, withdrawalCredentials [32]byte, slot primitives.Slot) *silapb.PendingDeposit {
	dm := &silapb.DepositMessage{
		PublicKey:             key.PublicKey().Marshal(),
		WithdrawalCredentials: withdrawalCredentials[:],
		Amount:                amount,
	}
	domain, err := signing.ComputeDomain(params.BeaconConfig().DomainDeposit, nil, nil)
	require.NoError(t, err)
	sr, err := signing.ComputeSigningRoot(dm, domain)
	require.NoError(t, err)
	sig := key.Sign(sr[:])
	depositData := &silapb.Deposit_Data{
		PublicKey:             bytesutil.SafeCopyBytes(dm.PublicKey),
		WithdrawalCredentials: bytesutil.SafeCopyBytes(dm.WithdrawalCredentials),
		Amount:                dm.Amount,
		Signature:             sig.Marshal(),
	}
	valid, err := helpers.IsValidDepositSignature(depositData)
	require.NoError(t, err)
	require.Equal(t, true, valid)
	return &silapb.PendingDeposit{
		PublicKey:             bytesutil.SafeCopyBytes(dm.PublicKey),
		WithdrawalCredentials: bytesutil.SafeCopyBytes(dm.WithdrawalCredentials),
		Amount:                dm.Amount,
		Signature:             sig.Marshal(),
		Slot:                  slot,
	}
}
