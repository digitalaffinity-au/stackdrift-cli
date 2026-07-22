package main

import (
	"fmt"
	"io"
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
	os.Exit(run(os.Args[1:], registry(), os.Stdout, os.Stderr))
}

func registry() map[string]command {
	return map[string]command{
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
}

func run(args []string, registry map[string]command, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stdout, registry)
		return 1
	}

	name := args[0]
	if name == "help" || name == "-h" || name == "--help" {
		usage(stdout, registry)
		return 0
	}

	cmd, ok := registry[name]
	if !ok {
		fmt.Fprintln(stderr, "unknown command: "+name)
		usage(stdout, registry)
		return 1
	}

	if err := cmd.run(args[1:]); err != nil {
		fmt.Fprintln(stderr, "error: "+err.Error())
		return 1
	}
	return 0
}

func showVersion([]string) error {
	fmt.Printf("stackdrift %s (server %s)\n", version, config.BaseURL())
	return nil
}

func runUpdate(args []string) error {
	return commands.Update(version, args)
}

func usage(out io.Writer, registry map[string]command) {
	fmt.Fprintln(out, "StackDrift CLI")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage: stackdrift <command>")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Commands:")
	order := []string{"login", "scan", "status", "check", "remove", "whoami", "logout", "update", "version"}
	for _, name := range order {
		fmt.Fprintf(out, "  %-9s %s\n", name, registry[name].help)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Set STACKDRIFT_URL to point at a different server.")
}
