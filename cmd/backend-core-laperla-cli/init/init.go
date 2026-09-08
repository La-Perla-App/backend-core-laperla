package init_cmd

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/internal"
	"github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli/internal/assets"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

// pagesCmd represents the pages command
var cmd = &cobra.Command{
	Use:   "init [packageName]",
	Short: "Init a new microservice",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")
		pkg, _ := cmd.Flags().GetString("package")
		hasGoMod := internal.FileExists(path.Join(output, "go.mod"))
		if pkg == "" && hasGoMod {
			var err error
			pkg, err = internal.ParseGoMod(output)
			cobra.CheckErr(err)
		}
		if pkg == "" {
			cobra.CheckErr(errors.New("go.mod file not found and --package flag is not supplied"))
		}

		pkgShort := strings.TrimSuffix(path.Base(pkg), ".git")
		pkgShort = strings.TrimSuffix(pkgShort, "-api")
		pkgShort = strings.TrimSuffix(pkgShort, "-backend")
		pkgShort = strings.TrimSuffix(pkgShort, "-go")
		pkgShort = strings.TrimPrefix(pkgShort, "laperla-")

		pathPrefix, _ := cmd.Flags().GetString("api-prefix")
		if pathPrefix == "" {
			pathPrefix = path.Join("api", pkgShort)
		}
		if !strings.HasPrefix(pathPrefix, "/") {
			pathPrefix = "/" + strings.TrimSpace(pathPrefix)
		}

		r := strings.NewReplacer(
			"__package__", pkg,
			"__package_short__", pkgShort,
			"__api_name__", strcase.ToCamel(pkgShort),
			"__api_prefix__", pathPrefix,
		)

		cobra.CheckErr(internal.CreateFilesFromFS(assets.TemplateFS, output, force, r))

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

func Init(rootCmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "", "output folder")
	cmd.Flags().String("api-prefix", "", "API REST path prefix")
	cmd.Flags().StringP("package", "p", "", "go package name")
	cmd.Flags().BoolP("force", "f", false, "replace files if exists")

	rootCmd.AddCommand(cmd)
}
