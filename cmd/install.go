package cmd

import (
	install "github.com/Sabique-Islam/catalyst/internal/install"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return install.InstallDependencies()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
