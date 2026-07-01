package v1alpha1_test
import (
	"testing"

	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func TestCopyAttestation_Fuzz(t *testing.T) {
	fuzzCopies(t, &sila.Checkpoint{})
	fuzzCopies(t, &sila.AttestationData{})
	fuzzCopies(t, &sila.Attestation{})
	fuzzCopies(t, &sila.AttestationElectra{})
	fuzzCopies(t, &sila.PendingAttestation{})
	fuzzCopies(t, &sila.IndexedAttestation{})
	fuzzCopies(t, &sila.IndexedAttestationElectra{})
	fuzzCopies(t, &sila.AttesterSlashing{})
	fuzzCopies(t, &sila.AttesterSlashingElectra{})
}
