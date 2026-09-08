package config

import (
	_ "embed"
	"os"

	"dario.cat/mergo"
	"github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/internal"
	"github.com/spf13/cobra"
)

type CliConfigData struct{}

var CliConfig = CliConfigData{}

func InitConfig(cfgFile string) {
	lookupExts := []string{
		internal.FileExtensions.Json,
		internal.FileExtensions.Yml,
		internal.FileExtensions.Yaml,
	}

	if cfgFile == "" {
		cwd, err := os.Getwd()
		cobra.CheckErr(err)
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		var homeCfg CliConfigData
		if _, err := internal.LookupAndUnmarshalFile(os.DirFS(home), "", ".backend-core-laperla-cli.config", lookupExts, &homeCfg); err != nil {
			cobra.CheckErr(err)
		}

		var cwdCfg CliConfigData
		if _, err := internal.LookupAndUnmarshalFile(os.DirFS(cwd), "", "backend-core-laperla-cli.config", lookupExts, &cwdCfg); err != nil {
			cobra.CheckErr(err)
		}

		mergo.Merge(&CliConfig, homeCfg, mergo.WithOverride)
		mergo.Merge(&CliConfig, cwdCfg, mergo.WithOverride)
	} else {
		var cfg CliConfigData
		if _, err := internal.UnmarshalFile(cfgFile, lookupExts, &cfg); err != nil {
			cobra.CheckErr(err)
		}
		mergo.Merge(&CliConfig, cfg, mergo.WithOverride)
	}
}
