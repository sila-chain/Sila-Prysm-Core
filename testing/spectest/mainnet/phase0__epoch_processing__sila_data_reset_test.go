package mainnet

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/phase0/epoch_processing"
)

func TestMainnet_Phase0_EpochProcessing_SilaDataReset(t *testing.T) {
	epoch_processing.RunSilaDataResetTests(t, "mainnet")
}
