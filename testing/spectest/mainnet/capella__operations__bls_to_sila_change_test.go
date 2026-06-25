package mainnet

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/capella/operations"
)

func TestMainnet_Capella_Operations_BLSToSilaChange(t *testing.T) {
	operations.RunBLSToSilaChangeTest(t, "mainnet")
}
