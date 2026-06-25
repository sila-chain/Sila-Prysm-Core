package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	validatorpb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/validator-client"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SubmitPayloadAttestation submits a payload attestation message for a PTC member.
func (v *validator) SubmitPayloadAttestation(ctx context.Context, slot primitives.Slot, pubKey [fieldparams.BLSPubkeyLength]byte) {
	ctx, span := trace.StartSpan(ctx, "validator.SubmitPayloadAttestation")
	defer span.End()
	span.SetAttributes(trace.StringAttribute("validator", fmt.Sprintf("%#x", pubKey)))

	if slots.ToEpoch(slot) < params.BeaconConfig().GloasForkEpoch {
		return
	}

	v.waitForPayloadAvailableOrDeadline(ctx, slot)

	data, err := v.validatorClient.PayloadAttestationData(ctx, slot)
	if err != nil {
		if status.Code(errors.Cause(err)) == codes.Unavailable {
			validatorPayloadAttestationSubmissionTotal.WithLabelValues("skipped_unavailable").Inc()
			log.WithFields(logrus.Fields{
				"slot":   slot,
				"reason": status.Convert(errors.Cause(err)).Message(),
			}).Info("Skipping payload attestation: data unavailable")
			tracing.AnnotateError(span, err)
			return
		}
		validatorPayloadAttestationSubmissionTotal.WithLabelValues("failed").Inc()
		log.WithError(err).Error("Could not request payload attestation data")
		tracing.AnnotateError(span, err)
		return
	}

	d, err := v.domainData(ctx, slots.ToEpoch(slot), params.BeaconConfig().DomainPTCAttester[:])
	if err != nil {
		validatorPayloadAttestationSubmissionTotal.WithLabelValues("failed").Inc()
		log.WithError(err).Error("Could not get PTC attester domain data")
		return
	}

	r, err := signing.ComputeSigningRoot(data, d.SignatureDomain)
	if err != nil {
		validatorPayloadAttestationSubmissionTotal.WithLabelValues("failed").Inc()
		log.WithError(err).Error("Could not compute payload attestation signing root")
		return
	}

	sig, err := v.km.Sign(ctx, &validatorpb.SignRequest{
		PublicKey:       pubKey[:],
		SigningRoot:     r[:],
		SignatureDomain: d.SignatureDomain,
		Object: &validatorpb.SignRequest_PayloadAttestationData{
			PayloadAttestationData: data,
		},
		SigningSlot: slot,
	})
	if err != nil {
		validatorPayloadAttestationSubmissionTotal.WithLabelValues("failed").Inc()
		log.WithError(err).Error("Could not sign payload attestation")
		return
	}

	duty, err := v.duty(pubKey)
	if err != nil {
		validatorPayloadAttestationSubmissionTotal.WithLabelValues("failed").Inc()
		log.WithError(err).Error("Could not fetch validator assignment")
		return
	}

	msg := &silapb.PayloadAttestationMessage{
		ValidatorIndex: duty.ValidatorIndex,
		Data:           data,
		Signature:      sig.Marshal(),
	}
	if _, err := v.validatorClient.SubmitPayloadAttestation(ctx, msg); err != nil {
		validatorPayloadAttestationSubmissionTotal.WithLabelValues("failed").Inc()
		log.WithError(err).Error("Could not submit payload attestation")
		return
	}
	validatorPayloadAttestationSubmissionTotal.WithLabelValues("success").Inc()

	slotTime, err := slots.StartTime(v.genesisTime, slot)
	if err != nil {
		log.WithError(err).Error("Failed to determine slot start time")
	}
	log.WithFields(logrus.Fields{
		"slot":               slot,
		"slotStartTime":      slotTime,
		"timeSinceSlotStart": time.Since(slotTime),
		"blockRoot":          fmt.Sprintf("%#x", bytesutil.Trunc(data.BeaconBlockRoot)),
		"payloadPresent":     data.PayloadPresent,
		"blobDataAvailable":  data.BlobDataAvailable,
		"validatorIndex":     duty.ValidatorIndex,
	}).Info("Submitted new payload attestation")
}
