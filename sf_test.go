package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// sfStub puts a fake sf CLI on PATH so the /api/sf handlers have a binary to
// drive; body is the shell script, canned --json and all.
func sfStub(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sf"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

const sfBothScript = `#!/bin/sh
case "$1 $2" in
  "org list")
    cat <<'EOF'
{"status":0,"result":{"nonScratchOrgs":[
 {"alias":"dev","username":"u@dev.com","instanceUrl":"https://d.my.salesforce.com","isDefaultUsername":true,"connectedStatus":"Connected"}],
 "scratchOrgs":[],"devHubs":[],"sandboxes":[],"other":[]},"warnings":[]}
EOF
    ;;
  "config get")
    cat <<'EOF'
{"status":0,"result":[{"name":"target-org","value":"dev","location":"Local","success":true}],"warnings":[]}
EOF
    ;;
esac
`

func TestSFOrgsNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	s, _ := newTestServer(t)
	code, m := get(t, s, "/api/sf/orgs")
	if code != 200 || m["installed"] != false {
		t.Fatalf("code=%d body=%v, want 200 {installed:false}", code, m)
	}
}

func TestSFOrgsListsOrgs(t *testing.T) {
	sfStub(t, sfBothScript)
	s, _ := newTestServer(t)
	code, m := get(t, s, "/api/sf/orgs")
	if code != 200 || m["installed"] != true {
		t.Fatalf("code=%d body=%v", code, m)
	}
	orgs, ok := m["orgs"].([]any)
	if !ok || len(orgs) != 1 {
		t.Fatalf("orgs = %v", m["orgs"])
	}
	row := orgs[0].(map[string]any)
	if row["alias"] != "dev" || row["connectedStatus"] != "Connected" ||
		row["isDefaultUsername"] != true || row["instanceUrl"] != "https://d.my.salesforce.com" {
		t.Fatalf("row = %v", row)
	}
}

func TestSFDefaultJoinsOrgStatus(t *testing.T) {
	sfStub(t, sfBothScript)
	s, _ := newTestServer(t)
	code, m := get(t, s, "/api/sf/default")
	if code != 200 || m["installed"] != true {
		t.Fatalf("code=%d body=%v", code, m)
	}
	if m["value"] != "dev" {
		t.Fatalf("value = %v", m["value"])
	}
	org, ok := m["org"].(map[string]any)
	if !ok || org["alias"] != "dev" || org["connectedStatus"] != "Connected" {
		t.Fatalf("org = %v", m["org"])
	}
}

// A configured target-org that is not authenticated still reports its name;
// the org row just stays absent.
func TestSFDefaultUnmatchedOrg(t *testing.T) {
	sfStub(t, `#!/bin/sh
case "$1 $2" in
  "org list") echo '{"status":0,"result":{"nonScratchOrgs":[],"scratchOrgs":[],"other":[]},"warnings":[]}' ;;
  "config get") echo '{"status":0,"result":[{"name":"target-org","value":"ghost","success":true}],"warnings":[]}' ;;
esac
`)
	s, _ := newTestServer(t)
	code, m := get(t, s, "/api/sf/default")
	if code != 200 || m["value"] != "ghost" || m["org"] != nil {
		t.Fatalf("code=%d body=%v, want value=ghost org=nil", code, m)
	}
}

// A failing sf run is UI state, not a server fault: 200 with the error text.
func TestSFOrgsCommandFailure(t *testing.T) {
	sfStub(t, "#!/bin/sh\necho 'sf blew up' >&2\nexit 1\n")
	s, _ := newTestServer(t)
	code, m := get(t, s, "/api/sf/orgs")
	if code != 200 || m["installed"] != true || m["error"] == nil {
		t.Fatalf("code=%d body=%v, want 200 {installed:true,error}", code, m)
	}
}

func TestSFMethodNotAllowed(t *testing.T) {
	sfStub(t, sfBothScript)
	s, _ := newTestServer(t)
	for _, u := range []string{"/api/sf/orgs", "/api/sf/default"} {
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, u, nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s = %d, want 405", u, rec.Code)
		}
	}
}
