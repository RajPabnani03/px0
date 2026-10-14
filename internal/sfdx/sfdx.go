// Package sfdx wraps the Salesforce CLI (sf) the way agent.go wraps coding
// harnesses: every call is a non-interactive --json subprocess behind a
// timeout with telemetry off, and a missing binary is reported as data, not
// an error. Nothing is cached: org state moves, so each call asks sf again.
package sfdx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Timeout bounds one sf subprocess. The CLI is a Node program: most reads
// finish in a second or two, but cold starts and big org lists need room.
const Timeout = 15 * time.Second

// Org is one row of `sf org list --json`, trimmed to what the UI reads. The
// raw result carries secrets (accessToken) that stay inside the subprocess.
type Org struct {
	Alias           string `json:"alias,omitempty"`
	Username        string `json:"username,omitempty"`
	InstanceURL     string `json:"instanceUrl,omitempty"`
	IsDefault       bool   `json:"isDefaultUsername,omitempty"`
	ConnectedStatus string `json:"connectedStatus,omitempty"`
}

// Installed reports whether the sf binary is on PATH right now, so a CLI
// installed since startup shows up without a restart.
func Installed() bool {
	_, err := exec.LookPath("sf")
	return err == nil
}

// Orgs runs `sf org list --json` in dir and flattens every result bucket
// (nonScratchOrgs, scratchOrgs, devHubs, sandboxes, other) into one list.
func Orgs(ctx context.Context, dir string) ([]Org, error) {
	out, err := run(ctx, dir, "org", "list", "--json")
	if err != nil {
		return nil, err
	}
	return parseOrgs(out)
}

// ConfigValue runs `sf config get <name> --json` in dir and returns the
// resolved value ("" when unset). A Local setting wins over a Global one.
func ConfigValue(ctx context.Context, dir, name string) (string, error) {
	out, err := run(ctx, dir, "config", "get", name, "--json")
	if err != nil {
		return "", err
	}
	return parseConfigValue(out, name), nil
}

// run executes sf non-interactively in dir: --json output only, telemetry
// off, and the context bounding the wait.
func run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sf", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "SF_DISABLE_TELEMETRY=true")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return nil, errors.New(msg)
		}
		return nil, err
	}
	return out, nil
}

func parseOrgs(data []byte) ([]Org, error) {
	var env struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(env.Result))
	for k := range env.Result {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	orgs := []Org{}
	for _, k := range keys {
		var bucket []Org
		if json.Unmarshal(env.Result[k], &bucket) == nil {
			orgs = append(orgs, bucket...)
		}
	}
	return orgs, nil
}

func parseConfigValue(data []byte, name string) string {
	var env struct {
		Result []struct {
			Name     string `json:"name"`
			Value    string `json:"value"`
			Location string `json:"location"`
		} `json:"result"`
	}
	if json.Unmarshal(data, &env) != nil {
		return ""
	}
	value := ""
	for _, e := range env.Result {
		if e.Name == name && e.Value != "" && (value == "" || e.Location == "Local") {
			value = e.Value
		}
	}
	return value
}
