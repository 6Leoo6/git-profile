package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// ── Config ────────────────────────────────────────────────────────────────────

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Config struct {
	Profiles map[string]Profile `json:"profiles"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".git-profiles.json")
}

func loadConfig() (Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First run: create a sample config
			sample := Config{
				Profiles: map[string]Profile{
					"work":     {Name: "Your Name", Email: "you@work.com"},
					"personal": {Name: "Your Name", Email: "you@personal.com"},
				},
			}
			writeConfig(sample)
			fmt.Printf("Created sample config at %s\n", path)
			fmt.Println("Edit it with your real names and emails, then run git-profile again.")
			os.Exit(0)
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config JSON: %w", err)
	}
	return cfg, nil
}

func writeConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

// ── ANSI colors ───────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	red    = "\033[31m"
)

// enableANSI enables virtual terminal processing on Windows
// so ANSI escape codes render as colors instead of raw characters.
func enableANSI() {
	stdout := windows.Handle(os.Stdout.Fd())
	var mode uint32
	windows.GetConsoleMode(stdout, &mode)
	windows.SetConsoleMode(stdout, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

// ── Git helpers ───────────────────────────────────────────────────────────────

func gitGet(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitSet(scope, key, value string) error {
	return exec.Command("git", "config", scope, key, value).Run()
}

func insideRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// ── Commands ──────────────────────────────────────────────────────────────────

func showStatus(cfg Config) {
	globalName := gitGet("config", "--global", "user.name")
	globalEmail := gitGet("config", "--global", "user.email")

	fmt.Println()
	fmt.Printf("  %sProfiles%s\n", bold, reset)
	for key, p := range cfg.Profiles {
		fmt.Printf("    %s%-12s%s  %s (%s)\n", yellow, key, reset, p.Name, p.Email)
	}

	fmt.Println()
	fmt.Printf("  %sActive identity%s\n", bold, reset)

	if insideRepo() {
		localName := gitGet("config", "--local", "user.name")
		localEmail := gitGet("config", "--local", "user.email")

		if localName != "" {
			fmt.Printf("    %slocal (this repo)%s  %s (%s)\n", green, reset, localName, localEmail)
			fmt.Printf("    %sglobal             %s (%s)%s\n", dim, globalName, globalEmail, reset)

			// Check if local matches a profile
			matched := false
			for _, p := range cfg.Profiles {
				if p.Email == localEmail {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("\n  %s[!] Local identity does not match any saved profile%s\n", yellow, reset)
			}
		} else {
			fmt.Printf("    %sglobal%s  %s (%s)\n", green, reset, globalName, globalEmail)
			fmt.Printf("    %sno local override for this repo%s\n", dim, reset)

			matched := false
			for _, p := range cfg.Profiles {
				if p.Email == globalEmail {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("\n  %s[!] Global identity does not match any saved profile%s\n", yellow, reset)
			}
		}
	} else {
		fmt.Printf("    %sglobal%s  %s (%s)\n", green, reset, globalName, globalEmail)
		fmt.Printf("    %s(not inside a git repo)%s\n", dim, reset)
	}

	fmt.Println()
}

func switchProfile(cfg Config, profileName string, local bool) {
	p, ok := cfg.Profiles[profileName]
	if !ok {
		fmt.Printf("\n  %s[x] Unknown profile: %q%s\n", red, profileName, reset)
		fmt.Printf("    Available: %s\n\n", availableKeys(cfg))
		os.Exit(1)
	}

	scope := "--global"
	scopeLabel := "global"
	if local {
		if !insideRepo() {
			fmt.Printf("\n  %s[x] Not inside a Git repository - cannot set --local config%s\n\n", red, reset)
			os.Exit(1)
		}
		scope = "--local"
		scopeLabel = "local (this repo)"
	}

	if err := gitSet(scope, "user.name", p.Name); err != nil {
		fmt.Printf("\n  %s[x] Failed to set user.name: %v%s\n\n", red, err, reset)
		os.Exit(1)
	}
	if err := gitSet(scope, "user.email", p.Email); err != nil {
		fmt.Printf("\n  %s[x] Failed to set user.email: %v%s\n\n", red, err, reset)
		os.Exit(1)
	}

	fmt.Printf("\n  %s[ok] Switched to %q %s(%s)%s\n", green, profileName, dim, scopeLabel, reset)
	fmt.Printf("    %s (%s)\n\n", p.Name, p.Email)
}

func addProfile(cfg Config, profileName, name, email string) {
	if _, exists := cfg.Profiles[profileName]; exists {
		fmt.Printf("\n  %s[!] Profile %q already exists. Overwriting.%s\n", yellow, profileName, reset)
	}
	cfg.Profiles[profileName] = Profile{Name: name, Email: email}
	if err := writeConfig(cfg); err != nil {
		fmt.Printf("\n  %s[x] Failed to save config: %v%s\n\n", red, err, reset)
		os.Exit(1)
	}
	fmt.Printf("\n  %s[ok] Saved profile %q%s\n", green, profileName, reset)
	fmt.Printf("    %s (%s)\n\n", name, email)
}

func showHelp(cfg Config) {
	fmt.Printf(`
  %sgit-profile%s - manage Git identities

  %sUSAGE%s
    %sgit-profile%s                              show profiles and active identity
    %sgit-profile [name]%s                       apply profile to current repo (default)
    %sgit-profile [name] --global%s              apply profile globally
    %sgit-profile add [name] [Full Name] [email]%s  add or update a profile
    %sgit-profile --help%s                       show this help

  %sCONFIG FILE%s
    %s

  %sPROFILES%s
`, bold, reset,
		bold, reset,
		cyan, reset,
		cyan, reset,
		cyan, reset,
		cyan, reset,
		cyan, reset,
		bold, reset,
		configPath(),
		bold, reset)

	for key, p := range cfg.Profiles {
		fmt.Printf("    %s%-12s%s  %s (%s)\n", yellow, key, reset, p.Name, p.Email)
	}
	fmt.Println()
}

func availableKeys(cfg Config) string {
	keys := make([]string, 0, len(cfg.Profiles))
	for k := range cfg.Profiles {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	enableANSI()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]

	if len(args) == 0 {
		showStatus(cfg)
		return
	}

	switch args[0] {
	case "--help", "-h":
		showHelp(cfg)

	case "add":
		if len(args) != 4 {
			fmt.Printf("\n  %s[x] Usage: git-profile add [profile-name] [Full Name] [email]%s\n\n", red, reset)
			os.Exit(1)
		}
		addProfile(cfg, args[1], args[2], args[3])

	default:
		global := len(args) >= 2 && args[1] == "--global"
		switchProfile(cfg, args[0], !global)
	}
}
