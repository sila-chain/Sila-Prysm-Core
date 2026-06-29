package wallet

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/cmd"
	"github.com/sila-chain/Sila-Consensus-Core/v7/cmd/validator/flags"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/tos"
	"github.com/urfave/cli/v2"
)

// Commands for wallets for Sila validators.
var Commands = &cli.Command{
	Name:     "wallet",
	Category: "wallet",
	Usage:    "Defines commands for interacting with validator wallets.",
	Subcommands: []*cli.Command{
		{
			Name: "create",
			Usage: "creates a new wallet with a desired type of keymanager: " +
				"either on-disk (imported), derived, or using remote credentials",
			Flags: cmd.WrapFlags([]cli.Flag{
				flags.WalletDirFlag,
				flags.KeymanagerKindFlag,
				flags.RemoteSignerCertPathFlag,
				flags.RemoteSignerKeyPathFlag,
				flags.RemoteSignerCACertPathFlag,
				flags.WalletPasswordFileFlag,
				flags.Mnemonic25thWordFileFlag,
				flags.SkipMnemonic25thWordCheckFlag,
				features.LegacyMainNetwork,
				features.SilaCompatTestnet,
				features.SilaValidatorScaleTestnet,
				features.HoodiTestnet,
				cmd.AcceptTosFlag,
			}),
			Before: func(cliCtx *cli.Context) error {
				if err := cmd.LoadFlagsFromConfig(cliCtx, cliCtx.Command.Flags); err != nil {
					return err
				}
				if err := tos.VerifyTosAcceptedOrPrompt(cliCtx); err != nil {
					return err
				}
				return features.ConfigureValidator(cliCtx)
			},
			Action: func(cliCtx *cli.Context) error {
				if err := walletCreate(cliCtx); err != nil {
					log.WithError(err).Fatal("Could not create a wallet")
				}
				return nil
			},
		},
		{
			Name:  "recover",
			Usage: "uses a derived wallet seed recovery phase to recreate an existing HD wallet",
			Flags: cmd.WrapFlags([]cli.Flag{
				flags.WalletDirFlag,
				flags.MnemonicFileFlag,
				flags.WalletPasswordFileFlag,
				flags.NumAccountsFlag,
				flags.Mnemonic25thWordFileFlag,
				flags.SkipMnemonic25thWordCheckFlag,
				features.LegacyMainNetwork,
				features.SilaCompatTestnet,
				features.SilaValidatorScaleTestnet,
				features.HoodiTestnet,
				cmd.AcceptTosFlag,
			}),
			Before: func(cliCtx *cli.Context) error {
				if err := cmd.LoadFlagsFromConfig(cliCtx, cliCtx.Command.Flags); err != nil {
					return err
				}
				if err := tos.VerifyTosAcceptedOrPrompt(cliCtx); err != nil {
					return err
				}
				return features.ConfigureBeaconChain(cliCtx)
			},
			Action: func(cliCtx *cli.Context) error {
				if err := walletRecover(cliCtx); err != nil {
					log.WithError(err).Fatal("Could not recover wallet")
				}
				return nil
			},
		},
	},
}
