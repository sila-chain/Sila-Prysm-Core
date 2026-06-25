package migration

import (
	silapbv1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaapi/v1"
	silapbalpha "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// V1ValidatorToV1Alpha1 converts a v1 validator to v1alpha1.
func V1ValidatorToV1Alpha1(v1Validator *silapbv1.Validator) *silapbalpha.Validator {
	if v1Validator == nil {
		return &silapbalpha.Validator{}
	}
	return &silapbalpha.Validator{
		PublicKey:                  v1Validator.Pubkey,
		WithdrawalCredentials:      v1Validator.WithdrawalCredentials,
		EffectiveBalance:           v1Validator.EffectiveBalance,
		Slashed:                    v1Validator.Slashed,
		ActivationEligibilityEpoch: v1Validator.ActivationEligibilityEpoch,
		ActivationEpoch:            v1Validator.ActivationEpoch,
		ExitEpoch:                  v1Validator.ExitEpoch,
		WithdrawableEpoch:          v1Validator.WithdrawableEpoch,
	}
}
