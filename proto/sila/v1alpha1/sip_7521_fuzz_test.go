package v1alpha1_test
import (
	"testing"

	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func TestCopySip7521Types_Fuzz(t *testing.T) {
	fuzzCopies(t, &sila.PendingDeposit{})
	fuzzCopies(t, &sila.PendingPartialWithdrawal{})
	fuzzCopies(t, &sila.PendingConsolidation{})
}
