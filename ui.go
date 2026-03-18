package main

import (
	"fmt"
	"os"
)

// ── Color helpers ─────────────────────────────────────────────────────────────

func bold(s string) string   { return "\033[1m" + s + "\033[0m" }
func dim(s string) string    { return "\033[2m" + s + "\033[0m" }
func green(s string) string  { return "\033[32m" + s + "\033[0m" }
func yellow(s string) string { return "\033[33m" + s + "\033[0m" }
func cyan(s string) string   { return "\033[36m" + s + "\033[0m" }
func red(s string) string    { return "\033[31m" + s + "\033[0m" }

// ── Print helpers ─────────────────────────────────────────────────────────────

func printError(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "\n  "+red("[x]")+" "+format+"\n\n", a...)
}

func printWarning(format string, a ...any) {
	fmt.Printf("  "+yellow("[!]")+" "+format+"\n", a...)
}

func printSuccess(format string, a ...any) {
	fmt.Printf("\n  "+green("[ok]")+" "+format+"\n", a...)
}

// ── Commands ──────────────────────────────────────────────────────────────────

func showStatus(cfg Config) {
	globalName := gitGet("config", "--global", "user.name")
	globalEmail := gitGet("config", "--global", "user.email")

	fmt.Println()
	fmt.Printf("  %s\n", bold("Profiles"))
	for _, key := range sortedKeys(cfg) {
		p := cfg.Profiles[key]
		fmt.Printf("    %-16s %s <%s>\n", yellow(key), p.Name, p.Email)
	}

	fmt.Println()
	fmt.Printf("  %s\n", bold("Active identity"))

	if insideRepo() {
		localName := gitGet("config", "--local", "user.name")
		localEmail := gitGet("config", "--local", "user.email")

		if localName != "" {
			fmt.Printf("    %-16s %s <%s>\n", green("local"), localName, localEmail)
			fmt.Printf("    %-16s %s <%s>\n", dim("global"), globalName, globalEmail)
			warnIfUnmatched(cfg, localEmail)
		} else {
			fmt.Printf("    %-16s %s <%s>\n", green("global"), globalName, globalEmail)
			fmt.Printf("    %-16s %s\n", dim("local"), dim("no override set for this repo"))
			warnIfUnmatched(cfg, globalEmail)
		}
	} else {
		fmt.Printf("    %-16s %s <%s>\n", green("global"), globalName, globalEmail)
		fmt.Printf("    %-16s %s\n", dim(""), dim("not inside a git repo"))
		warnIfUnmatched(cfg, globalEmail)
	}

	fmt.Println()
}

// warnIfUnmatched prints a warning if the active email isn't in any saved profile.
func warnIfUnmatched(cfg Config, activeEmail string) {
	if activeEmail == "" {
		printWarning("no identity configured — run %s to set one", bold("git-profile <profile>"))
		return
	}
	for _, p := range cfg.Profiles {
		if p.Email == activeEmail {
			return
		}
	}
	fmt.Println()
	printWarning("active identity does not match any saved profile")
}

func showHelp(cfg Config) {
	fmt.Printf(`
  %s  v`+version+`

  %s
    git-profile                                 show profiles and active identity
    git-profile <profile>                       apply profile to current repo
    git-profile <profile> --global              apply profile globally
    git-profile add <profile> <name> <email>    save a new or updated profile
    git-profile remove <profile>                delete a profile
    git-profile --version                       show version
    git-profile --help                          show this help

  %s
    %s

  %s
`,
		bold("git-profile"),
		bold("USAGE"),
		bold("CONFIG"),
		cyan(configPath()),
		bold("PROFILES"),
	)

	if len(cfg.Profiles) == 0 {
		fmt.Printf("    %s\n", dim("no profiles saved yet"))
	} else {
		for _, key := range sortedKeys(cfg) {
			p := cfg.Profiles[key]
			fmt.Printf("    %-16s %s <%s>\n", yellow(key), p.Name, p.Email)
		}
	}
	fmt.Println()
}
