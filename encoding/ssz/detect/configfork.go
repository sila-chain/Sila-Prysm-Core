package detect

import (
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
	ssz "github.com/sila-chain/fastssz"
)

// VersionedUnmarshaler represents the intersection of Configuration (eg mainnet, testnet) and Fork (eg phase0, altair).
// Using a detected VersionedUnmarshaler, a BeaconState or ReadOnlySignedBeaconBlock can be correctly unmarshaled without the need to
// hard code a concrete type in paths where only the marshaled bytes, or marshaled bytes and a version, are available.
type VersionedUnmarshaler struct {
	Config *params.BeaconChainConfig
	// Fork aligns with the fork names in config/params/values.go
	Fork int
	// Version corresponds to the Version type defined in the beacon-chain spec, aka a "fork version number":
	// https://github.com/sila-chain/Sila-Consensus-Specs/blob/master/specs/phase0/beacon-chain.md#custom-types
	Version [fieldparams.VersionLength]byte
}

var beaconStateCurrentVersion = fieldSpec{
	// 52 = 8 (genesis_time) + 32 (genesis_validators_root) + 8 (slot) + 4 (previous_version)
	offset: 52,
	t:      typeBytes4,
}

// FromState exploits the fixed-size lower-order bytes in a BeaconState as a heuristic to obtain the value of the
// state.version field without first unmarshaling the BeaconState. The Version is then internally used to lookup
// the correct ConfigVersion.
func FromState(marshaled []byte) (*VersionedUnmarshaler, error) {
	cv, err := beaconStateCurrentVersion.bytes4(marshaled)
	if err != nil {
		return nil, err
	}
	return FromForkVersion(cv)
}

// FromBlock uses the known size of an offset and signature to determine the slot of a block without unmarshalling it.
// The slot is used to determine the version along with the respective config.
func FromBlock(marshaled []byte) (*VersionedUnmarshaler, error) {
	slot, err := slotFromBlock(marshaled)
	if err != nil {
		return nil, err
	}
	epoch := slots.ToEpoch(slot)
	fs, err := params.Fork(epoch)
	if err != nil {
		return nil, err
	}
	return FromForkVersion([4]byte(fs.CurrentVersion))
}

var ErrForkNotFound = errors.New("version found in fork schedule but can't be matched to a named fork")

// FromForkVersion uses a lookup table to resolve a Version (from a beacon node api for instance, or obtained by peeking at
// the bytes of a marshaled BeaconState) to a VersionedUnmarshaler.
func FromForkVersion(cv [fieldparams.VersionLength]byte) (*VersionedUnmarshaler, error) {
	cfg, err := params.ByVersion(cv)
	if err != nil {
		return nil, err
	}
	var fork int
	switch cv {
	case bytesutil.ToBytes4(cfg.GenesisForkVersion):
		fork = version.Phase0
	case bytesutil.ToBytes4(cfg.AltairForkVersion):
		fork = version.Altair
	case bytesutil.ToBytes4(cfg.BellatrixForkVersion):
		fork = version.Bellatrix
	case bytesutil.ToBytes4(cfg.CapellaForkVersion):
		fork = version.Capella
	case bytesutil.ToBytes4(cfg.DenebForkVersion):
		fork = version.Deneb
	case bytesutil.ToBytes4(cfg.ElectraForkVersion):
		fork = version.Electra
	case bytesutil.ToBytes4(cfg.FuluForkVersion):
		fork = version.Fulu
	case bytesutil.ToBytes4(cfg.GloasForkVersion):
		fork = version.Gloas
	default:
		return nil, errors.Wrapf(ErrForkNotFound, "version=%#x", cv)
	}
	return &VersionedUnmarshaler{
		Config:  cfg,
		Fork:    fork,
		Version: cv,
	}, nil
}

