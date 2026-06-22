package params

import (
	"bytes"
	"testing"

	fieldparams "github.com/OffchainLabs/prysm/v7/config/fieldparams"
)

func TestSilaPublicTestnetConfig(t *testing.T) {
	cfg := SilaPublicTestnetConfig()

	if cfg.ConfigName != SilaPublicTestnetName {
		t.Fatalf("unexpected config name: got %q want %q", cfg.ConfigName, SilaPublicTestnetName)
	}
	if cfg.DepositChainID != 20263001 {
		t.Fatalf("unexpected deposit chain id: got %d want %d", cfg.DepositChainID, uint64(20263001))
	}
	if cfg.DepositNetworkID != 20263001 {
		t.Fatalf("unexpected deposit network id: got %d want %d", cfg.DepositNetworkID, uint64(20263001))
	}
	if cfg.TerminalTotalDifficulty != "0" {
		t.Fatalf("unexpected terminal total difficulty: got %q want %q", cfg.TerminalTotalDifficulty, "0")
	}

	expected := [][]byte{
		{0x01, 0x35, 0x30, 0x59},
		{0x01, 0x35, 0x30, 0x5a},
		{0x01, 0x35, 0x30, 0x5b},
		{0x01, 0x35, 0x30, 0x5c},
		{0x01, 0x35, 0x30, 0x5d},
		{0x01, 0x35, 0x30, 0x5e},
		{0x01, 0x35, 0x30, 0x5f},
		{0x01, 0x35, 0x30, 0x60},
	}
	actual := [][]byte{
		cfg.GenesisForkVersion,
		cfg.AltairForkVersion,
		cfg.BellatrixForkVersion,
		cfg.CapellaForkVersion,
		cfg.DenebForkVersion,
		cfg.ElectraForkVersion,
		cfg.FuluForkVersion,
		cfg.GloasForkVersion,
	}

	for i := range expected {
		if !bytes.Equal(actual[i], expected[i]) {
			t.Fatalf("unexpected fork version at index %d: got %#x want %#x", i, actual[i], expected[i])
		}
	}

	mainnet := SilaMainnetConfig()
	if bytes.Equal(cfg.GenesisForkVersion, mainnet.GenesisForkVersion) {
		t.Fatal("Sila public testnet genesis fork version must not collide with Sila mainnet")
	}
}

func TestSilaPublicTestnetConfigByVersion(t *testing.T) {
	genesisVersion := [fieldparams.VersionLength]byte{0x01, 0x35, 0x30, 0x59}

	cfg, err := ByVersion(genesisVersion)
	if err != nil {
		t.Fatalf("ByVersion returned error: %v", err)
	}
	if cfg.ConfigName != SilaPublicTestnetName {
		t.Fatalf("ByVersion returned %q, want %q", cfg.ConfigName, SilaPublicTestnetName)
	}

	internalCfg, err := configs.byVersion(genesisVersion)
	if err != nil {
		t.Fatalf("configs.byVersion returned error: %v", err)
	}
	if internalCfg.ConfigName != SilaPublicTestnetName {
		t.Fatalf("configs.byVersion returned %q, want %q", internalCfg.ConfigName, SilaPublicTestnetName)
	}
}

func TestConfigset_SilaPublicTestnetRegistered(t *testing.T) {
	cfg, err := ByName(SilaPublicTestnetName)
	if err != nil {
		t.Fatalf("ByName(%q) returned error: %v", SilaPublicTestnetName, err)
	}
	if cfg.ConfigName != SilaPublicTestnetName {
		t.Fatalf("ByName returned %q, want %q", cfg.ConfigName, SilaPublicTestnetName)
	}
}
