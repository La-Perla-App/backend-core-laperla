package create_cmd

import (
	create_handler_cmd "github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/create/handler"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "create",
	Short: "Commands for creating services handlers",
}

func Init(rootCmd *cobra.Command) {
	create_handler_cmd.Init(cmd)

	rootCmd.AddCommand(cmd)
}
