package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Profile holds a named Git author identity.
type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Config is the top-level structure of ~/.git-profiles.json.
type Config struct {
	Profiles map[string]Profile `json:"profiles"`
}

// configPath returns the path to the user's profile config file.
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".git-profiles.json")
}

// loadConfig reads and parses the config file.
// On first run, it bootstraps a sample file and exits.
func loadConfig() (Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return bootstrapConfig(path)
		}
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	return cfg, nil
}

// bootstrapConfig creates a sample config file and exits, prompting the user to edit it.
func bootstrapConfig(path string) (Config, error) {
	sample := Config{
		Profiles: map[string]Profile{
			"work":     {Name: "Your Name", Email: "you@work.com"},
			"personal": {Name: "Your Name", Email: "you@personal.com"},
		},
	}
	if err := writeConfig(sample); err != nil {
		return Config{}, fmt.Errorf("creating sample config: %w", err)
	}
	fmt.Printf("\n  Created sample config at %s\n", cyan(path))
	fmt.Printf("  Edit it with your real details, then run %s again.\n\n", bold("git-profile"))
	os.Exit(0)
	return Config{}, nil // unreachable
}

// writeConfig saves the config atomically using a temp file + rename.
func writeConfig(cfg Config) error {
	path := configPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// saveProfile adds or overwrites a profile and persists the config.
func saveProfile(cfg Config, profileName, name, email string) {
	_, exists := cfg.Profiles[profileName]
	cfg.Profiles[profileName] = Profile{Name: name, Email: email}
	if err := writeConfig(cfg); err != nil {
		fatal("saving config: %v", err)
	}
	verb := "added"
	if exists {
		verb = "updated"
	}
	printSuccess("%s profile %s", verb, yellow(profileName))
	fmt.Printf("         %s <%s>\n\n", name, email)
}

// deleteProfile removes a profile and persists the config.
func deleteProfile(cfg Config, profileName string) {
	if _, ok := cfg.Profiles[profileName]; !ok {
		fatal("no profile named %q", profileName)
	}
	delete(cfg.Profiles, profileName)
	if err := writeConfig(cfg); err != nil {
		fatal("saving config: %v", err)
	}
	printSuccess("removed profile %s\n", yellow(profileName))
}

// sortedKeys returns profile names in alphabetical order.
func sortedKeys(cfg Config) []string {
	keys := make([]string, 0, len(cfg.Profiles))
	for k := range cfg.Profiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// validateEmail does a minimal sanity check on an email address.
func validateEmail(email string) error {
	if !strings.Contains(email, "@") ||
		strings.HasPrefix(email, "@") ||
		strings.HasSuffix(email, "@") {
		return fmt.Errorf("invalid email address: %q", email)
	}
	return nil
}
