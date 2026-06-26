package state_native_test

import (
	"fmt"
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestSetLatestSilaPayloadHeader(t *testing.T) {
	versionOffset := version.Bellatrix // PayloadHeader only applies in Bellatrix and beyond.
	payloads := []interfaces.SilaData{
		func() interfaces.SilaData {
			e := util.NewBeaconBlockBellatrix().Block.Body.SilaPayload
			ee, err := blocks.WrappedSilaPayload(e)
			require.NoError(t, err)
			return ee
		}(),
		func() interfaces.SilaData {
			e := util.NewBeaconBlockCapella().Block.Body.SilaPayload
			ee, err := blocks.WrappedSilaPayloadCapella(e)
			require.NoError(t, err)
			return ee
		}(),
		func() interfaces.SilaData {
			e := util.NewBeaconBlockDeneb().Block.Body.SilaPayload
			ee, err := blocks.WrappedSilaPayloadDeneb(e)
			require.NoError(t, err)
			return ee
		}(),
		func() interfaces.SilaData {
			e := util.NewBeaconBlockElectra().Block.Body.SilaPayload
			ee, err := blocks.WrappedSilaPayloadDeneb(e)
			require.NoError(t, err)
			return ee
		}(),
	}

	payloadHeaders := []interfaces.SilaData{
		func() interfaces.SilaData {
			e := util.NewBlindedBeaconBlockBellatrix().Block.Body.SilaPayloadHeader
			ee, err := blocks.WrappedSilaPayloadHeader(e)
			require.NoError(t, err)
			return ee
		}(),
		func() interfaces.SilaData {
			e := util.NewBlindedBeaconBlockCapella().Block.Body.SilaPayloadHeader
			ee, err := blocks.WrappedSilaPayloadHeaderCapella(e)
			require.NoError(t, err)
			return ee
		}(),
		func() interfaces.SilaData {
			e := util.NewBlindedBeaconBlockDeneb().Message.Body.SilaPayloadHeader
			ee, err := blocks.WrappedSilaPayloadHeaderDeneb(e)
			require.NoError(t, err)
			return ee
		}(),
		func() interfaces.SilaData {
			e := util.NewBlindedBeaconBlockElectra().Message.Body.SilaPayloadHeader
			ee, err := blocks.WrappedSilaPayloadHeaderDeneb(e)
			require.NoError(t, err)
			return ee
		}(),
	}

	t.Run("can set payload", func(t *testing.T) {
		for i, p := range payloads {
			t.Run(version.String(i+versionOffset), func(t *testing.T) {
				s := state_native.EmptyStateFromVersion(t, i+versionOffset)
				require.NoError(t, s.SetLatestSilaPayloadHeader(p))
			})
		}
	})

	t.Run("can set payload header", func(t *testing.T) {
		for i, ph := range payloadHeaders {
			t.Run(version.String(i+versionOffset), func(t *testing.T) {
				s := state_native.EmptyStateFromVersion(t, i+versionOffset)
				require.NoError(t, s.SetLatestSilaPayloadHeader(ph))
			})
		}
	})

	t.Run("mismatched type version returns error", func(t *testing.T) {
		require.Equal(t, len(payloads), len(payloadHeaders), "This test will fail if the payloads and payload headers are not same length")
		for i := range payloads {
			for j := range payloads {
				if i == j {
					continue
				}
				// Skip Deneb-Electra combinations
				if i == len(payloads)-1 && j == len(payloads)-2 {
					continue
				}
				if i == len(payloads)-2 && j == len(payloads)-1 {
					continue
				}
				t.Run(fmt.Sprintf("%s state with %s payload", version.String(i+versionOffset), version.String(j+versionOffset)), func(t *testing.T) {
					s := state_native.EmptyStateFromVersion(t, i+versionOffset)
					p := payloads[j]
					require.ErrorContains(t, "wrong state version", s.SetLatestSilaPayloadHeader(p))
					ph := payloadHeaders[j]
					require.ErrorContains(t, "wrong state version", s.SetLatestSilaPayloadHeader(ph))
				})
			}
		}
	})

}
