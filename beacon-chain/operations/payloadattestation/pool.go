package payloadattestation

import (
	"sync"

	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

var errNilPayloadAttestationMessage = errors.New("nil payload attestation message")

type payloadAttestationDataKey struct {
	beaconBlockRoot   [32]byte
	slot              primitives.Slot
	payloadPresent    bool
	blobDataAvailable bool
}

// PoolManager manages pending, aggregated payload attestations keyed by
// payload-attestation data.
type PoolManager interface {
	// PendingPayloadAttestations returns pending attestations for the requested slot.
	PendingPayloadAttestations(slot primitives.Slot) []*silapb.PayloadAttestation
	// InsertPayloadAttestation inserts or aggregates a payload attestation
	// message into the pool. The idx parameter is the PTC committee index
	// of the validator (position in the bitvector).
	InsertPayloadAttestation(msg *silapb.PayloadAttestationMessage, idx uint64) error
	// Seen returns true if the PTC committee index has already been seen
	// for the given PayloadAttestationData.
	Seen(data *silapb.PayloadAttestationData, idx uint64) bool
}

// Pool is an in-memory implementation of PoolManager.
// Entries are keyed by payload-attestation data fields and store an aggregated
// PayloadAttestation value.
type Pool struct {
	lock    sync.RWMutex
	pending map[payloadAttestationDataKey]*silapb.PayloadAttestation
}

// NewPool returns an initialized pool.
func NewPool() *Pool {
	pool := &Pool{
		pending: make(map[payloadAttestationDataKey]*silapb.PayloadAttestation),
	}
	payloadAttestationPoolSize.Set(0)
	return pool
}

// PendingPayloadAttestations returns payload attestations for the requested slot.
func (p *Pool) PendingPayloadAttestations(slot primitives.Slot) []*silapb.PayloadAttestation {
	p.lock.Lock()
	defer p.lock.Unlock()

	result := make([]*silapb.PayloadAttestation, 0, len(p.pending))
	for _, att := range p.pending {
		if att.Data.Slot == slot {
			result = append(result, att)
		}
	}
	return result
}

// InsertPayloadAttestation inserts a payload attestation message into the pool.
// If an attestation with matching data already exists, it aggregates the BLS
// signature and sets the aggregation bit for idx.
// idx is the validator's position in the PTC committee bitfield. It also prunes
// stale entries with slot lower than msg.Data.Slot.
func (p *Pool) InsertPayloadAttestation(msg *silapb.PayloadAttestationMessage, idx uint64) error {
	if msg == nil || msg.Data == nil {
		return errNilPayloadAttestationMessage
	}
	if idx >= uint64(fieldparams.PTCSize) {
		return errors.Errorf("invalid payload attestation committee index: %d", idx)
	}

	key, err := dataKey(msg.Data)
	if err != nil {
		return errors.Wrap(err, "could not compute data key")
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.pruneOlderSlotsLocked(msg.Data.Slot)

	existing, ok := p.pending[key]
	if !ok {
		p.pending[key] = messageToPayloadAttestation(msg, idx)
		payloadAttestationPoolSize.Set(float64(len(p.pending)))
		return nil
	}

	if existing.AggregationBits.BitAt(idx) {
		return nil
	}

	sig, err := aggregateSigFromMessage(existing, msg)
	if err != nil {
		return errors.Wrap(err, "could not aggregate signatures")
	}
	existing.Signature = sig
	existing.AggregationBits.SetBitAt(idx, true)
	payloadAttestationPoolSize.Set(float64(len(p.pending)))
	return nil
}

func (p *Pool) pruneOlderSlotsLocked(slot primitives.Slot) {
	for key, att := range p.pending {
		if att == nil || att.Data == nil || att.Data.Slot < slot {
			delete(p.pending, key)
		}
	}
	payloadAttestationPoolSize.Set(float64(len(p.pending)))
}

// Seen reports whether idx has already been observed for the given
// PayloadAttestationData.
func (p *Pool) Seen(data *silapb.PayloadAttestationData, idx uint64) bool {
	if data == nil {
		return false
	}

	key, err := dataKey(data)
	if err != nil {
		return false
	}

	p.lock.RLock()
	defer p.lock.RUnlock()

	existing, ok := p.pending[key]
	if !ok {
		return false
	}
	return existing.AggregationBits.BitAt(idx)
}

// messageToPayloadAttestation creates an aggregated PayloadAttestation with a
// single bit set at idx from msg.
func messageToPayloadAttestation(msg *silapb.PayloadAttestationMessage, idx uint64) *silapb.PayloadAttestation {
	bits := silapb.NewPayloadAttestationAggregationBits()
	bits.SetBitAt(idx, true)
	data := &silapb.PayloadAttestationData{
		BeaconBlockRoot:   bytesutil.SafeCopyBytes(msg.Data.BeaconBlockRoot),
		Slot:              msg.Data.Slot,
		PayloadPresent:    msg.Data.PayloadPresent,
		BlobDataAvailable: msg.Data.BlobDataAvailable,
	}
	return &silapb.PayloadAttestation{
		AggregationBits: bits,
		Data:            data,
		Signature:       bytesutil.SafeCopyBytes(msg.Signature),
	}
}

// aggregateSigFromMessage aggregates the existing signature with the new
// message signature.
func aggregateSigFromMessage(aggregated *silapb.PayloadAttestation, message *silapb.PayloadAttestationMessage) ([]byte, error) {
	aggSig, err := bls.SignatureFromBytesNoValidation(aggregated.Signature)
	if err != nil {
		return nil, err
	}
	sig, err := bls.SignatureFromBytesNoValidation(message.Signature)
	if err != nil {
		return nil, err
	}
	return bls.AggregateSignatures([]bls.Signature{aggSig, sig}).Marshal(), nil
}

// dataKey derives the map key directly from PayloadAttestationData fields.
// BeaconBlockRoot must be 32 bytes.
func dataKey(data *silapb.PayloadAttestationData) (payloadAttestationDataKey, error) {
	if data == nil {
		return payloadAttestationDataKey{}, errNilPayloadAttestationMessage
	}
	if len(data.BeaconBlockRoot) != fieldparams.RootLength {
		return payloadAttestationDataKey{}, errors.Errorf("invalid beacon block root length: %d", len(data.BeaconBlockRoot))
	}
	return payloadAttestationDataKey{
		beaconBlockRoot:   bytesutil.ToBytes32(data.BeaconBlockRoot),
		slot:              data.Slot,
		payloadPresent:    data.PayloadPresent,
		blobDataAvailable: data.BlobDataAvailable,
	}, nil
}
