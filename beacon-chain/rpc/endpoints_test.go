package rpc

import (
	"maps"
	"net/http"
	"slices"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
)

func Test_endpoints(t *testing.T) {
	rewardsRoutes := map[string][]string{
		"/sila/v1/beacon/rewards/blocks/{block_id}":         {http.MethodGet},
		"/sila/v1/beacon/rewards/attestations/{epoch}":      {http.MethodPost},
		"/sila/v1/beacon/rewards/sync_committee/{block_id}": {http.MethodPost},
	}

	beaconRoutes := map[string][]string{
		"/sila/v1/beacon/genesis":                                       {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/root":                        {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/fork":                        {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/finality_checkpoints":        {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/validators":                  {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/states/{state_id}/validators/{validator_id}":   {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/validator_balances":          {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/states/{state_id}/validator_identities":        {http.MethodPost},
		"/sila/v1/beacon/states/{state_id}/committees":                  {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/sync_committees":             {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/randao":                      {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/pending_deposits":            {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/pending_partial_withdrawals": {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/pending_consolidations":      {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/proposer_lookahead":          {http.MethodGet},
		"/sila/v1/beacon/execution_payload_envelopes/{block_id}":        {http.MethodGet},
		"/sila/v1/beacon/execution_payload_envelopes":                   {http.MethodPost},
		"/sila/v1/beacon/execution_payload_bids":                        {http.MethodPost},
		"/sila/v1/beacon/headers":                                       {http.MethodGet},
		"/sila/v1/beacon/headers/{block_id}":                            {http.MethodGet},
		"/sila/v2/beacon/blinded_blocks":                                {http.MethodPost},
		"/sila/v2/beacon/blocks":                                        {http.MethodPost},
		"/sila/v2/beacon/blocks/{block_id}":                             {http.MethodGet},
		"/sila/v1/beacon/blocks/{block_id}/root":                        {http.MethodGet},
		"/sila/v2/beacon/blocks/{block_id}/attestations":                {http.MethodGet},
		"/sila/v1/beacon/blob_sidecars/{block_id}":                      {http.MethodGet},
		"/sila/v1/beacon/blinded_blocks/{block_id}":                     {http.MethodGet},
		"/sila/v2/beacon/pool/attestations":                             {http.MethodGet, http.MethodPost},
		"/sila/v2/beacon/pool/attester_slashings":                       {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/pool/proposer_slashings":                       {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/pool/sync_committees":                          {http.MethodPost},
		"/sila/v1/beacon/pool/voluntary_exits":                          {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/pool/bls_to_execution_changes":                 {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/pool/payload_attestations":                     {http.MethodGet, http.MethodPost},
		"/sila/v1/beacon/individual_votes":                             {http.MethodPost},
	}

	lightClientRoutes := map[string][]string{
		"/sila/v1/beacon/light_client/bootstrap/{block_root}": {http.MethodGet},
		"/sila/v1/beacon/light_client/updates":                {http.MethodGet},
		"/sila/v1/beacon/light_client/finality_update":        {http.MethodGet},
		"/sila/v1/beacon/light_client/optimistic_update":      {http.MethodGet},
	}

	blobRoutes := map[string][]string{
		"/sila/v1/beacon/blob_sidecars/{block_id}": {http.MethodGet},
		"/sila/v1/beacon/blobs/{block_id}":         {http.MethodGet},
	}

	configRoutes := map[string][]string{
		"/sila/v1/config/fork_schedule":    {http.MethodGet},
		"/sila/v1/config/spec":             {http.MethodGet},
		"/sila/v1/config/deposit_contract": {http.MethodGet},
	}

	debugRoutes := map[string][]string{
		"/sila/v2/debug/beacon/states/{state_id}":               {http.MethodGet},
		"/sila/v2/debug/beacon/heads":                           {http.MethodGet},
		"/sila/v1/debug/fork_choice":                            {http.MethodGet},
		"/sila/v2/debug/fork_choice":                            {http.MethodGet},
		"/sila/v1/debug/beacon/data_column_sidecars/{block_id}": {http.MethodGet},
	}

	eventsRoutes := map[string][]string{
		"/sila/v1/events": {http.MethodGet},
	}

	nodeRoutes := map[string][]string{
		"/sila/v1/node/identity":        {http.MethodGet},
		"/sila/v1/node/peers":           {http.MethodGet},
		"/sila/v1/node/peers/{peer_id}": {http.MethodGet},
		"/sila/v1/node/peer_count":      {http.MethodGet},
		"/sila/v1/node/version":         {http.MethodGet},
		"/sila/v2/node/version":         {http.MethodGet},
		"/sila/v1/node/syncing":         {http.MethodGet},
		"/sila/v1/node/health":          {http.MethodGet},
	}

	validatorRoutes := map[string][]string{
		"/sila/v1/validator/duties/attester/{epoch}":                                {http.MethodPost},
		"/sila/v1/validator/duties/proposer/{epoch}":                                {http.MethodGet},
		"/sila/v2/validator/duties/proposer/{epoch}":                                {http.MethodGet},
		"/sila/v1/validator/duties/sync/{epoch}":                                    {http.MethodPost},
		"/sila/v1/validator/duties/ptc/{epoch}":                                     {http.MethodPost},
		"/sila/v3/validator/blocks/{slot}":                                          {http.MethodGet},
		"/sila/v4/validator/blocks/{slot}":                                          {http.MethodGet},
		"/sila/v1/validator/attestation_data":                                       {http.MethodGet},
		"/sila/v2/validator/aggregate_attestation":                                  {http.MethodGet},
		"/sila/v2/validator/aggregate_and_proofs":                                   {http.MethodPost},
		"/sila/v1/validator/beacon_committee_subscriptions":                         {http.MethodPost},
		"/sila/v1/validator/sync_committee_subscriptions":                           {http.MethodPost},
		"/sila/v1/validator/beacon_committee_selections":                            {http.MethodPost},
		"/sila/v1/validator/sync_committee_selections":                              {http.MethodPost},
		"/sila/v1/validator/execution_payload_envelopes/{slot}/{beacon_block_root}": {http.MethodGet},
		"/sila/v1/validator/sync_committee_contribution":                            {http.MethodGet},
		"/sila/v1/validator/contribution_and_proofs":                                {http.MethodPost},
		"/sila/v1/validator/prepare_beacon_proposer":                                {http.MethodPost},
		"/sila/v1/validator/proposer_preferences":                                   {http.MethodPost},
		"/sila/v1/validator/register_validator":                                     {http.MethodPost},
		"/sila/v1/validator/liveness/{epoch}":                                       {http.MethodPost},
		"/sila/v1/validator/payload_attestation_data/{slot}":                        {http.MethodGet},
	}

	silaBeaconRoutes := map[string][]string{
		"/sila/v1/beacon/weak_subjectivity":                 {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/validator_count":  {http.MethodGet},
		"/sila/v1/beacon/states/{state_id}/validator_count": {http.MethodGet},
		"/sila/v1/beacon/chain_head":                        {http.MethodGet},
		"/sila/v1/beacon/blobs":                             {http.MethodPost},
		"/sila/v1/beacon/states/{state_id}/query":           {http.MethodPost},
		"/sila/v1/beacon/blocks/{block_id}/query":           {http.MethodPost},
	}

	silaNodeRoutes := map[string][]string{
		"/sila/node/trusted_peers":              {http.MethodGet, http.MethodPost},
		"/sila/v1/node/trusted_peers":           {http.MethodGet, http.MethodPost},
		"/sila/node/trusted_peers/{peer_id}":    {http.MethodDelete},
		"/sila/v1/node/trusted_peers/{peer_id}": {http.MethodDelete},
	}

	silaValidatorRoutes := map[string][]string{
		"/sila/validators/performance":                      {http.MethodPost},
		"/sila/v1/validators/performance":                   {http.MethodPost},
		"/sila/v1/validators/{state_id}/participation":      {http.MethodGet},
		"/sila/v1/validators/{state_id}/active_set_changes": {http.MethodGet},
	}

	testCases := []struct {
		name                     string
		flag                     *features.Flags
		additionalExpectedRoutes []map[string][]string
	}{
		{
			name: "no flags",
		},
		{
			name: "light client enabled",
			flag: &features.Flags{
				EnableLightClient: true,
			},
			additionalExpectedRoutes: []map[string][]string{
				lightClientRoutes,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetFn := features.InitWithReset(tc.flag)
			defer resetFn()

			s := &Service{cfg: &Config{}}

			endpoints := s.endpoints(true, nil, nil, nil, nil, nil, nil)
			actualRoutes := make(map[string][]string, len(endpoints))
			for _, e := range endpoints {
				if _, ok := actualRoutes[e.template]; ok {
					actualRoutes[e.template] = append(actualRoutes[e.template], e.methods...)
				} else {
					actualRoutes[e.template] = e.methods
				}
			}
			expectedRoutes := make(map[string][]string)
			for _, m := range []map[string][]string{
				beaconRoutes, configRoutes, debugRoutes, eventsRoutes,
				nodeRoutes, validatorRoutes, rewardsRoutes, blobRoutes,
				silaValidatorRoutes, silaNodeRoutes, silaBeaconRoutes,
			} {
				maps.Copy(expectedRoutes, m)
			}
			for _, m := range tc.additionalExpectedRoutes {
				maps.Copy(expectedRoutes, m)
			}
			maps.Copy(expectedRoutes, ethRouteAliases(expectedRoutes))

			assert.Equal(t, true, maps.EqualFunc(expectedRoutes, actualRoutes, func(actualMethods []string, expectedMethods []string) bool {
				return slices.Equal(expectedMethods, actualMethods)
			}))
		})
	}
}
func ethRouteAliases(routes map[string][]string) map[string][]string {
	aliases := make(map[string][]string)
	for route, methods := range routes {
		if len(route) >= len("/sila/") && route[:len("/sila/")] == "/sila/" {
			aliases["/silaapi/"+route[len("/sila/"):]] = methods
		}
	}
	return aliases
}
