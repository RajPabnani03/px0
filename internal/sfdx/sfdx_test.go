package sfdx

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeSf puts a shell stub named sf on PATH; body is the script, which can
// branch on argv to answer different subcommands.
func fakeSf(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sf"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

const orgsScript = `#!/bin/sh
case "$1 $2" in
  "org list")
    cat <<'EOF'
{"status":0,"result":{
  "nonScratchOrgs":[{"alias":"dev","username":"u@dev.com","instanceUrl":"https://d.my.salesforce.com","isDefaultUsername":true,"connectedStatus":"Connected","accessToken":"00DSECRET"}],
  "scratchOrgs":[{"username":"s@scratch.com","connectedStatus":"Unknown"}],
  "devHubs":[],"sandboxes":[],"other":[]},"warnings":[]}
EOF
    ;;
  "config get")
    cat <<'EOF'
{"status":0,"result":[{"name":"target-org","value":"dev","location":"Local","success":true}],"warnings":[]}
EOF
    ;;
esac
`

func TestOrgsFlattensBucketsAndTrimsSecrets(t *testing.T) {
	fakeSf(t, orgsScript)
	orgs, err := Orgs(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) != 2 {
		t.Fatalf("orgs = %v, want 2 rows", orgs)
	}
	var dev *Org
	for i := range orgs {
		if orgs[i].Alias == "dev" {
			dev = &orgs[i]
		}
	}
	if dev == nil {
		t.Fatalf("no dev org in %v", orgs)
	}
	if dev.Username != "u@dev.com" || dev.InstanceURL != "https://d.my.salesforce.com" ||
		!dev.IsDefault || dev.ConnectedStatus != "Connected" {
		t.Fatalf("dev row = %+v", dev)
	}
	raw, _ := json.Marshal(orgs)
	if strings.Contains(string(raw), "00DSECRET") {
		t.Fatalf("accessToken leaked into response: %s", raw)
	}
}

func TestConfigValue(t *testing.T) {
	fakeSf(t, orgsScript)
	got, err := ConfigValue(context.Background(), t.TempDir(), "target-org")
	if err != nil {
		t.Fatal(err)
	}
	if got != "dev" {
		t.Fatalf("target-org = %q, want dev", got)
	}
}

func TestConfigValuePrefersLocal(t *testing.T) {
	fakeSf(t, `#!/bin/sh
cat <<'EOF'
{"status":0,"result":[
 {"name":"target-org","value":"global-alias","location":"Global","success":true},
 {"name":"target-org","value":"local-alias","location":"Local","success":true}],"warnings":[]}
EOF
`)
	got, err := ConfigValue(context.Background(), t.TempDir(), "target-org")
	if err != nil {
		t.Fatal(err)
	}
	if got != "local-alias" {
		t.Fatalf("target-org = %q, want local-alias", got)
	}
}

func TestConfigValueUnset(t *testing.T) {
	fakeSf(t, `#!/bin/sh
cat <<'EOF'
{"status":0,"result":[{"name":"target-org","success":true}],"warnings":[]}
EOF
`)
	got, err := ConfigValue(context.Background(), t.TempDir(), "target-org")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("unset target-org = %q, want \"\"", got)
	}
}

// The subprocess runs with telemetry off and cwd pinned to the workspace.
func TestRunEnvAndCwd(t *testing.T) {
	fakeSf(t, "#!/bin/sh\necho \"tel=$SF_DISABLE_TELEMETRY cwd=$(pwd)\"\n")
	dir := t.TempDir()
	// macOS TempDir lives under /var -> /private/var; the subshell's pwd is real.
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	out, err := run(context.Background(), dir, "org", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	want := "tel=true cwd=" + dir
	if strings.TrimSpace(string(out)) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
}

func TestOrgsPropagatesStderr(t *testing.T) {
	fakeSf(t, "#!/bin/sh\necho 'not authorized' >&2\nexit 1\n")
	_, err := Orgs(context.Background(), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("err = %v, want stderr text", err)
	}
}

func TestInstalled(t *testing.T) {
	fakeSf(t, "#!/bin/sh\n")
	if !Installed() {
		t.Fatal("sf stub on PATH not detected")
	}
	t.Setenv("PATH", t.TempDir())
	if Installed() {
		t.Fatal("sf reported installed with empty PATH")
	}
}
