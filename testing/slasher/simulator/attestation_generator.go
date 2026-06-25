package simulator

import (
	"bytes"
	"context"
	"math"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/rand"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Simulator) generateAttestationsForSlot(ctx context.Context, ver int, slot primitives.Slot) ([]silapb.IndexedAtt, []silapb.AttSlashing, error) {
	attestations := make([]silapb.IndexedAtt, 0)
	slashings := make([]silapb.AttSlashing, 0)
	currentEpoch := slots.ToEpoch(slot)

	committeesPerSlot := helpers.SlotCommitteeCount(s.srvConfig.Params.NumValidators)
	valsPerCommittee := s.srvConfig.Params.NumValidators /
		(committeesPerSlot * uint64(s.srvConfig.Params.SlotsPerEpoch))
	valsPerSlot := committeesPerSlot * valsPerCommittee

	if currentEpoch < 2 {
		return nil, nil, nil
	}
	sourceEpoch := currentEpoch - 1

	var slashedIndices []uint64
	startIdx := valsPerSlot * uint64(slot%s.srvConfig.Params.SlotsPerEpoch)
	endIdx := startIdx + valsPerCommittee
	for c := primitives.CommitteeIndex(0); uint64(c) < committeesPerSlot; c++ {
		attData := &silapb.AttestationData{
			Slot:            slot,
			CommitteeIndex:  c,
			BeaconBlockRoot: bytesutil.PadTo([]byte("block"), 32),
			Source: &silapb.Checkpoint{
				Epoch: sourceEpoch,
				Root:  bytesutil.PadTo([]byte("source"), 32),
			},
			Target: &silapb.Checkpoint{
				Epoch: currentEpoch,
				Root:  bytesutil.PadTo([]byte("target"), 32),
			},
		}

		valsPerAttestation := uint64(math.Floor(s.srvConfig.Params.AggregationPercent * float64(valsPerCommittee)))
		for i := startIdx; i < endIdx; i += valsPerAttestation {
			attEndIdx := min(i+valsPerAttestation, endIdx)
			indices := make([]uint64, 0, valsPerAttestation)
			for idx := i; idx < attEndIdx; idx++ {
				indices = append(indices, idx)
			}

			var att silapb.IndexedAtt
			if ver >= version.Electra {
				att = &silapb.IndexedAttestationElectra{
					AttestingIndices: indices,
					Data:             attData,
					Signature:        params.BeaconConfig().EmptySignature[:],
				}
			} else {
				att = &silapb.IndexedAttestation{
					AttestingIndices: indices,
					Data:             attData,
					Signature:        params.BeaconConfig().EmptySignature[:],
				}
			}

			beaconState, err := s.srvConfig.AttestationStateFetcher.AttestationTargetState(ctx, att.GetData().Target)
			if err != nil {
				return nil, nil, err
			}

			// Sign the attestation with a valid signature.
			aggSig, err := s.aggregateSigForAttestation(beaconState, att)
			if err != nil {
				return nil, nil, err
			}

			if ver >= version.Electra {
				att.(*silapb.IndexedAttestationElectra).Signature = aggSig.Marshal()
			} else {
				att.(*silapb.IndexedAttestation).Signature = aggSig.Marshal()
			}

			attestations = append(attestations, att)
			if rand.NewGenerator().Float64() < s.srvConfig.Params.AttesterSlashingProbab {
				slashableAtt := makeSlashableFromAtt(att, []uint64{indices[0]})
				aggSig, err := s.aggregateSigForAttestation(beaconState, slashableAtt)
				if err != nil {
					return nil, nil, err
				}

				if ver >= version.Electra {
					slashableAtt.(*silapb.IndexedAttestationElectra).Signature = aggSig.Marshal()
				} else {
					slashableAtt.(*silapb.IndexedAttestation).Signature = aggSig.Marshal()
				}

				slashedIndices = append(slashedIndices, slashableAtt.GetAttestingIndices()...)

				attDataRoot, err := att.GetData().HashTreeRoot()
				if err != nil {
					return nil, nil, errors.Wrap(err, "cannot compte `att` hash tree root")
				}

				slashableAttDataRoot, err := slashableAtt.GetData().HashTreeRoot()
				if err != nil {
					return nil, nil, errors.Wrap(err, "cannot compte `slashableAtt` hash tree root")
				}

				var slashing silapb.AttSlashing
				if ver >= version.Electra {
					slashing = &silapb.AttesterSlashingElectra{
						Attestation_1: att.(*silapb.IndexedAttestationElectra),
						Attestation_2: slashableAtt.(*silapb.IndexedAttestationElectra),
					}
				} else {
					slashing = &silapb.AttesterSlashing{
						Attestation_1: att.(*silapb.IndexedAttestation),
						Attestation_2: slashableAtt.(*silapb.IndexedAttestation),
					}
				}

				// Ensure the attestation with the lower data root is the first attestation.
				if bytes.Compare(attDataRoot[:], slashableAttDataRoot[:]) > 0 {
					if ver >= version.Electra {
						slashing = &silapb.AttesterSlashingElectra{
							Attestation_1: slashableAtt.(*silapb.IndexedAttestationElectra),
							Attestation_2: att.(*silapb.IndexedAttestationElectra),
						}
					} else {
						slashing = &silapb.AttesterSlashing{
							Attestation_1: slashableAtt.(*silapb.IndexedAttestation),
							Attestation_2: att.(*silapb.IndexedAttestation),
						}
					}
				}

				slashings = append(slashings, slashing)
				attestations = append(attestations, slashableAtt)
			}
		}
		startIdx += valsPerCommittee
		endIdx += valsPerCommittee
	}
	if len(slashedIndices) > 0 {
		log.WithFields(logrus.Fields{
			"amount":  len(slashedIndices),
			"indices": slashedIndices,
		}).Infof("Slashable attestation made")
	}
	return attestations, slashings, nil
}

