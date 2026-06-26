package silaexec

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/cmd/beacon-chain/flags"
	"github.com/sila-chain/Sila-Consensus-Core/v7/io/file"
	"github.com/urfave/cli/v2"
)

// FlagOptions for execution service flag configurations.
func FlagOptions(c *cli.Context) ([]silaexec.Option, error) {
	endpoint, err := parseSilaChainEndpoint(c)
	if err != nil {
		return nil, err
	}
	jwtSecret, err := parseJWTSecretFromFile(c)
	if err != nil {
		return nil, errors.Wrap(err, "could not read JWT secret file for authenticating execution API")
	}
	headers := strings.Split(c.String(flags.ExecutionEngineHeaders.Name), ",")
	opts := []silaexec.Option{
		silaexec.WithSilaHeaderRequestLimit(c.Uint64(flags.SilaHeaderReqLimit.Name)),
		silaexec.WithHeaders(headers),
	}
	if len(jwtSecret) > 0 {
		opts = append(opts, silaexec.WithHttpEndpointAndJWTSecret(endpoint, jwtSecret))
	} else {
		opts = append(opts, silaexec.WithHttpEndpoint(endpoint))
	}
	return opts, nil
}

// Parses a JWT secret from a file path. This secret is required when connecting to sila nodes
// over HTTP, and must be the same one used in Sila and the Sila node server Sila is connecting to.
// The SilaEngine API specification here https://github.com/sila-chain/Sila-Execution-APIs/blob/main/src/engine/authentication.md
// Explains how we should validate this secret and the format of the file a user can specify.
//
// The secret must be stored as a hex-encoded string within a file in the filesystem.
// If the --jwt-secret flag is provided to Sila, but the file cannot be read, or does not contain a hex-encoded
// key of 256 bits, the client should treat this as an error and abort the startup.
func parseJWTSecretFromFile(c *cli.Context) ([]byte, error) {
	jwtSecretFile := c.String(flags.ExecutionJWTSecretFlag.Name)
	if jwtSecretFile == "" {
		return nil, nil
	}
	enc, err := file.ReadFileAsBytes(jwtSecretFile)
	if err != nil {
		return nil, err
	}
	strData := strings.TrimSpace(string(enc))
	if strData == "" {
		return nil, fmt.Errorf("provided JWT secret in file %s cannot be empty", jwtSecretFile)
	}
	secret, err := hex.DecodeString(strings.TrimPrefix(strData, "0x"))
	if err != nil {
		return nil, err
	}
	if len(secret) != 32 {
		return nil, errors.New("provided JWT secret should be a hex string of 32 bytes")
	}
	log.Infof("Finished reading JWT secret from %s", jwtSecretFile)
	return secret, nil
}

func parseSilaChainEndpoint(c *cli.Context) (string, error) {
	if c.String(flags.ExecutionEngineEndpoint.Name) == "" {
		return "", fmt.Errorf(
			"you need to specify %s to provide a connection endpoint to a Sila client "+
				"for your Sila beacon node. This is a requirement for running a node. You can read more about "+
				"how to configure this Sila client connection in our docs here "+
				"https://docs.prylabs.network/docs/install/install-with-script",
			flags.ExecutionEngineEndpoint.Name,
		)
	}
	return c.String(flags.ExecutionEngineEndpoint.Name), nil
}
