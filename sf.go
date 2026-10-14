package main

import (
	"net/http"

	"px0/internal/sfdx"
)

// Salesforce org awareness through the sf CLI, on the same model as
// agent.go's harness wrapping: reads only, --json subprocesses behind a
// timeout, nothing cached under the workspace. A missing sf binary answers
// {installed:false}; a failed one answers {installed:true, error} -- both
// are 200s, since the status chip renders either as UI state, not a fault.

// handleSFOrgs lists every org the sf CLI knows about in this workspace:
// `sf org list --json` flattened into org rows.
func (s *Server) handleSFOrgs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !sfdx.Installed() {
		writeJSON(w, map[string]any{"installed": false})
		return
	}
	orgs, err := sfdx.Orgs(r.Context(), s.ix.Root())
	resp := map[string]any{"installed": true, "orgs": orgs}
	if err != nil {
		resp["error"] = err.Error()
	}
	writeJSON(w, resp)
}

// handleSFDefault reports the workspace's target org: `sf config get
// target-org --json` resolved with the workspace root as cwd, then matched
// back into `sf org list` so the reply also carries its connected status.
// value is what was configured (alias or username); org is the list row it
// resolved to, absent when the configured org is not authenticated.
func (s *Server) handleSFDefault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !sfdx.Installed() {
		writeJSON(w, map[string]any{"installed": false})
		return
	}
	resp := map[string]any{"installed": true}
	value, err := sfdx.ConfigValue(r.Context(), s.ix.Root(), "target-org")
	if err != nil {
		resp["error"] = err.Error()
		writeJSON(w, resp)
		return
	}
	resp["value"] = value
	if value == "" {
		writeJSON(w, resp)
		return
	}
	orgs, err := sfdx.Orgs(r.Context(), s.ix.Root())
	if err != nil {
		resp["error"] = err.Error()
		writeJSON(w, resp)
		return
	}
	for _, o := range orgs {
		if o.Alias == value || o.Username == value {
			resp["org"] = o
			break
		}
	}
	writeJSON(w, resp)
}
