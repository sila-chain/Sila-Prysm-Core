package silaenginev1_test

import (
	"fmt"
	"testing"

	fuzz "github.com/google/gofuzz"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestCopySilaPayload_Fuzz(t *testing.T) {
	fuzzCopies(t, &silaenginev1.SilaPayloadDeneb{})
	fuzzCopies(t, &silaenginev1.SilaPayloadCapella{})
	fuzzCopies(t, &silaenginev1.SilaPayload{})
}

func TestCopySilaPayloadHeader_Fuzz(t *testing.T) {
	fuzzCopies(t, &silaenginev1.SilaPayloadHeaderDeneb{})
	fuzzCopies(t, &silaenginev1.SilaPayloadHeaderCapella{})
	fuzzCopies(t, &silaenginev1.SilaPayloadHeader{})
}

func fuzzCopies[T any, C silaenginev1.Copier[T]](t *testing.T, obj C) {
	fuzzer := fuzz.NewWithSeed(0)
	amount := 1000
	t.Run(fmt.Sprintf("%T", obj), func(t *testing.T) {
		for range amount {
			fuzzer.Fuzz(obj) // Populate thing with random values
			got := obj.Copy()
			require.DeepEqual(t, obj, got)
			// check shallow copy working
			fuzzer.Fuzz(got)
			require.DeepNotEqual(t, obj, got)
			// TODO: think of deeper not equal fuzzing
		}
	})
}
