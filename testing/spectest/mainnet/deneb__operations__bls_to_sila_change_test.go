package mainnet

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/deneb/operations"
)

func TestMainnet_Deneb_Operations_BLSToSilaChange(t *testing.T) {
	operations.RunBLSToSilaChangeTest(t, "mainnet")
}
