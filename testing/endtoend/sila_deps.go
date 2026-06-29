package endtoend

// This file contains the dependencies required for the Sila execution client.
// Having these dependencies listed here helps go mod understand that these dependencies are
// necessary for end to end tests since we build Sila binary for this test.
import (
	_ "github.com/sila-chain/Sila/accounts"          // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/accounts/keystore" // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/cmd/utils"         // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/common"            // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/console"           // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/log"               // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/metrics"           // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/node"              // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/sila"              // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/sila/downloader"   // Required for Sila e2e.
	_ "github.com/sila-chain/Sila/silaclient"        // Required for Sila e2e.
)
