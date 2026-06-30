// package mock
// lint:nopanic -- This is test / mock code, allowed to panic.
package mock

import (
	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	validatorpb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/validator-client"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaapi/v1"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	ssz "github.com/sila-chain/fastssz"
	"google.golang.org/protobuf/proto"
)

type SignedBeaconBlock struct {
	BeaconBlock interfaces.ReadOnlyBeaconBlock
}

func (SignedBeaconBlock) PbGenericBlock() (*sila.GenericSignedBeaconBlock, error) {
	panic("implement me")
}

func (m SignedBeaconBlock) Block() interfaces.ReadOnlyBeaconBlock {
	return m.BeaconBlock
}

func (SignedBeaconBlock) Signature() [field_params.BLSSignatureLength]byte {
	panic("implement me")
}

func (SignedBeaconBlock) SetSignature([]byte) {
	panic("implement me")
}

func (m SignedBeaconBlock) IsNil() bool {
	return m.BeaconBlock == nil || m.Block().IsNil()
}

func (SignedBeaconBlock) Copy() (interfaces.SignedBeaconBlock, error) {
	panic("implement me")
}

func (SignedBeaconBlock) Proto() (proto.Message, error) {
	panic("implement me")
}

func (SignedBeaconBlock) MarshalSSZTo(_ []byte) ([]byte, error) {
	panic("implement me")
}

func (SignedBeaconBlock) MarshalSSZ() ([]byte, error) {
	panic("implement me")
}

func (SignedBeaconBlock) SizeSSZ() int {
	panic("implement me")
}

func (SignedBeaconBlock) UnmarshalSSZ(_ []byte) error {
	panic("implement me")
}

func (SignedBeaconBlock) Version() int {
	panic("implement me")
}

func (SignedBeaconBlock) IsBlinded() bool {
	return false
}

func (SignedBeaconBlock) ToBlinded() (interfaces.ReadOnlySignedBeaconBlock, error) {
	panic("implement me")
}

func (SignedBeaconBlock) Header() (*sila.SignedBeaconBlockHeader, error) {
	panic("implement me")
}

type BeaconBlock struct {
	Htr             [field_params.RootLength]byte
	HtrErr          error
	BeaconBlockBody interfaces.ReadOnlyBeaconBlockBody
	BlockSlot       primitives.Slot
}

func (BeaconBlock) AsSignRequestObject() (validatorpb.SignRequestObject, error) {
	panic("implement me")
}

func (m BeaconBlock) HashTreeRoot() ([field_params.RootLength]byte, error) {
	return m.Htr, m.HtrErr
}

func (m BeaconBlock) Slot() primitives.Slot {
	return m.BlockSlot
}

func (BeaconBlock) ProposerIndex() primitives.ValidatorIndex {
	panic("implement me")
}

func (BeaconBlock) ParentRoot() [field_params.RootLength]byte {
	panic("implement me")
}

func (BeaconBlock) StateRoot() [field_params.RootLength]byte {
	panic("implement me")
}

func (m BeaconBlock) Body() interfaces.ReadOnlyBeaconBlockBody {
	return m.BeaconBlockBody
}

func (BeaconBlock) IsNil() bool {
	return false
}

func (BeaconBlock) IsBlinded() bool {
	return false
}

func (BeaconBlock) Proto() (proto.Message, error) {
	panic("implement me")
}

func (BeaconBlock) MarshalSSZTo(_ []byte) ([]byte, error) {
	panic("implement me")
}

func (BeaconBlock) MarshalSSZ() ([]byte, error) {
	panic("implement me")
}

func (BeaconBlock) SizeSSZ() int {
	panic("implement me")
}

func (BeaconBlock) UnmarshalSSZ(_ []byte) error {
	panic("implement me")
}

func (BeaconBlock) HashTreeRootWith(_ *ssz.Hasher) error {
	panic("implement me")
}

func (BeaconBlock) Version() int {
	panic("implement me")
}

