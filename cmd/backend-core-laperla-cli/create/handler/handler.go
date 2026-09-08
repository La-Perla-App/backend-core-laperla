package create_handler_cmd

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/internal"
	"github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/internal/assets"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "handler [service name]",
	Short: "Create a new handler",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := strcase.ToCamel(args[0])
		output, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")
		pkg, _ := cmd.Flags().GetString("package")
		version, _ := cmd.Flags().GetInt("version")
		if version == 0 {
			cobra.CheckErr(errors.New("the --version flag can not be empty or zero"))
		}
		hasGoMod := internal.FileExists(path.Join(output, "go.mod"))
		if pkg == "" && hasGoMod {
			var err error
			pkg, err = internal.ParseGoMod(output)
			cobra.CheckErr(err)
		}
		if pkg == "" {
			cobra.CheckErr(errors.New("go.mod file not found and --package flag is not supplied"))
		}
		servicePkg := genPkgName(name)
		versionStr := fmt.Sprintf("v%v", version)

		handlerBasePath, _ := cmd.Flags().GetString("handlers-path")
		handlerPath := path.Join(output, handlerBasePath, servicePkg, versionStr)

		protoBasePath, _ := cmd.Flags().GetString("proto-path")
		protoPath := path.Join(output, protoBasePath, "services", servicePkg, versionStr)

		r := strings.NewReplacer(
			"__package__", pkg,
			"__pkg__", servicePkg,
			"__handler_pkg__", servicePkg+versionStr,
			"__proto_path__", protoBasePath,
			"__handlers_path__", handlerBasePath,
			"__service__", name,
			"__version__", versionStr,
		)

		cobra.CheckErr(internal.CreateFilesFromFS(assets.HandlerFS, handlerPath, force, r))
		cobra.CheckErr(internal.CreateFilesFromFS(assets.ProtoFS, protoPath, force, r))

		fmt.Print("Done! Run the next commands:\n\n")
		if output != "" && output != "." {
			fmt.Printf("cd %s\n", output)
		}
		if !hasGoMod {
			fmt.Printf("go mod init %s\n", pkg)
		}
		fmt.Println("buf dep update")
		fmt.Println("buf generate")
		fmt.Printf("go mod tidy\n\n")
	},
}

func genPkgName(key string) string {
	key = regexp.MustCompile(`^\d+`).ReplaceAllString(key, "")
	return strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(key, ""))
}

func Init(rootCmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "", "output folder")
	cmd.Flags().String("handlers-path", "handlers", "handlers folder")
	cmd.Flags().String("proto-path", "proto", "protocolbuffers folder")
	cmd.Flags().StringP("package", "p", "", "go package name")
	cmd.Flags().BoolP("force", "f", false, "replace files if exists")
	cmd.Flags().IntP("version", "v", 0, "specifies the handler version")

	rootCmd.AddCommand(cmd)
}
