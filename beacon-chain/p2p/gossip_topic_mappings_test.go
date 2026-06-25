package p2p

import (
	"reflect"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"google.golang.org/protobuf/proto"
)

func TestMappingHasNoDuplicates(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	m := make(map[reflect.Type]bool)
	for _, v := range gossipTopicMappings {
		if _, ok := m[reflect.TypeOf(v())]; ok {
			t.Errorf("%T is duplicated in the topic mapping", v)
		}
		m[reflect.TypeFor[func() proto.Message]()] = true
	}
}

func TestGossipTopicMappings_CorrectType(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	bCfg := params.BeaconConfig().Copy()
	altairForkEpoch := primitives.Epoch(100)
	bellatrixForkEpoch := primitives.Epoch(200)
	capellaForkEpoch := primitives.Epoch(300)
	denebForkEpoch := primitives.Epoch(400)
	electraForkEpoch := primitives.Epoch(500)
	gloasForkEpoch := primitives.Epoch(550)
	fuluForkEpoch := primitives.Epoch(600)

	bCfg.AltairForkEpoch = altairForkEpoch
	bCfg.BellatrixForkEpoch = bellatrixForkEpoch
	bCfg.CapellaForkEpoch = capellaForkEpoch
	bCfg.DenebForkEpoch = denebForkEpoch
	bCfg.ElectraForkEpoch = electraForkEpoch
	bCfg.GloasForkEpoch = gloasForkEpoch
	bCfg.FuluForkEpoch = fuluForkEpoch
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.AltairForkVersion)] = primitives.Epoch(100)
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.BellatrixForkVersion)] = primitives.Epoch(200)
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.CapellaForkVersion)] = primitives.Epoch(300)
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.DenebForkVersion)] = primitives.Epoch(400)
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.ElectraForkVersion)] = primitives.Epoch(500)
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.GloasForkVersion)] = primitives.Epoch(550)
	bCfg.ForkVersionSchedule[bytesutil.ToBytes4(bCfg.FuluForkVersion)] = primitives.Epoch(600)
	params.OverrideBeaconConfig(bCfg)

	// Phase 0
	pMessage := GossipTopicMappings(BlockSubnetTopicFormat, 0)
	_, ok := pMessage.(*silapb.SignedBeaconBlock)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttestationSubnetTopicFormat, 0)
	_, ok = pMessage.(*silapb.Attestation)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttesterSlashingSubnetTopicFormat, 0)
	_, ok = pMessage.(*silapb.AttesterSlashing)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AggregateAndProofSubnetTopicFormat, 0)
	_, ok = pMessage.(*silapb.SignedAggregateAttestationAndProof)
	assert.Equal(t, true, ok)

	// Altair Fork
	pMessage = GossipTopicMappings(BlockSubnetTopicFormat, altairForkEpoch)
	_, ok = pMessage.(*silapb.SignedBeaconBlockAltair)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttestationSubnetTopicFormat, altairForkEpoch)
	_, ok = pMessage.(*silapb.Attestation)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttesterSlashingSubnetTopicFormat, altairForkEpoch)
	_, ok = pMessage.(*silapb.AttesterSlashing)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AggregateAndProofSubnetTopicFormat, altairForkEpoch)
	_, ok = pMessage.(*silapb.SignedAggregateAttestationAndProof)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientOptimisticUpdateTopicFormat, altairForkEpoch)
	_, ok = pMessage.(*silapb.LightClientOptimisticUpdateAltair)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientFinalityUpdateTopicFormat, altairForkEpoch)
	_, ok = pMessage.(*silapb.LightClientFinalityUpdateAltair)
	assert.Equal(t, true, ok)

	// Bellatrix Fork
	pMessage = GossipTopicMappings(BlockSubnetTopicFormat, bellatrixForkEpoch)
	_, ok = pMessage.(*silapb.SignedBeaconBlockBellatrix)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttestationSubnetTopicFormat, bellatrixForkEpoch)
	_, ok = pMessage.(*silapb.Attestation)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttesterSlashingSubnetTopicFormat, bellatrixForkEpoch)
	_, ok = pMessage.(*silapb.AttesterSlashing)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AggregateAndProofSubnetTopicFormat, bellatrixForkEpoch)
	_, ok = pMessage.(*silapb.SignedAggregateAttestationAndProof)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientOptimisticUpdateTopicFormat, bellatrixForkEpoch)
	_, ok = pMessage.(*silapb.LightClientOptimisticUpdateAltair)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientFinalityUpdateTopicFormat, bellatrixForkEpoch)
	_, ok = pMessage.(*silapb.LightClientFinalityUpdateAltair)
	assert.Equal(t, true, ok)

	// Capella Fork
	pMessage = GossipTopicMappings(BlockSubnetTopicFormat, capellaForkEpoch)
	_, ok = pMessage.(*silapb.SignedBeaconBlockCapella)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttestationSubnetTopicFormat, capellaForkEpoch)
	_, ok = pMessage.(*silapb.Attestation)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttesterSlashingSubnetTopicFormat, capellaForkEpoch)
	_, ok = pMessage.(*silapb.AttesterSlashing)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AggregateAndProofSubnetTopicFormat, capellaForkEpoch)
	_, ok = pMessage.(*silapb.SignedAggregateAttestationAndProof)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientOptimisticUpdateTopicFormat, capellaForkEpoch)
	_, ok = pMessage.(*silapb.LightClientOptimisticUpdateCapella)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientFinalityUpdateTopicFormat, capellaForkEpoch)
	_, ok = pMessage.(*silapb.LightClientFinalityUpdateCapella)
	assert.Equal(t, true, ok)

	// Deneb Fork
	pMessage = GossipTopicMappings(BlockSubnetTopicFormat, denebForkEpoch)
	_, ok = pMessage.(*silapb.SignedBeaconBlockDeneb)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttestationSubnetTopicFormat, denebForkEpoch)
	_, ok = pMessage.(*silapb.Attestation)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttesterSlashingSubnetTopicFormat, denebForkEpoch)
	_, ok = pMessage.(*silapb.AttesterSlashing)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AggregateAndProofSubnetTopicFormat, denebForkEpoch)
	_, ok = pMessage.(*silapb.SignedAggregateAttestationAndProof)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientOptimisticUpdateTopicFormat, denebForkEpoch)
	_, ok = pMessage.(*silapb.LightClientOptimisticUpdateDeneb)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientFinalityUpdateTopicFormat, denebForkEpoch)
	_, ok = pMessage.(*silapb.LightClientFinalityUpdateDeneb)
	assert.Equal(t, true, ok)

	// Electra Fork
	pMessage = GossipTopicMappings(BlockSubnetTopicFormat, electraForkEpoch)
	_, ok = pMessage.(*silapb.SignedBeaconBlockElectra)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttestationSubnetTopicFormat, electraForkEpoch)
	_, ok = pMessage.(*silapb.SingleAttestation)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AttesterSlashingSubnetTopicFormat, electraForkEpoch)
	_, ok = pMessage.(*silapb.AttesterSlashingElectra)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(AggregateAndProofSubnetTopicFormat, electraForkEpoch)
	_, ok = pMessage.(*silapb.SignedAggregateAttestationAndProofElectra)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientOptimisticUpdateTopicFormat, electraForkEpoch)
	_, ok = pMessage.(*silapb.LightClientOptimisticUpdateDeneb)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(LightClientFinalityUpdateTopicFormat, electraForkEpoch)
	_, ok = pMessage.(*silapb.LightClientFinalityUpdateElectra)
	assert.Equal(t, true, ok)

	// Gloas Fork
	pMessage = GossipTopicMappings(BlockSubnetTopicFormat, gloasForkEpoch)
	_, ok = pMessage.(*silapb.SignedBeaconBlockGloas)
	assert.Equal(t, true, ok)
	pMessage = GossipTopicMappings(ExecutionPayloadBidTopicFormat, gloasForkEpoch)
	_, ok = pMessage.(*silapb.SignedExecutionPayloadBid)
	assert.Equal(t, true, ok)
	assert.Equal(t, ExecutionPayloadBidTopicFormat, GossipTypeMapping[reflect.TypeFor[*silapb.SignedExecutionPayloadBid]()])
	pMessage = GossipTopicMappings(SignedProposerPreferencesTopicFormat, gloasForkEpoch)
	_, ok = pMessage.(*silapb.SignedProposerPreferences)
	assert.Equal(t, true, ok)
	assert.Equal(t, SignedProposerPreferencesTopicFormat, GossipTypeMapping[reflect.TypeFor[*silapb.SignedProposerPreferences]()])
}
