package cmd

import (
	"github.com/ofux/deluge/api"
	"github.com/spf13/cobra"
)

var (
	workerPort int32
)

// serveCmd represents the serve command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Starts a server that will execute some jobs such as running scenarios.",
	Long: `A worker is a Deluge server instance that is meant to execute some jobs like running scenarios and generating reports.
It can be used together with an orchestrator.`,
	Run: func(cmd *cobra.Command, args []string) {
		api.Serve(int(workerPort))
	},
}

func init() {
	startCmd.AddCommand(workerCmd)

	workerCmd.Flags().Int32VarP(&workerPort, "port", "p", 33033, "The port on which deluge worker will be listening")

}
