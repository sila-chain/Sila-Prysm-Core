package stateutil

import (
	"bytes"
	"encoding/binary"

	params "github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// SilaDataRootWithHasher returns the hash tree root of input `silaexecData`.
func SilaDataRootWithHasher(silaexecData *silapb.SilaData) ([32]byte, error) {
	if silaexecData == nil {
		return [32]byte{}, errors.New("nil silaexec data")
	}

	fieldRoots := make([][32]byte, 3)
	for i := range fieldRoots {
		fieldRoots[i] = [32]byte{}
	}

	if len(silaexecData.DepositRoot) > 0 {
		fieldRoots[0] = bytesutil.ToBytes32(silaexecData.DepositRoot)
	}

	silaexecDataCountBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(silaexecDataCountBuf, silaexecData.DepositCount)
	fieldRoots[1] = bytesutil.ToBytes32(silaexecDataCountBuf)
	if len(silaexecData.BlockHash) > 0 {
		fieldRoots[2] = bytesutil.ToBytes32(silaexecData.BlockHash)
	}
	root, err := ssz.BitwiseMerkleize(fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
	if err != nil {
		return [32]byte{}, err
	}
	return root, nil
}

// SilaDatasRoot returns the hash tree root of input `silaexecDatas`.
func SilaDatasRoot(silaexecDatas []*silapb.SilaData) ([32]byte, error) {
	silaexecVotesRoots := make([][32]byte, 0, len(silaexecDatas))
	for i := range silaexecDatas {
		silaexec, err := SilaDataRootWithHasher(silaexecDatas[i])
		if err != nil {
			return [32]byte{}, errors.Wrap(err, "could not compute silaData merkleization")
		}
		silaexecVotesRoots = append(silaexecVotesRoots, silaexec)
	}

	silaexecVotesRootsRoot, err := ssz.BitwiseMerkleize(silaexecVotesRoots, uint64(len(silaexecVotesRoots)), params.BeaconConfig().SilaDataVotesLength())
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not compute silaData votes merkleization")
	}
	silaexecVotesRootBuf := new(bytes.Buffer)
	if err := binary.Write(silaexecVotesRootBuf, binary.LittleEndian, uint64(len(silaexecDatas))); err != nil {
		return [32]byte{}, errors.Wrap(err, "could not marshal silaData votes length")
	}
	// We need to mix in the length of the slice.
	silaexecVotesRootBufRoot := make([]byte, 32)
	copy(silaexecVotesRootBufRoot, silaexecVotesRootBuf.Bytes())
	root := ssz.MixInLength(silaexecVotesRootsRoot, silaexecVotesRootBufRoot)

	return root, nil
}
