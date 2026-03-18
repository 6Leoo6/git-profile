package main

import (
	"fmt"
	"os"
	"strings"
)

const version = "1.0.0"

func main() {
	enableANSI()

	cfg, err := loadConfig()
	if err != nil {
		fatal("loading config: %v", err)
	}

	args := os.Args[1:]

	if len(args) == 0 {
		showStatus(cfg)
		return
	}

	switch args[0] {
	case "--help", "-h", "help":
		showHelp(cfg)
	case "--version", "-v", "version":
		fmt.Printf("git-profile %s\n", version)
	case "add":
		runAdd(cfg, args[1:])
	case "remove", "rm":
		runRemove(cfg, args[1:])
	default:
		runSwitch(cfg, args)
	}
}

// runAdd validates args and saves a new or updated profile.
func runAdd(cfg Config, args []string) {
	if len(args) != 3 {
		fatal("usage: git-profile add <profile> \"Full Name\" <email>")
	}
	profileName, name, email := args[0], args[1], args[2]
	if err := validateEmail(email); err != nil {
		fatal("%v", err)
	}
	saveProfile(cfg, profileName, name, email)
}

// runRemove deletes a profile from the config file.
func runRemove(cfg Config, args []string) {
	if len(args) != 1 {
		fatal("usage: git-profile remove <profile>")
	}
	deleteProfile(cfg, args[0])
}

// runSwitch applies a profile locally (default) or globally.
func runSwitch(cfg Config, args []string) {
	useGlobal := len(args) >= 2 && args[1] == "--global"
	applyProfile(cfg, args[0], !useGlobal)
}

// applyProfile writes user.name and user.email via git config.
func applyProfile(cfg Config, profileName string, local bool) {
	p, ok := cfg.Profiles[profileName]
	if !ok {
		fatal("unknown profile %q\n\n    available: %s", profileName, strings.Join(sortedKeys(cfg), ", "))
	}

	scope := "--global"
	scopeLabel := "global"
	if local {
		if !insideRepo() {
			fatal("not inside a Git repository — use --global to apply globally")
		}
		scope = "--local"
		scopeLabel = "this repo"
	}

	if err := gitSet(scope, "user.name", p.Name); err != nil {
		fatal("setting user.name: %v", err)
	}
	if err := gitSet(scope, "user.email", p.Email); err != nil {
		fatal("setting user.email: %v", err)
	}

	printSuccess("switched to %s %s", yellow(profileName), dim("("+scopeLabel+")"))
	fmt.Printf("         %s <%s>\n\n", p.Name, p.Email)
}

func fatal(format string, a ...any) {
	printError(format, a...)
	os.Exit(1)
}
