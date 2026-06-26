package minimal

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/electra/epoch_processing"
)

func TestMinimal_Electra_EpochProcessing_SilaDataReset(t *testing.T) {
	epoch_processing.RunSilaDataResetTests(t, "minimal")
}
