package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mx-scribe/scribe/internal/version"
)

var showFullVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version number and build information of SCRIBE.`,
	Run: func(cmd *cobra.Command, args []string) {
		if showFullVersion {
			fmt.Println("SCRIBE", version.Full())
		} else {
			fmt.Println("SCRIBE", version.Info())
		}
	},
}

func init() {
	versionCmd.Flags().BoolVar(&showFullVersion, "full", false, "show full version with build info")
	rootCmd.AddCommand(versionCmd)
}
