package altair

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
)

// ProcessPreGenesisDeposits processes a deposit for the beacon state before chainstart.
func ProcessPreGenesisDeposits(
	ctx context.Context,
	beaconState state.BeaconState,
	deposits []*silapb.Deposit,
) (state.BeaconState, error) {
	var err error
	beaconState, err = ProcessDeposits(ctx, beaconState, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process deposit")
	}
	beaconState, err = helpers.ActivateValidatorWithEffectiveBalance(beaconState, deposits)
	if err != nil {
		return nil, err
	}
	return beaconState, nil
}

// ProcessDeposits processes validator deposits for beacon state Altair.
func ProcessDeposits(
	ctx context.Context,
	beaconState state.BeaconState,
	deposits []*silapb.Deposit,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "altair.ProcessDeposits")
	defer span.End()

	span.SetAttributes(trace.Int64Attribute("count", int64(len(deposits))))

	allSignaturesVerified, err := helpers.BatchVerifyDepositsSignatures(ctx, deposits)
	if err != nil {
		return nil, err
	}

	for _, deposit := range deposits {
		if deposit == nil || deposit.Data == nil {
			return nil, errors.New("got a nil deposit in block")
		}
		beaconState, err = ProcessDeposit(beaconState, deposit, allSignaturesVerified)
		if err != nil {
			return nil, errors.Wrapf(err, "could not process deposit from %#x", bytesutil.Trunc(deposit.Data.PublicKey))
		}
	}
	return beaconState, nil
}

// ProcessDeposit takes in a deposit object and inserts it
// into the registry as a new validator or balance change.
// Returns the resulting state, a boolean to indicate whether or not the deposit
// resulted in a new validator entry into the beacon state, and any error.
//
// Spec pseudocode definition:
// def process_deposit(state: BeaconState, deposit: Deposit) -> None:
//
//	# Verify the Merkle branch
//	assert is_valid_merkle_branch(
//		leaf=hash_tree_root(deposit.data),
//		branch=deposit.proof,
//		depth=SILA_DEPOSIT_TREE_DEPTH + 1,  # Add 1 for the List length mix-in
//		index=state.silaexec_deposit_index,
//		root=state.sila_data.deposit_root,
//	)
//
//	 # Deposits must be processed in order
//	 state.silaexec_deposit_index += 1
//
//	 apply_deposit(
//	  state=state,
//	  pubkey=deposit.data.pubkey,
//	  withdrawal_credentials=deposit.data.withdrawal_credentials,
//	  amount=deposit.data.amount,
//	  signature=deposit.data.signature,
//	 )
func ProcessDeposit(beaconState state.BeaconState, deposit *silapb.Deposit, allSignaturesVerified bool) (state.BeaconState, error) {
	if err := helpers.VerifyDeposit(beaconState, deposit); err != nil {
		if deposit == nil || deposit.Data == nil {
			return nil, err
		}
		return nil, errors.Wrapf(err, "could not verify deposit from %#x", bytesutil.Trunc(deposit.Data.PublicKey))
	}
	if err := beaconState.SetSilaExecutionDepositIndex(beaconState.SilaExecutionDepositIndex() + 1); err != nil {
		return nil, err
	}

	return ApplyDeposit(beaconState, deposit.Data, allSignaturesVerified)
}

// ApplyDeposit
// Spec pseudocode definition:
// def apply_deposit(state: BeaconState, pubkey: BLSPubkey, withdrawal_credentials: Bytes32, amount: uint64, signature: BLSSignature) -> None:
//
//	validator_pubkeys = [v.pubkey for v in state.validators]
//	if pubkey not in validator_pubkeys:
//	    # Verify the deposit signature (proof of possession) which is not checked by the sila deposit
//	    deposit_message = DepositMessage(
//	        pubkey=pubkey,
//	        withdrawal_credentials=withdrawal_credentials,
//	        amount=amount,
//	    )
//	    domain = compute_domain(DOMAIN_DEPOSIT)  # Fork-agnostic domain since deposits are valid across forks
//	    signing_root = compute_signing_root(deposit_message, domain)
//	    if bls.Verify(pubkey, signing_root, signature):
//	        add_validator_to_registry(state, pubkey, withdrawal_credentials, amount)
//	else:
//	    # Increase balance by deposit amount
//	    index = ValidatorIndex(validator_pubkeys.index(pubkey))
//	    increase_balance(state, index, amount)
func ApplyDeposit(beaconState state.BeaconState, data *silapb.Deposit_Data, allSignaturesVerified bool) (state.BeaconState, error) {
	pubKey := data.PublicKey
	amount := data.Amount
	withdrawalCredentials := data.WithdrawalCredentials
	index, ok := beaconState.ValidatorIndexByPubkey(bytesutil.ToBytes48(pubKey))
	if !ok {
		if !allSignaturesVerified {
			valid, err := helpers.IsValidDepositSignature(data)
			if err != nil {
				return nil, err
			}
			if !valid {
				return beaconState, nil
			}
		}
		if err := AddValidatorToRegistry(beaconState, pubKey, withdrawalCredentials, amount); err != nil {
			return nil, err
		}
	} else {
		if err := helpers.IncreaseBalance(beaconState, index, amount); err != nil {
			return nil, err
		}
	}
	return beaconState, nil
}

