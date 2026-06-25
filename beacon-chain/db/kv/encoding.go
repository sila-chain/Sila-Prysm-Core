package kv

import (
	"context"
	"errors"
	"reflect"

	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/golang/snappy"
	fastssz "github.com/sila-chain/fastssz"
	"google.golang.org/protobuf/proto"
)

func decode(ctx context.Context, data []byte, dst proto.Message) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.decode")
	defer span.End()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	data, err := snappy.Decode(nil, data)
	if err != nil {
		return err
	}
	if isSSZStorageFormat(dst) {
		return dst.(fastssz.Unmarshaler).UnmarshalSSZ(data)
	}
	return proto.Unmarshal(data, dst)
}

func encode(ctx context.Context, msg proto.Message) ([]byte, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.encode")
	defer span.End()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if msg == nil || reflect.ValueOf(msg).IsNil() {
		return nil, errors.New("cannot encode nil message")
	}
	var enc []byte
	var err error
	if isSSZStorageFormat(msg) {
		enc, err = msg.(fastssz.Marshaler).MarshalSSZ()
		if err != nil {
			return nil, err
		}
	} else {
		enc, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	}
	return snappy.Encode(nil, enc), nil
}

// isSSZStorageFormat returns true if the object type should be saved in SSZ encoded format.
func isSSZStorageFormat(obj any) bool {
	switch obj.(type) {
	case *silapb.BeaconState:
		return true
	case *silapb.SignedBeaconBlock:
		return true
	case *silapb.SignedAggregateAttestationAndProof:
		return true
	case *silapb.BeaconBlock:
		return true
	case *silapb.Attestation, *silapb.AttestationElectra:
		return true
	case *silapb.Deposit:
		return true
	case *silapb.AttesterSlashing, *silapb.AttesterSlashingElectra:
		return true
	case *silapb.ProposerSlashing:
		return true
	case *silapb.VoluntaryExit:
		return true
	case *silapb.ValidatorRegistrationV1:
		return true
	default:
		return false
	}
}
