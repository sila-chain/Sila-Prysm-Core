package kv

import (
	"context"
	"fmt"
	"slices"

	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	"github.com/sila-chain/Sila/common"
	bolt "go.etcd.io/bbolt"
)

// SilaDepositAddress returns contract address is the address of
// the sila deposit on the proof of work chain.
func (s *Store) SilaDepositAddress(ctx context.Context) ([]byte, error) {
	_, span := trace.StartSpan(ctx, "BeaconDB.SilaDepositAddress")
	defer span.End()
	var addr []byte
	if err := s.db.View(func(tx *bolt.Tx) error {
		chainInfo := tx.Bucket(chainMetadataBucket)
		stored := chainInfo.Get(silaDepositAddressKey)
		if len(stored) > 0 {
			addr = slices.Clone(stored)
		}
		return nil
	}); err != nil { // This view never returns an error, but we'll handle anyway for sanity.
		panic(err) // lint:nopanic -- View never returns an error.
	}
	return addr, nil
}

// SaveSilaDepositAddress to the db. It returns an error if an address has been previously saved.
func (s *Store) SaveSilaDepositAddress(ctx context.Context, addr common.Address) error {
	_, span := trace.StartSpan(ctx, "BeaconDB.VerifyContractAddress")
	defer span.End()

	return s.db.Update(func(tx *bolt.Tx) error {
		chainInfo := tx.Bucket(chainMetadataBucket)
		expectedAddress := chainInfo.Get(silaDepositAddressKey)
		if expectedAddress != nil {
			return fmt.Errorf("cannot override sila deposit address: %v", expectedAddress)
		}
		return chainInfo.Put(silaDepositAddressKey, addr.Bytes())
	})
}
