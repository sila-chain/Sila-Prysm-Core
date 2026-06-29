package slashings

import silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"

// IsSurround checks if an attestation, a, is surrounding
// another one, b, based on the Sila slashing conditions specified
// by @sila-chain https://github.com/sila-chain/silaconsensus-surround#definition.
//
//	s: source
//	t: target
//
//	a surrounds b if: s_a < s_b and t_b < t_a
func IsSurround(a, b silapb.IndexedAtt) bool {
	return a.GetData().Source.Epoch < b.GetData().Source.Epoch && b.GetData().Target.Epoch < a.GetData().Target.Epoch
}
