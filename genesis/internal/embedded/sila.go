package embedded

import (
	_ "embed"

	"github.com/OffchainLabs/prysm/v7/config/params"
)

var (
	//go:embed sila.ssz.snappy
	silaRawSSZCompressed []byte
)

func init() {
	embeddedStates[params.SilaMainnetName] = &silaRawSSZCompressed
}
