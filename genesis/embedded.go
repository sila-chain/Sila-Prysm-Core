package genesis

import (
	"time"

	"github.com/OffchainLabs/prysm/v7/beacon-chain/state"
	"github.com/OffchainLabs/prysm/v7/config/params"
	"github.com/OffchainLabs/prysm/v7/genesis/internal/embedded"
)

var embeddedGenesisData map[string]GenesisData

func init() {
	embeddedGenesisData = make(map[string]GenesisData)
	embeddedGenesisData[params.MainnetName] = GenesisData{
		ValidatorsRoot: [32]byte{75, 54, 61, 185, 78, 40, 97, 32, 215, 110, 185, 5, 52, 15, 221, 78, 84, 191, 233, 240, 107, 243, 63, 246, 207, 90, 210, 127, 81, 27, 254, 149},
		Time:           time.Unix(1606824023, 0),
		embeddedBytes: func() ([]byte, error) {
			return embedded.BytesByName(params.MainnetName)
		},
		embeddedState: func() (state.BeaconState, error) {
			return embedded.ByName(params.MainnetName)
		},
	}
	embeddedGenesisData[params.SilaMainnetName] = GenesisData{
		ValidatorsRoot: [32]byte{0x83, 0x43, 0x1e, 0xc7, 0xfc, 0xf9, 0x2c, 0xfc, 0x44, 0x94, 0x7f, 0xc0, 0x41, 0x8e, 0x83, 0x1c, 0x25, 0xe1, 0xd0, 0x80, 0x65, 0x90, 0x23, 0x1c, 0x43, 0x98, 0x30, 0xdb, 0x7a, 0xd5, 0x4f, 0xda},
		Time:           time.Unix(1893456000, 0),
		embeddedBytes: func() ([]byte, error) {
			return embedded.BytesByName(params.SilaMainnetName)
		},
		embeddedState: func() (state.BeaconState, error) {
			return embedded.ByName(params.SilaMainnetName)
		},
	}
}
