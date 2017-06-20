package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	orchestratorPort int32
)

// serveCmd represents the serve command
var orchestratorCmd = &cobra.Command{
	Use:   "orchestrator",
	Short: "(NOT YET IMPLEMENTED) Starts a server that will spread jobs across multiple workers.",
	Long: `(NOT YET IMPLEMENTED) An orchestrator is a Deluge server instance that is meant to spread jobs across multiple workers.

First, start the orchestrator.
Then, start each worker with the --orchestrator flag so that they will join the orchestrator and will act as slaves for it.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This feature is not yet implemented.")
	},
}

func init() {
	startCmd.AddCommand(orchestratorCmd)

	orchestratorCmd.Flags().Int32VarP(&orchestratorPort, "port", "p", 33044, "The port on which deluge orchestrator will be listening")

}
