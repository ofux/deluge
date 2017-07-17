package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ofux/deluge/api"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	runRemoteAddr string
)

// serveCmd represents the serve command
var runCmd = &cobra.Command{
	Use:   "run <file containing the deluge script> <output file>",
	Short: "Runs deluge script from the given file.",
	Long: `Runs deluge script from the given file.

If a worker/orchestrator address is given (see --remote flag) the script will be executed by this worker/orchestrator.
Otherwise, a local worker will be silently started on a random port to run the script and will be shutdown right after.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}
		// read input file
		fileContent, err := ioutil.ReadFile(args[0])
		if err != nil {
			die(err, 1)
		}
		// prepare output file
		fo, err := os.Create(args[1])
		if err != nil {
			die(err, 1)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		if runRemoteAddr == "" {
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				die(err, 1)
			}
			l.Close()
			randomPort := l.Addr().(*net.TCPAddr).Port
			runRemoteAddr = "http://localhost:" + strconv.Itoa(randomPort)
			go api.Serve(randomPort)
		}

		dlg := postDeluge(fileContent)

		// polling
		for dlg.Status == api.DelugeVirgin || dlg.Status == api.DelugeInProgress {
			dlg = getDeluge(dlg.ID)
			time.Sleep(500 * time.Millisecond)
		}

		result, err := json.Marshal(dlg)
		if err != nil {
			die(err, 1)
		}
		if _, err := fo.Write(result); err != nil {
			die(err, 1)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runRemoteAddr, "remote", "r", "", "The worker/orchestrator address on which the deluge script will be executed")

}

func die(err error, code int) {
	fmt.Println(err.Error())
	os.Exit(code)
}

func postDeluge(fileContent []byte) *api.Deluge {
	resp, err := http.Post(runRemoteAddr+"/v1/jobs", "text/plain", bytes.NewReader(fileContent))
	if err != nil {
		die(err, 2)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		die(err, 1)
	}
	if resp.StatusCode >= 300 {
		dtoErr := &api.Error{}
		if err = json.Unmarshal(body, dtoErr); err != nil {
			die(fmt.Errorf("Something went wrong. Received code %d from worker.", resp.StatusCode), 3)
		}
		die(fmt.Errorf("Something went wrong. Received code %d from worker with error: %s", resp.StatusCode, dtoErr.Error), 3)
	}

	dlg := &api.Deluge{}
	if err = json.Unmarshal(body, dlg); err != nil {
		die(err, 1)
	}
	return dlg
}

func getDeluge(id string) *api.Deluge {
	resp, err := http.Get(runRemoteAddr + "/v1/jobs/" + id)
	if err != nil {
		die(err, 2)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		die(err, 1)
	}
	if resp.StatusCode != http.StatusOK {
		dtoErr := &api.Error{}
		if err = json.Unmarshal(body, dtoErr); err != nil {
			die(fmt.Errorf("Something went wrong. Received code %d from worker.", resp.StatusCode), 3)
		}
		die(fmt.Errorf("Something went wrong. Received code %d from worker with error: %s", resp.StatusCode, dtoErr.Error), 3)
	}

	dlg := &api.Deluge{}
	if err = json.Unmarshal(body, dlg); err != nil {
		die(err, 1)
	}
	return dlg
}
