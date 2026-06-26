package state_native

import (
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native/types"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	_ "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
)

// SetLatestSilaPayloadHeader for the beacon state.
func (b *BeaconState) SetLatestSilaPayloadHeader(val interfaces.SilaData) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version < version.Bellatrix {
		return errNotSupported("SetLatestSilaPayloadHeader", b.version)
	}

	switch header := val.Proto().(type) {
	case *silaenginev1.SilaPayload:
		if b.version != version.Bellatrix {
			return fmt.Errorf("wrong state version (%s) for bellatrix sila payload", version.String(b.version))
		}
		latest, err := consensusblocks.PayloadToHeader(val)
		if err != nil {
			return errors.Wrap(err, "could not convert payload to header")
		}
		b.latestSilaPayloadHeader = latest
		b.markFieldAsDirty(types.LatestSilaPayloadHeader)
		return nil
	case *silaenginev1.SilaPayloadCapella:
		if b.version != version.Capella {
			return fmt.Errorf("wrong state version (%s) for capella sila payload", version.String(b.version))
		}
		latest, err := consensusblocks.PayloadToHeaderCapella(val)
		if err != nil {
			return errors.Wrap(err, "could not convert payload to header")
		}
		b.latestSilaPayloadHeaderCapella = latest
		b.markFieldAsDirty(types.LatestSilaPayloadHeaderCapella)
		return nil
	case *silaenginev1.SilaPayloadDeneb:
		if !(b.version >= version.Deneb) {
			return fmt.Errorf("wrong state version (%s) for deneb sila payload", version.String(b.version))
		}
		latest, err := consensusblocks.PayloadToHeaderDeneb(val)
		if err != nil {
			return errors.Wrap(err, "could not convert payload to header")
		}
		b.latestSilaPayloadHeaderDeneb = latest
		b.markFieldAsDirty(types.LatestSilaPayloadHeaderDeneb)
		return nil
	case *silaenginev1.SilaPayloadHeader:
		if b.version != version.Bellatrix {
			return fmt.Errorf("wrong state version (%s) for bellatrix sila payload header", version.String(b.version))
		}
		b.latestSilaPayloadHeader = header
		b.markFieldAsDirty(types.LatestSilaPayloadHeader)
		return nil
	case *silaenginev1.SilaPayloadHeaderCapella:
		if b.version != version.Capella {
			return fmt.Errorf("wrong state version (%s) for capella sila payload header", version.String(b.version))
		}
		b.latestSilaPayloadHeaderCapella = header
		b.markFieldAsDirty(types.LatestSilaPayloadHeaderCapella)
		return nil
	case *silaenginev1.SilaPayloadHeaderDeneb:
		if !(b.version >= version.Deneb) {
			return fmt.Errorf("wrong state version (%s) for deneb sila payload header", version.String(b.version))
		}
		b.latestSilaPayloadHeaderDeneb = header
		b.markFieldAsDirty(types.LatestSilaPayloadHeaderDeneb)
		return nil
	default:
		return errors.New("value must be an sila payload header")
	}
}
