package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var serverURL string

var rootCmd = &cobra.Command{
	Use:   "dns-client",
	Short: "CLI client for remote DNS management",
	Long:  `Command-line interface for managing DNS nameservers on a remote server.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured nameservers",
	Run:   listNameservers,
}

var addCmd = &cobra.Command{
	Use:   "add [ip-address]",
	Short: "Add a new nameserver",
	Args:  cobra.ExactArgs(1),
	Run:   addNameserver,
}

var removeCmd = &cobra.Command{
	Use:   "remove [ip-address]",
	Short: "Remove a nameserver",
	Args:  cobra.ExactArgs(1),
	Run:   removeNameserver,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "DNS Manager server URL")
	rootCmd.AddCommand(listCmd, addCmd, removeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func listNameservers(cmd *cobra.Command, args []string) {
	resp, err := http.Get(serverURL + "/dns")
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(body))
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

func addNameserver(cmd *cobra.Command, args []string) {
	sendRequest("POST", args[0])
}

func removeNameserver(cmd *cobra.Command, args []string) {
	sendRequest("DELETE", args[0])
}

func sendRequest(method, ip string) {
	data := map[string]string{"nameserver": ip}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	var resp *http.Response
	var reqErr error

	if method == "POST" {
		resp, reqErr = http.Post(serverURL+"/dns", "application/json", bytes.NewBuffer(jsonData))
	} else {
		req, err := http.NewRequest("DELETE", serverURL+"/dns", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, reqErr = client.Do(req)
	}

	if reqErr != nil {
		fmt.Printf("Request error: %v\n", reqErr)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
