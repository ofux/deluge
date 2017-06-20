package cmd

import (
	"github.com/spf13/cobra"
)

// startCmd represents the serve command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts Deluge as a worker or an orchestrator.",
	Long:  `Starts Deluge as a worker or an orchestrator.`,
}

func init() {
	RootCmd.AddCommand(startCmd)
}
