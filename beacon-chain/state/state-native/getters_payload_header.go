package state_native

import (
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// LatestSilaPayloadHeader of the beacon state.
func (b *BeaconState) LatestSilaPayloadHeader() (interfaces.SilaData, error) {
	if b.version < version.Bellatrix {
		return nil, errNotSupported("LatestSilaPayloadHeader", b.version)
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.version >= version.Deneb {
		return blocks.WrappedSilaPayloadHeaderDeneb(b.latestSilaPayloadHeaderDeneb.Copy())
	}

	if b.version >= version.Capella {
		return blocks.WrappedSilaPayloadHeaderCapella(b.latestSilaPayloadHeaderCapella.Copy())
	}

	if b.version >= version.Bellatrix {
		return blocks.WrappedSilaPayloadHeader(b.latestSilaPayloadHeader.Copy())
	}

	return nil, fmt.Errorf("unsupported version (%s) for latest sila payload header", version.String(b.version))
}
