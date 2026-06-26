package util

import (
	"context"
	"fmt"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	v1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
)

// GenerateFullBlockCapella generates a fully valid Capella block with the requested parameters.
// Use BlockGenConfig to declare the conditions you would like the block generated under.
// This function modifies the passed state as follows:
func GenerateFullBlockCapella(
	bState state.BeaconState,
	privs []bls.SecretKey,
	conf *BlockGenConfig,
	slot primitives.Slot,
) (*silapb.SignedBeaconBlockCapella, error) {
	ctx := context.Background()
	currentSlot := bState.Slot()
	if currentSlot > slot {
		return nil, fmt.Errorf("current slot in state is larger than given slot. %d > %d", currentSlot, slot)
	}
	bState = bState.Copy()

	if conf == nil {
		conf = &BlockGenConfig{}
	}

	var err error
	var pSlashings []*silapb.ProposerSlashing
	numToGen := conf.NumProposerSlashings
	if numToGen > 0 {
		pSlashings, err = generateProposerSlashings(bState, privs, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d proposer slashings:", numToGen)
		}
	}

	numToGen = conf.NumAttesterSlashings
	var aSlashings []*silapb.AttesterSlashing
	if numToGen > 0 {
		generated, err := generateAttesterSlashings(bState, privs, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d attester slashings:", numToGen)
		}
		aSlashings = make([]*silapb.AttesterSlashing, len(generated))
		var ok bool
		for i, s := range generated {
			aSlashings[i], ok = s.(*silapb.AttesterSlashing)
			if !ok {
				return nil, fmt.Errorf("attester slashing has the wrong type (expected %T, got %T)", &silapb.AttesterSlashing{}, s)
			}
		}
	}

	numToGen = conf.NumAttestations
	var atts []*silapb.Attestation
	if numToGen > 0 {
		generatedAtts, err := GenerateAttestations(bState, privs, numToGen, slot, false)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d attestations:", numToGen)
		}
		atts = make([]*silapb.Attestation, len(generatedAtts))
		var ok bool
		for i, a := range generatedAtts {
			atts[i], ok = a.(*silapb.Attestation)
			if !ok {
				return nil, fmt.Errorf("attestation has the wrong type (expected %T, got %T)", &silapb.Attestation{}, a)
			}
		}
	}

	numToGen = conf.NumDeposits
	var newDeposits []*silapb.Deposit
	silaexecData := bState.SilaData()
	if numToGen > 0 {
		newDeposits, silaexecData, err = generateDepositsAndSilaData(bState, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d deposits:", numToGen)
		}
	}

	numToGen = conf.NumVoluntaryExits
	var exits []*silapb.SignedVoluntaryExit
	if numToGen > 0 {
		exits, err = generateVoluntaryExits(bState, privs, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d attester slashings:", numToGen)
		}
	}

	numToGen = conf.NumTransactions
	newTransactions := make([][]byte, numToGen)
	for i := uint64(0); i < numToGen; i++ {
		newTransactions[i] = bytesutil.Uint64ToBytesLittleEndian(i)
	}
	newWithdrawals := make([]*v1.Withdrawal, 0)

	random, err := helpers.RandaoMix(bState, time.CurrentEpoch(bState))
	if err != nil {
		return nil, errors.Wrap(err, "could not process randao mix")
	}

	timestamp, err := slots.StartTime(bState.GenesisTime(), slot)
	if err != nil {
		return nil, errors.Wrap(err, "could not get current timestamp")
	}

	stCopy := bState.Copy()
	stCopy, err = transition.ProcessSlots(context.Background(), stCopy, slot)
	if err != nil {
		return nil, err
	}

	parentExecution, err := stCopy.LatestSilaPayloadHeader()
	if err != nil {
		return nil, err
	}
	blockHash := indexToHash(uint64(slot))
	newSilaPayloadCapella := &v1.SilaPayloadCapella{
		ParentHash:    parentExecution.BlockHash(),
		FeeRecipient:  make([]byte, 20),
		StateRoot:     params.BeaconConfig().ZeroHash[:],
		ReceiptsRoot:  params.BeaconConfig().ZeroHash[:],
		LogsBloom:     make([]byte, 256),
		PrevRandao:    random,
		BlockNumber:   uint64(slot),
		ExtraData:     params.BeaconConfig().ZeroHash[:],
		BaseFeePerGas: params.BeaconConfig().ZeroHash[:],
		BlockHash:     blockHash[:],
		Timestamp:     uint64(timestamp.Unix()),
		Transactions:  newTransactions,
		Withdrawals:   newWithdrawals,
	}
	var syncCommitteeBits []byte
	currSize := new(silapb.SyncAggregate).SyncCommitteeBits.Len()
	switch currSize {
	case 512:
		syncCommitteeBits = bitfield.NewBitvector512()
	case 32:
		syncCommitteeBits = bitfield.NewBitvector32()
	default:
		return nil, errors.New("invalid bit vector size")
	}
	newSyncAggregate := &silapb.SyncAggregate{
		SyncCommitteeBits:      syncCommitteeBits,
		SyncCommitteeSignature: append([]byte{0xC0}, make([]byte, 95)...),
	}

	newHeader := bState.LatestBlockHeader()
	prevStateRoot, err := bState.HashTreeRoot(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not hash state")
	}
	newHeader.StateRoot = prevStateRoot[:]
	parentRoot, err := newHeader.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not hash the new header")
	}

	if slot == currentSlot {
		slot = currentSlot + 1
	}

	reveal, err := RandaoReveal(stCopy, time.CurrentEpoch(stCopy), privs)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute randao reveal")
	}

	idx, err := helpers.BeaconProposerIndex(ctx, stCopy)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute beacon proposer index")
	}

	changes := make([]*silapb.SignedBLSToSilaChange, conf.NumBLSChanges)
	for i := uint64(0); i < conf.NumBLSChanges; i++ {
		changes[i], err = GenerateBLSToSilaChange(bState, privs[i+1], primitives.ValidatorIndex(i))
		if err != nil {
			return nil, err
		}
	}

	block := &silapb.BeaconBlockCapella{
		Slot:          slot,
		ParentRoot:    parentRoot[:],
		ProposerIndex: idx,
		Body: &silapb.BeaconBlockBodyCapella{
			SilaData:              silaexecData,
			RandaoReveal:          reveal,
			ProposerSlashings:     pSlashings,
			AttesterSlashings:     aSlashings,
			Attestations:          atts,
			VoluntaryExits:        exits,
			Deposits:              newDeposits,
			Graffiti:              make([]byte, fieldparams.RootLength),
			SyncAggregate:         newSyncAggregate,
			SilaPayload:      newSilaPayloadCapella,
			BlsToSilaChanges: changes,
		},
	}

	// The fork can change after processing the state
	signature, err := BlockSignature(bState, block, privs)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute block signature")
	}

	return &silapb.SignedBeaconBlockCapella{Block: block, Signature: signature.Marshal()}, nil
}

// GenerateBLSToSilaChange generates a valid bls to exec change for validator `val` and its private key `priv` with the given beacon state `st`.
func GenerateBLSToSilaChange(st state.BeaconState, priv bls.SecretKey, val primitives.ValidatorIndex) (*silapb.SignedBLSToSilaChange, error) {
	cred := indexToHash(uint64(val))
	pubkey := priv.PublicKey().Marshal()
	message := &silapb.BLSToSilaChange{
		ToSilaAddress: cred[12:],
		ValidatorIndex:     val,
		FromBlsPubkey:      pubkey,
	}
	c := params.BeaconConfig()
	domain, err := signing.ComputeDomain(c.DomainBLSToSilaChange, c.GenesisForkVersion, st.GenesisValidatorsRoot())
	if err != nil {
		return nil, err
	}
	sr, err := signing.ComputeSigningRoot(message, domain)
	if err != nil {
		return nil, err
	}
	signature := priv.Sign(sr[:]).Marshal()
	return &silapb.SignedBLSToSilaChange{
		Message:   message,
		Signature: signature,
	}, nil
}
