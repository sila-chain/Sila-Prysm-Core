package types

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	lightclientConsensusTypes "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/light-client"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/wrapper"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	enginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/engine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/metadata"
	ssz "github.com/sila-chain/fastssz"
)

func init() {
	// Initialize data maps.
	InitializeDataMaps()
}

// This file provides a mapping of fork version to the respective data type. This is
// to allow any service to appropriately use the correct data type with a provided
// fork-version.

var (
	// BlockMap maps the fork-version to the underlying data type for that
	// particular fork period.
	BlockMap map[[4]byte]func() (interfaces.ReadOnlySignedBeaconBlock, error)
	// MetaDataMap maps the fork-version to the underlying data type for that
	// particular fork period.
	MetaDataMap map[[4]byte]func() (metadata.Metadata, error)
	// AttestationMap maps the fork-version to the underlying data type for that
	// particular fork period.
	AttestationMap map[[4]byte]func() (silapb.Att, error)
	// AggregateAttestationMap maps the fork-version to the underlying data type for that
	// particular fork period.
	AggregateAttestationMap map[[4]byte]func() (silapb.SignedAggregateAttAndProof, error)
	// AttesterSlashingMap maps the fork-version to the underlying data type for that particular
	// fork period.
	AttesterSlashingMap map[[4]byte]func() (silapb.AttSlashing, error)
	// LightClientOptimisticUpdateMap maps the fork-version to the underlying data type for that
	// particular fork period.
	LightClientOptimisticUpdateMap map[[4]byte]func() (interfaces.LightClientOptimisticUpdate, error)
	// LightClientFinalityUpdateMap maps the fork-version to the underlying data type for that
	// particular fork period.
	LightClientFinalityUpdateMap map[[4]byte]func() (interfaces.LightClientFinalityUpdate, error)
	// DataColumnSidecarMap maps the fork-version to the underlying data column sidecar type.
	DataColumnSidecarMap map[[4]byte]func() (ssz.Unmarshaler, error)
)

// InitializeDataMaps initializes all the relevant object maps. This function is called to
// reset maps and reinitialize them.
func InitializeDataMaps() {
	// Reset our block map.
	BlockMap = map[[4]byte]func() (interfaces.ReadOnlySignedBeaconBlock, error){
		bytesutil.ToBytes4(params.BeaconConfig().GenesisForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlock{Block: &silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockAltair{Block: &silapb.BeaconBlockAltair{Body: &silapb.BeaconBlockBodyAltair{}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockBellatrix{Block: &silapb.BeaconBlockBellatrix{Body: &silapb.BeaconBlockBodyBellatrix{ExecutionPayload: &enginev1.ExecutionPayload{}}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockCapella{Block: &silapb.BeaconBlockCapella{Body: &silapb.BeaconBlockBodyCapella{ExecutionPayload: &enginev1.ExecutionPayloadCapella{}}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockDeneb{Block: &silapb.BeaconBlockDeneb{Body: &silapb.BeaconBlockBodyDeneb{ExecutionPayload: &enginev1.ExecutionPayloadDeneb{}}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockElectra{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{ExecutionPayload: &enginev1.ExecutionPayloadDeneb{}, ExecutionRequests: &enginev1.ExecutionRequests{}}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockGloas{Block: &silapb.BeaconBlockGloas{Body: &silapb.BeaconBlockBodyGloas{}}},
			)
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (interfaces.ReadOnlySignedBeaconBlock, error) {
			return blocks.NewSignedBeaconBlock(
				&silapb.SignedBeaconBlockFulu{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{ExecutionPayload: &enginev1.ExecutionPayloadDeneb{}, ExecutionRequests: &enginev1.ExecutionRequests{}}}},
			)
		},
	}

	// Reset our metadata map.
	MetaDataMap = map[[4]byte]func() (metadata.Metadata, error){
		bytesutil.ToBytes4(params.BeaconConfig().GenesisForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV0(&silapb.MetaDataV0{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV1(&silapb.MetaDataV1{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV2(&silapb.MetaDataV2{}), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (metadata.Metadata, error) {
			return wrapper.WrappedMetadataV2(&silapb.MetaDataV2{}), nil
		},
	}

	// Reset our attestation map.
	AttestationMap = map[[4]byte]func() (silapb.Att, error){
		bytesutil.ToBytes4(params.BeaconConfig().GenesisForkVersion): func() (silapb.Att, error) {
			return &silapb.Attestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (silapb.Att, error) {
			return &silapb.Attestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (silapb.Att, error) {
			return &silapb.Attestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (silapb.Att, error) {
			return &silapb.Attestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (silapb.Att, error) {
			return &silapb.Attestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (silapb.Att, error) {
			return &silapb.SingleAttestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (silapb.Att, error) {
			return &silapb.SingleAttestation{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (silapb.Att, error) {
			return &silapb.SingleAttestation{}, nil
		},
	}

	// Reset our aggregate attestation map.
	AggregateAttestationMap = map[[4]byte]func() (silapb.SignedAggregateAttAndProof, error){
		bytesutil.ToBytes4(params.BeaconConfig().GenesisForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProof{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProof{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProof{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProof{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProof{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProofElectra{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProofElectra{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (silapb.SignedAggregateAttAndProof, error) {
			return &silapb.SignedAggregateAttestationAndProofElectra{}, nil
		},
	}

	// Reset our aggregate attestation map.
	AttesterSlashingMap = map[[4]byte]func() (silapb.AttSlashing, error){
		bytesutil.ToBytes4(params.BeaconConfig().GenesisForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashing{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashing{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashing{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashing{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashing{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashingElectra{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashingElectra{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (silapb.AttSlashing, error) {
			return &silapb.AttesterSlashingElectra{}, nil
		},
	}

	// Reset our light client optimistic update map.
	LightClientOptimisticUpdateMap = map[[4]byte]func() (interfaces.LightClientOptimisticUpdate, error){
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateAltair(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateAltair(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateCapella(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateDeneb(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateDeneb(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateDeneb(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (interfaces.LightClientOptimisticUpdate, error) {
			return lightclientConsensusTypes.NewEmptyOptimisticUpdateDeneb(), nil
		},
	}

	// Reset our light client finality update map.
	LightClientFinalityUpdateMap = map[[4]byte]func() (interfaces.LightClientFinalityUpdate, error){
		bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateAltair(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateAltair(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateCapella(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().DenebForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateDeneb(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().ElectraForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateElectra(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateElectra(), nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (interfaces.LightClientFinalityUpdate, error) {
			return lightclientConsensusTypes.NewEmptyFinalityUpdateElectra(), nil
		},
	}

	DataColumnSidecarMap = map[[4]byte]func() (ssz.Unmarshaler, error){
		bytesutil.ToBytes4(params.BeaconConfig().FuluForkVersion): func() (ssz.Unmarshaler, error) {
			return &silapb.DataColumnSidecar{}, nil
		},
		bytesutil.ToBytes4(params.BeaconConfig().GloasForkVersion): func() (ssz.Unmarshaler, error) {
			return &silapb.DataColumnSidecarGloas{}, nil
		},
	}
}
