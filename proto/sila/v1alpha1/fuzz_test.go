package v1alpha1_test
import (
	"fmt"
	"testing"

	fuzz "github.com/google/gofuzz"
	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func fuzzCopies[T any, C sila.Copier[T]](t *testing.T, obj C) {
	fuzzer := fuzz.NewWithSeed(0)
	amount := 1000
	t.Run(fmt.Sprintf("%T", obj), func(t *testing.T) {
		for range amount {
			fuzzer.Fuzz(obj) // Populate thing with random values

			got := obj.Copy()
			require.DeepEqual(t, obj, got)
			// TODO: add deep fuzzing and checks for deep not equals
			// we should test that modifying the copy doesn't modify the original object
		}
	})
}
