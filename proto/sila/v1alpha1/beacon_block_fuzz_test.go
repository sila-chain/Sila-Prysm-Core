package v1alpha1_test
import (
	"testing"

	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func TestBeaconBlock_Fuzz(t *testing.T) {
	// Phase 0 Full
	fuzzCopies(t, &sila.SignedBeaconBlock{})
	fuzzCopies(t, &sila.BeaconBlock{})
	fuzzCopies(t, &sila.BeaconBlockBody{})
	// Altair Full
	fuzzCopies(t, &sila.SignedBeaconBlockAltair{})
	fuzzCopies(t, &sila.BeaconBlockAltair{})
	fuzzCopies(t, &sila.BeaconBlockBodyAltair{})
	// Bellatrix Full
	fuzzCopies(t, &sila.SignedBeaconBlockBellatrix{})
	fuzzCopies(t, &sila.BeaconBlockBellatrix{})
	fuzzCopies(t, &sila.BeaconBlockBodyBellatrix{})
	// Bellatrix Blinded
	fuzzCopies(t, &sila.SignedBlindedBeaconBlockBellatrix{})
	fuzzCopies(t, &sila.BlindedBeaconBlockBellatrix{})
	fuzzCopies(t, &sila.BlindedBeaconBlockBodyBellatrix{})
	// Capella Full
	fuzzCopies(t, &sila.SignedBeaconBlockCapella{})
	fuzzCopies(t, &sila.BeaconBlockCapella{})
	fuzzCopies(t, &sila.BeaconBlockBodyCapella{})
	// Capella Blinded
	fuzzCopies(t, &sila.SignedBlindedBeaconBlockCapella{})
	fuzzCopies(t, &sila.BlindedBeaconBlockCapella{})
	fuzzCopies(t, &sila.BlindedBeaconBlockBodyCapella{})
	// Deneb Full
	fuzzCopies(t, &sila.SignedBeaconBlockDeneb{})
	fuzzCopies(t, &sila.BeaconBlockDeneb{})
	fuzzCopies(t, &sila.BeaconBlockBodyDeneb{})
	// Deneb Blinded
	fuzzCopies(t, &sila.SignedBlindedBeaconBlockDeneb{})
	fuzzCopies(t, &sila.BlindedBeaconBlockDeneb{})
	fuzzCopies(t, &sila.BlindedBeaconBlockBodyDeneb{})
	// Electra Full
	fuzzCopies(t, &sila.SignedBeaconBlockElectra{})
	fuzzCopies(t, &sila.BeaconBlockElectra{})
	fuzzCopies(t, &sila.BeaconBlockBodyElectra{})
	// Electra Blinded
	fuzzCopies(t, &sila.SignedBlindedBeaconBlockElectra{})
	fuzzCopies(t, &sila.BlindedBeaconBlockElectra{})
	fuzzCopies(t, &sila.BlindedBeaconBlockBodyElectra{})
}

func TestCopyBeaconBlockFields_Fuzz(t *testing.T) {
	fuzzCopies(t, &sila.SilaData{})
	fuzzCopies(t, &sila.ProposerSlashing{})
	fuzzCopies(t, &sila.SignedBeaconBlockHeader{})
	fuzzCopies(t, &sila.BeaconBlockHeader{})
	fuzzCopies(t, &sila.Deposit{})
	fuzzCopies(t, &sila.Deposit_Data{})
	fuzzCopies(t, &sila.SignedVoluntaryExit{})
	fuzzCopies(t, &sila.VoluntaryExit{})
	fuzzCopies(t, &sila.SyncAggregate{})
	fuzzCopies(t, &sila.SignedBLSToSilaChange{})
	fuzzCopies(t, &sila.BLSToSilaChange{})
	fuzzCopies(t, &sila.HistoricalSummary{})
	fuzzCopies(t, &sila.PendingDeposit{})
}
