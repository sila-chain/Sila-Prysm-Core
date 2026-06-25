package mainnet

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/gloas/operations"
)

func TestMainnet_Gloas_Operations_BLSToSilaChange(t *testing.T) {
	operations.RunBLSToSilaChangeTest(t, "mainnet")
}