// AddValidatorToRegistry updates the beacon state with validator information
// def add_validator_to_registry(state: BeaconState,
//
//	                          pubkey: BLSPubkey,
//	                          withdrawal_credentials: Bytes32,
//	                          amount: uint64) -> None:
//	index = get_index_for_new_validator(state)
//	validator = get_validator_from_deposit(pubkey, withdrawal_credentials)
//	set_or_append_list(state.validators, index, validator)
//	set_or_append_list(state.balances, index, 0)
//	set_or_append_list(state.previous_epoch_participation, index, ParticipationFlags(0b0000_0000)) // New in Altair
//	set_or_append_list(state.current_epoch_participation, index, ParticipationFlags(0b0000_0000)) // New in Altair
//	set_or_append_list(state.inactivity_scores, index, uint64(0)) // New in Altair
func AddValidatorToRegistry(beaconState state.BeaconState, pubKey []byte, withdrawalCredentials []byte, amount uint64) error {
	val := GetValidatorFromDeposit(pubKey, withdrawalCredentials, amount)
	if err := beaconState.AppendValidator(val); err != nil {
		return err
	}
	if err := beaconState.AppendBalance(amount); err != nil {
		return err
	}

	// only active in altair and only when it's a new validator (after append balance)
	if beaconState.Version() >= version.Altair {
		if err := beaconState.AppendInactivityScore(0); err != nil {
			return err
		}
		if err := beaconState.AppendPreviousParticipationBits(0); err != nil {
			return err
		}
		if err := beaconState.AppendCurrentParticipationBits(0); err != nil {
			return err
		}
	}
	return nil
}

// GetValidatorFromDeposit gets a new validator object with provided parameters
//
// def get_validator_from_deposit(pubkey: BLSPubkey, withdrawal_credentials: Bytes32, amount: uint64) -> Validator:
//
//	effective_balance = min(amount - amount % EFFECTIVE_BALANCE_INCREMENT, MAX_EFFECTIVE_BALANCE)
//
//	return Validator(
//	    pubkey=pubkey,
//	    withdrawal_credentials=withdrawal_credentials,
//	    effective_balance=effective_balance,
//	    slashed=False,
//	    activation_eligibility_epoch=FAR_FUTURE_EPOCH,
//	    activation_epoch=FAR_FUTURE_EPOCH,
//	    exit_epoch=FAR_FUTURE_EPOCH,
//	    withdrawable_epoch=FAR_FUTURE_EPOCH,
//	)
func GetValidatorFromDeposit(pubKey []byte, withdrawalCredentials []byte, amount uint64) *silapb.Validator {
	effectiveBalance := min(params.BeaconConfig().MaxEffectiveBalance, amount-(amount%params.BeaconConfig().EffectiveBalanceIncrement))

	return &silapb.Validator{
		PublicKey:                  pubKey,
		WithdrawalCredentials:      withdrawalCredentials,
		EffectiveBalance:           effectiveBalance,
		Slashed:                    false,
		ActivationEligibilityEpoch: params.BeaconConfig().FarFutureEpoch,
		ActivationEpoch:            params.BeaconConfig().FarFutureEpoch,
		ExitEpoch:                  params.BeaconConfig().FarFutureEpoch,
		WithdrawableEpoch:          params.BeaconConfig().FarFutureEpoch,
	}
}
