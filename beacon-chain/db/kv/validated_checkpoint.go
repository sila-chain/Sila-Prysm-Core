package kv

import (
	"bytes"
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	bolt "go.etcd.io/bbolt"
)

// LastValidatedCheckpoint returns the latest fully validated checkpoint in beacon chain.
func (s *Store) LastValidatedCheckpoint(ctx context.Context) (*silapb.Checkpoint, error) {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.LastValidatedCheckpoint")
	defer span.End()
	var checkpoint *silapb.Checkpoint
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(checkpointBucket)
		enc := bkt.Get(lastValidatedCheckpointKey)
		if enc == nil {
			var finErr error
			checkpoint, finErr = s.FinalizedCheckpoint(ctx)
			if finErr != nil {
				return finErr
			}
			if bytes.Equal(checkpoint.Root, params.BeaconConfig().ZeroHash[:]) {
				bkt = tx.Bucket(blocksBucket)
				r := bkt.Get(genesisBlockRootKey)
				if r != nil {
					checkpoint.Root = r
				}
			}
			return nil
		}
		checkpoint = &silapb.Checkpoint{}
		return decode(ctx, enc, checkpoint)
	})
	return checkpoint, err
}

// SaveLastValidatedCheckpoint saves the last validated checkpoint in beacon chain.
func (s *Store) SaveLastValidatedCheckpoint(ctx context.Context, checkpoint *silapb.Checkpoint) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.SaveLastValidatedCheckpoint")
	defer span.End()

	return s.saveCheckpoint(ctx, lastValidatedCheckpointKey, checkpoint)
}
