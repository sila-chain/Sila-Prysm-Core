package kv

import (
	"context"
	"errors"

	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	v2 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

// SaveExecutionChainData saves the execution chain data.
func (s *Store) SaveExecutionChainData(ctx context.Context, data *v2.SilaExecutionChainData) error {
	_, span := trace.StartSpan(ctx, "BeaconDB.SaveExecutionChainData")
	defer span.End()

	if data == nil {
		err := errors.New("cannot save nil silaData")
		tracing.AnnotateError(span, err)
		return err
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(powchainBucket)
		enc, err := proto.Marshal(data)
		if err != nil {
			return err
		}
		return bkt.Put(powchainDataKey, enc)
	})
	tracing.AnnotateError(span, err)
	return err
}

// ExecutionChainData retrieves the execution chain data.
func (s *Store) ExecutionChainData(ctx context.Context) (*v2.SilaExecutionChainData, error) {
	_, span := trace.StartSpan(ctx, "BeaconDB.ExecutionChainData")
	defer span.End()

	var data *v2.SilaExecutionChainData
	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(powchainBucket)
		enc := bkt.Get(powchainDataKey)
		if len(enc) == 0 {
			return nil
		}
		data = &v2.SilaExecutionChainData{}
		return proto.Unmarshal(enc, data)
	})
	return data, err
}
