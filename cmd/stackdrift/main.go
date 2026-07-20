package main

import (
	"fmt"
	"os"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/commands"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
)

var version = "dev"

type command struct {
	run  func([]string) error
	help string
}

func main() {
	registry := map[string]command{
		"login":   {commands.Login, "sign in through the StackDrift website"},
		"logout":  {commands.Logout, "remove the saved credentials"},
		"whoami":  {commands.Whoami, "show the signed in account"},
		"scan":    {commands.Scan, "detect technologies and dependencies (add --yes to accept all)"},
		"remove":  {commands.Remove, "remove technologies or dependencies from this project"},
		"status":  {commands.Status, "show tracked technologies and dependencies"},
		"check":   {commands.Check, "report CVE status, exit non-zero if any are found"},
		"update":  {runUpdate, "download and install the latest release"},
		"version": {showVersion, "print the CLI version"},
	}

	if len(os.Args) < 2 {
		usage(registry)
		os.Exit(1)
	}

	name := os.Args[1]
	if name == "help" || name == "-h" || name == "--help" {
		usage(registry)
		return
	}

	cmd, ok := registry[name]
	if !ok {
		fmt.Fprintln(os.Stderr, "unknown command: "+name)
		usage(registry)
		os.Exit(1)
	}

	if err := cmd.run(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}
}

func showVersion([]string) error {
	fmt.Printf("stackdrift %s (server %s)\n", version, config.BaseURL())
	return nil
}

func runUpdate(args []string) error {
	return commands.Update(version, args)
}

func usage(registry map[string]command) {
	fmt.Println("StackDrift CLI")
	fmt.Println()
	fmt.Println("Usage: stackdrift <command>")
	fmt.Println()
	fmt.Println("Commands:")
	order := []string{"login", "scan", "status", "check", "remove", "whoami", "logout", "update", "version"}
	for _, name := range order {
		fmt.Printf("  %-9s %s\n", name, registry[name].help)
	}
	fmt.Println()
	fmt.Println("Set STACKDRIFT_URL to point at a different server.")
}