func (BeaconBlock) ToBlinded() (interfaces.ReadOnlyBeaconBlock, error) {
	panic("implement me")
}

func (BeaconBlock) SetSlot(_ primitives.Slot) {
	panic("implement me")
}

func (BeaconBlock) SetProposerIndex(_ primitives.ValidatorIndex) {
	panic("implement me")
}

func (BeaconBlock) SetParentRoot(_ []byte) {
	panic("implement me")
}

type BeaconBlockBody struct{}

func (BeaconBlockBody) RandaoReveal() [field_params.BLSSignatureLength]byte {
	panic("implement me")
}

func (BeaconBlockBody) SilaChainData() *sila.SilaData {
	panic("implement me")
}

func (BeaconBlockBody) Graffiti() [field_params.RootLength]byte {
	panic("implement me")
}

func (BeaconBlockBody) ProposerSlashings() []*sila.ProposerSlashing {
	panic("implement me")
}

func (BeaconBlockBody) AttesterSlashings() []sila.AttSlashing {
	panic("implement me")
}

func (BeaconBlockBody) Deposits() []*sila.Deposit {
	panic("implement me")
}

func (BeaconBlockBody) VoluntaryExits() []*sila.SignedVoluntaryExit {
	panic("implement me")
}

func (BeaconBlockBody) SyncAggregate() (*sila.SyncAggregate, error) {
	panic("implement me")
}

func (BeaconBlockBody) IsNil() bool {
	return false
}

func (BeaconBlockBody) HashTreeRoot() ([field_params.RootLength]byte, error) {
	panic("implement me")
}

func (BeaconBlockBody) Proto() (proto.Message, error) {
	panic("implement me")
}

func (BeaconBlockBody) SilaData() (interfaces.SilaData, error) {
	panic("implement me")
}

func (BeaconBlockBody) BLSToSilaChanges() ([]*sila.SignedBLSToSilaChange, error) {
	panic("implement me")
}

func (b *BeaconBlock) SetStateRoot(root []byte) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetRandaoReveal([]byte) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetSilaChainData(*sila.SilaData) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetGraffiti([]byte) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetProposerSlashings([]*sila.ProposerSlashing) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetAttesterSlashings([]silapb.AttesterSlashing) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetAttestations([]*sila.Attestation) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetDeposits([]*sila.Deposit) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetVoluntaryExits([]*sila.SignedVoluntaryExit) {
	panic("implement me")
}

func (b *BeaconBlockBody) SetSyncAggregate(*sila.SyncAggregate) error {
	panic("implement me")
}

func (b *BeaconBlockBody) SetSilaData(interfaces.SilaData) error {
	panic("implement me")
}

func (b *BeaconBlockBody) SetBLSToSilaChanges([]*sila.SignedBLSToSilaChange) error {
	panic("implement me")
}

// BlobKzgCommitments returns the blob kzg commitments in the block.
func (b *BeaconBlockBody) BlobKzgCommitments() ([][]byte, error) {
	panic("implement me")
}

func (b *BeaconBlockBody) SilaRequests() (*silaenginev1.SilaRequests, error) {
	panic("implement me")
}

func (b *BeaconBlockBody) PayloadAttestations() ([]*sila.PayloadAttestation, error) {
	panic("implement me")
}

func (b *BeaconBlockBody) SignedSilaPayloadBid() (*sila.SignedSilaPayloadBid, error) {
	panic("implement me")
}

func (b *BeaconBlockBody) ParentSilaRequests() (*silaenginev1.SilaRequests, error) {
	panic("implement me")
}

func (b *BeaconBlockBody) Attestations() []sila.Att {
	panic("implement me")
}
func (b *BeaconBlockBody) Version() int {
	panic("implement me")
}

var _ interfaces.ReadOnlySignedBeaconBlock = &SignedBeaconBlock{}
var _ interfaces.ReadOnlyBeaconBlock = &BeaconBlock{}
var _ interfaces.ReadOnlyBeaconBlockBody = &BeaconBlockBody{}