// UnmarshalBeaconState uses internal knowledge in the VersionedUnmarshaler to pick the right concrete BeaconState type,
// then Unmarshal()s the type and returns an instance of state.BeaconState if successful.
func (cf *VersionedUnmarshaler) UnmarshalBeaconState(marshaled []byte) (s state.BeaconState, err error) {
	forkName := version.String(cf.Fork)
	switch fork := cf.Fork; fork {
	case version.Phase0:
		st := &silapb.BeaconState{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafePhase0(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Altair:
		st := &silapb.BeaconStateAltair{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafeAltair(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Bellatrix:
		st := &silapb.BeaconStateBellatrix{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafeBellatrix(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Capella:
		st := &silapb.BeaconStateCapella{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafeCapella(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Deneb:
		st := &silapb.BeaconStateDeneb{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafeDeneb(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Electra:
		st := &silapb.BeaconStateElectra{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafeElectra(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Fulu:
		st := &silapb.BeaconStateFulu{}
		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}
		s, err = state_native.InitializeFromProtoUnsafeFulu(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	case version.Gloas:
		st := &silapb.BeaconStateGloas{}

		err = st.UnmarshalSSZ(marshaled)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal state, detected fork=%s", forkName)
		}

		s, err = state_native.InitializeFromProtoUnsafeGloas(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to init state trie from state, detected fork=%s", forkName)
		}
	default:
		return nil, fmt.Errorf("unable to initialize BeaconState for fork version=%s", forkName)
	}

	return s, nil
}

var beaconBlockSlot = fieldSpec{
	// ssz variable length offset (not to be confused with the fieldSpec offset) is a uint32
	// variable length. Offsets come before fixed length data, so that's 4 bytes at the beginning
	// then signature is 96 bytes, 4+96 = 100
	offset: 100,
	t:      typeUint64,
}

func slotFromBlock(marshaled []byte) (primitives.Slot, error) {
	slot, err := beaconBlockSlot.uint64(marshaled)
	if err != nil {
		return 0, err
	}
	return primitives.Slot(slot), nil
}

var errBlockForkMismatch = errors.New("fork or config detected in unmarshaler is different than block")

// UnmarshalBeaconBlock uses internal knowledge in the VersionedUnmarshaler to pick the right concrete ReadOnlySignedBeaconBlock type,
// then Unmarshal()s the type and returns an instance of block.ReadOnlySignedBeaconBlock if successful.
func (cf *VersionedUnmarshaler) UnmarshalBeaconBlock(marshaled []byte) (interfaces.SignedBeaconBlock, error) {
	slot, err := slotFromBlock(marshaled)
	if err != nil {
		return nil, err
	}
	if err := cf.validateVersion(slot); err != nil {
		return nil, err
	}

	var blk ssz.Unmarshaler
	switch cf.Fork {
	case version.Phase0:
		blk = &silapb.SignedBeaconBlock{}
	case version.Altair:
		blk = &silapb.SignedBeaconBlockAltair{}
	case version.Bellatrix:
		blk = &silapb.SignedBeaconBlockBellatrix{}
	case version.Capella:
		blk = &silapb.SignedBeaconBlockCapella{}
	case version.Deneb:
		blk = &silapb.SignedBeaconBlockDeneb{}
	case version.Electra:
		blk = &silapb.SignedBeaconBlockElectra{}
	case version.Fulu:
		blk = &silapb.SignedBeaconBlockFulu{}
	case version.Gloas:
		blk = &silapb.SignedBeaconBlockGloas{}
	default:
		forkName := version.String(cf.Fork)
		return nil, fmt.Errorf("unable to initialize ReadOnlyBeaconBlock for fork version=%s at slot=%d", forkName, slot)
	}
	err = blk.UnmarshalSSZ(marshaled)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ReadOnlySignedBeaconBlock in UnmarshalSSZ")
	}
	return blocks.NewSignedBeaconBlock(blk)
}

// UnmarshalBlindedBeaconBlock uses internal knowledge in the VersionedUnmarshaler to pick the right concrete blinded ReadOnlySignedBeaconBlock type,
// then Unmarshal()s the type and returns an instance of block.ReadOnlySignedBeaconBlock if successful.
// For Phase0 and Altair it works exactly line UnmarshalBeaconBlock.
func (cf *VersionedUnmarshaler) UnmarshalBlindedBeaconBlock(marshaled []byte) (interfaces.SignedBeaconBlock, error) {
	slot, err := slotFromBlock(marshaled)
	if err != nil {
		return nil, err
	}
	if err := cf.validateVersion(slot); err != nil {
		return nil, err
	}

	var blk ssz.Unmarshaler
	switch cf.Fork {
	case version.Phase0:
		blk = &silapb.SignedBeaconBlock{}
	case version.Altair:
		blk = &silapb.SignedBeaconBlockAltair{}
	case version.Bellatrix:
		blk = &silapb.SignedBlindedBeaconBlockBellatrix{}
	case version.Capella:
		blk = &silapb.SignedBlindedBeaconBlockCapella{}
	case version.Deneb:
		blk = &silapb.SignedBlindedBeaconBlockDeneb{}
	case version.Electra:
		blk = &silapb.SignedBlindedBeaconBlockElectra{}
	case version.Fulu:
		blk = &silapb.SignedBlindedBeaconBlockFulu{}
	case version.Gloas:
		blk = &silapb.SignedBeaconBlockGloas{}
	default:
		forkName := version.String(cf.Fork)
		return nil, fmt.Errorf("unable to initialize ReadOnlyBeaconBlock for fork version=%s at slot=%d", forkName, slot)
	}
	err = blk.UnmarshalSSZ(marshaled)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ReadOnlySignedBeaconBlock in UnmarshalSSZ")
	}
	return blocks.NewSignedBeaconBlock(blk)
}

// Heuristic to make sure block is from the same version as the VersionedUnmarshaler.
// Look up the version for the epoch that the block is from, then ensure that it matches the Version in the
// VersionedUnmarshaler.
func (cf *VersionedUnmarshaler) validateVersion(slot primitives.Slot) error {
	epoch := slots.ToEpoch(slot)
	fork, err := params.Fork(epoch)
	if err != nil {
		return err
	}
	ver := [4]byte(fork.CurrentVersion)
	if ver != cf.Version {
		return errors.Wrapf(errBlockForkMismatch, "slot=%d, epoch=%d, version=%#x", slot, epoch, fork.CurrentVersion)
	}
	return nil
}

func UnmarshalState(marshaled []byte) (state.BeaconState, error) {
	vu, err := FromState(marshaled)
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect version from state")
	}
	return vu.UnmarshalBeaconState(marshaled)
}