func (s *Simulator) aggregateSigForAttestation(
	beaconState state.ReadOnlyBeaconState, att silapb.IndexedAtt,
) (bls.Signature, error) {
	domain, err := signing.Domain(
		beaconState.Fork(),
		att.GetData().Target.Epoch,
		params.BeaconConfig().DomainBeaconAttester,
		beaconState.GenesisValidatorsRoot(),
	)
	if err != nil {
		return nil, err
	}
	signingRoot, err := signing.ComputeSigningRoot(att.GetData(), domain)
	if err != nil {
		return nil, err
	}
	sigs := make([]bls.Signature, len(att.GetAttestingIndices()))
	for i, validatorIndex := range att.GetAttestingIndices() {
		privKey := s.srvConfig.PrivateKeysByValidatorIndex[primitives.ValidatorIndex(validatorIndex)]
		sigs[i] = privKey.Sign(signingRoot[:])
	}
	return bls.AggregateSignatures(sigs), nil
}

func makeSlashableFromAtt(att silapb.IndexedAtt, indices []uint64) silapb.IndexedAtt {
	if att.GetData().Source.Epoch <= 2 {
		return makeDoubleVoteFromAtt(att, indices)
	}
	attData := &silapb.AttestationData{
		Slot:            att.GetData().Slot,
		CommitteeIndex:  att.GetData().CommitteeIndex,
		BeaconBlockRoot: att.GetData().BeaconBlockRoot,
		Source: &silapb.Checkpoint{
			Epoch: att.GetData().Source.Epoch - 3,
			Root:  att.GetData().Source.Root,
		},
		Target: &silapb.Checkpoint{
			Epoch: att.GetData().Target.Epoch,
			Root:  att.GetData().Target.Root,
		},
	}

	if att.Version() >= version.Electra {
		return &silapb.IndexedAttestationElectra{
			AttestingIndices: indices,
			Data:             attData,
			Signature:        params.BeaconConfig().EmptySignature[:],
		}
	}

	return &silapb.IndexedAttestation{
		AttestingIndices: indices,
		Data:             attData,
		Signature:        params.BeaconConfig().EmptySignature[:],
	}
}

func makeDoubleVoteFromAtt(att silapb.IndexedAtt, indices []uint64) silapb.IndexedAtt {
	attData := &silapb.AttestationData{
		Slot:            att.GetData().Slot,
		CommitteeIndex:  att.GetData().CommitteeIndex,
		BeaconBlockRoot: bytesutil.PadTo([]byte("slash me"), 32),
		Source: &silapb.Checkpoint{
			Epoch: att.GetData().Source.Epoch,
			Root:  att.GetData().Source.Root,
		},
		Target: &silapb.Checkpoint{
			Epoch: att.GetData().Target.Epoch,
			Root:  att.GetData().Target.Root,
		},
	}

	if att.Version() >= version.Electra {
		return &silapb.IndexedAttestationElectra{
			AttestingIndices: indices,
			Data:             attData,
			Signature:        params.BeaconConfig().EmptySignature[:],
		}
	}

	return &silapb.IndexedAttestation{
		AttestingIndices: indices,
		Data:             attData,
		Signature:        params.BeaconConfig().EmptySignature[:],
	}
}
