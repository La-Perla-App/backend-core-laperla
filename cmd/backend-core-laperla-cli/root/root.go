package root_cmd

import (
	"os"

	create_cmd "github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/create"
	init_cmd "github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/init"
	"github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/internal/config"
	"github.com/La-Perla-App/backend-core-laperla/cmd/protoc-gen-backend-core-laperla-cli/run"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use: "backend-core-laperla-cli",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cfgFile != "" {
			config.InitConfig(cfgFile)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fi, err := os.Stdin.Stat()
		if err != nil {
			cmd.Usage()
			return
		}

		// Comprobar si stdin es un pipe o no
		if fi.Mode()&os.ModeNamedPipe != 0 || fi.Size() > 0 {
			run.Run()
		} else {
			cmd.Usage()
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	config.InitConfig(cfgFile)

	create_cmd.Init(rootCmd)
	init_cmd.Init(rootCmd)
}
