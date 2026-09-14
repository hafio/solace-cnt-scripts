package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	"solace/internal/config"
)

// replSite is a kubernetes-reached replication site, with the cluster carried in the
// command the way a DR mate in another cluster has to be.
func replSite() config.ReplSite {
	return config.ReplSite{
		VirtualRouterName: "v:sol-dr",
		RouterNames:       []string{"sol-dr1"},
		Endpoints:         []config.ReplEndpoint{{Host: "10.0.1.1", Port: 55443, Transport: config.TransportSSL}},
		Via: config.ReplVia{Kubernetes: &config.ReplViaKube{
			Command:   config.Command{"kubectl", "--context", "dr"},
			Namespace: "solace-dr",
			// Deliberately unlike the local deployment's name and namespace: if the
			// two matched, an assertion that the SITE's values were used could not fail.
			Name: "dr-broker",
		}},
	}
}

// replCfg is a valid local config that also declares the DR pair, so SiteCommand has
// something to look up.
func replCfg(t *testing.T) *config.Config {
	t.Helper()
	c := loadK8s(t)
	c.Replication = config.Replication{
		Sites: []config.ReplSite{
			{
				VirtualRouterName: "v:sol-local",
				RouterNames:       []string{"sol-local1"},
				Endpoints: []config.ReplEndpoint{
					{Host: "10.0.0.1", Port: 55443, Transport: config.TransportSSL},
				},
			},
			replSite(),
		},
		VPNs: []config.ReplVPN{{Name: "ORDERS", ActiveAt: "v:sol-local"}},
	}
	return c
}

// unconfiguredReplication is the smallest reply that carries a software banner and a
// parseable (empty) mate report, which is all Preflight needs.
const unconfiguredReplication = "Solace Event Broker Software Enterprise Version 10.26.0.8827\n" +
	"license-software\n\nReplication Interface:\nReplication Mate:\n\nConfigSync:\n"

// TestNewMateChannelArgv pins the exact kubectl argv, which is the whole point of this
// file living in internal/k8s rather than internal/broker.
//
// Three things it holds. The SITE's own command, namespace and broker name are used --
// not this env file's, because a DR mate is a different deployment in a different
// cluster. The pod is named with the role's LETTER (`-p-0`), the way podName builds it;
// spelling the word would name a pod that does not exist. And the script is uploaded on
// stdin and then run by its RELATIVE name, because the CLI's working directory is the
// scripts folder -- the same pair Ops.RunCLI has always used.
func TestNewMateChannelArgv(t *testing.T) {
	r := &recRunner{out: []byte(unconfiguredReplication)}
	c, err := NewMateChannel(r, replCfg(t), replSite(), nil)
	if err != nil {
		t.Fatalf("NewMateChannel: %v", err)
	}
	if err := c.Preflight(context.Background()); err != nil {
		t.Fatalf("Preflight: %v", err)
	}
	if len(r.calls) != 3 {
		t.Fatalf("made %d calls, want an upload, an exec and a cleanup: %+v", len(r.calls), r.calls)
	}

	upload, exec := r.calls[0], r.calls[1]

	wantUpload := []string{"--context", "dr", "exec", "-i", "-n", "solace-dr", "dr-broker-pubsubplus-p-0", "--", "sh", "-c"}
	if upload.name != "kubectl" || !hasPrefixArgv(upload.args, wantUpload) {
		t.Errorf("upload = %s %v,\nwant kubectl %v ...", upload.name, upload.args, wantUpload)
	}
	wantExec := []string{"--context", "dr", "exec", "-n", "solace-dr", "dr-broker-pubsubplus-p-0", "--",
		"/usr/sw/loads/currentload/bin/cli", "-Apes", ".repl-show-replication.cli"}
	if exec.name != "kubectl" || strings.Join(exec.args, " ") != strings.Join(wantExec, " ") {
		t.Errorf("exec = %s %v,\nwant kubectl %v", exec.name, exec.args, wantExec)
	}
	if strings.Contains(strings.Join(exec.args, " "), "-primary-0") {
		t.Error("the pod is named with the role LETTER; `-primary-0` names a pod that does not exist")
	}

	// The script body rides stdin, never argv -- the transport's rule, and not worth a
	// local exception just because this particular script carries no secret.
	if !strings.Contains(upload.stdin, "show replication") {
		t.Errorf("the script must be uploaded on stdin, got %q", upload.stdin)
	}
	if strings.Contains(strings.Join(upload.args, " "), "show replication") {
		t.Error("the script body must never reach argv")
	}
	// And it opens the way every generated script does.
	if !strings.HasPrefix(upload.stdin, "home\nno paging\n") {
		t.Errorf("the script must turn paging off after home, got %q", upload.stdin)
	}
}

// TestNewMateChannelUsesTheSiteNotTheLocalDeployment is the mistake this design exists
// to prevent: reaching a DR mate with THIS env file's namespace and broker name, which
// would quietly drive the local broker while reporting on the remote one.
func TestNewMateChannelUsesTheSiteNotTheLocalDeployment(t *testing.T) {
	cfg := replCfg(t)
	r := &recRunner{out: []byte(unconfiguredReplication)}
	c, err := NewMateChannel(r, cfg, replSite(), nil)
	if err != nil {
		t.Fatalf("NewMateChannel: %v", err)
	}
	if err := c.Preflight(context.Background()); err != nil {
		t.Fatalf("Preflight: %v", err)
	}
	// Matched as the exact operand of `-n`, not by substring: the local namespace here
	// is "solace", which IS a substring of the site's "solace-dr", so a containment
	// check would pass no matter which one the code used.
	ns := operandOf(r.calls[1].args, "-n")
	if ns != "solace-dr" {
		t.Errorf("argv %v passed -n %q, want the SITE's namespace; the local one is %q",
			r.calls[1].args, ns, cfg.K8s.Namespace)
	}
	// Same for the pod, built from the SITE's broker name rather than cfg.K8s.Name.
	pod := r.calls[1].args[indexOf(r.calls[1].args, "-n")+2]
	if pod != "dr-broker-pubsubplus-p-0" {
		t.Errorf("pod = %q, want one named from the site's broker name %q, not the local %q",
			pod, "dr-broker", cfg.K8s.Name)
	}
}

// TestNewMateChannelRemovesTheScriptItUploaded is the leak this leg shipped with.
//
// The LOCAL leg has always deleted its uploaded scripts (Ops.removeCLI); this one did
// not, so every `configure dr` and `perform dr` over kubectl left a `.cli` file inside a
// broker belonging to the OTHER site. That is the one place this tool writes into
// somebody else's cluster, which is where a leak is least defensible -- and the comment
// justifying the upload says the script carries no secret TODAY, which is the word that
// makes leaving files behind expensive later.
//
// The cleanup is DEFERRED, so it must also run when the script itself failed: a run that
// was rejected is exactly the one an operator re-runs, and the first attempt's file must
// not still be sitting there.
func TestNewMateChannelRemovesTheScriptItUploaded(t *testing.T) {
	for _, tc := range []struct {
		name     string
		runFails bool
	}{
		{"clean run", false},
		{"failed run", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &recRunner{out: []byte(unconfiguredReplication)}
			if tc.runFails {
				// Fail the EXEC, not the upload: a script that never landed has
				// nothing to clean up, but one that ran and was refused does. The
				// queue is popped per Output call, so nil lets the upload through
				// and the error lands on the run that follows it.
				r.outErrQueue = []error{nil, errors.New("the broker refused the script")}
			}
			c, err := NewMateChannel(r, replCfg(t), replSite(), nil)
			if err != nil {
				t.Fatalf("NewMateChannel: %v", err)
			}
			_ = c.Preflight(context.Background())

			last := r.calls[len(r.calls)-1]
			joined := strings.Join(last.args, " ")
			if !strings.Contains(joined, "rm -f") {
				t.Fatalf("the uploaded script was left on the MATE's broker; last call was %s %v",
					last.name, last.args)
			}
			// Removed by its ABSOLUTE path: the cleanup is its own exec and does not
			// inherit the CLI's working directory the way `-Apes` does.
			if !strings.Contains(joined, "/usr/sw/jail/cliscripts/.repl-show-replication.cli") {
				t.Errorf("cleanup names %q, want the script's absolute in-broker path", joined)
			}
			// And against the SITE, not this deployment.
			if operandOf(last.args, "-n") != "solace-dr" {
				t.Errorf("cleanup went to namespace %q, want the site's", operandOf(last.args, "-n"))
			}
		})
	}
}

// TestPodNameForIsTheOneSpelling pins that the mate channel names its pod through the
// shared rule rather than rebuilding it. names.go exists so this rule lives once, and
// the mate leg could not call podName because a DR site is described by its own
// `via.kubernetes` block rather than a *config.Config -- so it spelled the rule out by
// hand, which is how two spellings of one rule start drifting apart.
func TestPodNameForIsTheOneSpelling(t *testing.T) {
	if got := podNameFor("dr-broker", config.Primary); got != "dr-broker-pubsubplus-p-0" {
		t.Errorf("podNameFor = %q", got)
	}
	// podName is the same rule for THIS deployment, and must stay that way.
	cfg := replCfg(t)
	if podName(cfg, config.Backup) != podNameFor(cfg.K8s.Name, config.Backup) {
		t.Error("podName must be podNameFor for the local deployment, not a second spelling")
	}
}

// operandOf returns the token after flag, or "" when the flag is absent or last.
func operandOf(argv []string, flag string) string {
	i := indexOf(argv, flag)
	if i < 0 || i+1 >= len(argv) {
		return ""
	}
	return argv[i+1]
}

func indexOf(argv []string, want string) int {
	for i, a := range argv {
		if a == want {
			return i
		}
	}
	return -1
}

// TestNewMateChannelGuardsTheSiteCommand proves the execution guard runs when the
// channel is BUILT, not only in Validate.
//
// The channel is constructed straight from a *config.Config and must not assume Validate
// ever ran -- the same reason Cluster.cmd and kubectlTransport.cmd re-check their own
// command on every call.
func TestNewMateChannelGuardsTheSiteCommand(t *testing.T) {
	cfg := replCfg(t)
	bad := replSite()
	bad.Via.Kubernetes.Command = config.Command{"curl"}
	cfg.Replication.Sites[1] = bad

	if _, err := NewMateChannel(&recRunner{}, cfg, bad, nil); err == nil {
		t.Fatal("a site command outside the kubernetes allowlist must be refused at construction")
	}
}

// TestNewMateChannelRefusesANonKubeSite proves a site reached another way is refused
// here rather than silently driven over the wrong mechanism.
func TestNewMateChannelRefusesANonKubeSite(t *testing.T) {
	cfg := replCfg(t)
	s := replSite()
	s.Via = config.ReplVia{SEMP: &config.ReplViaSEMP{Host: "h", Port: 1943, Pass: "p"}}
	cfg.Replication.Sites[1] = s
	if _, err := NewMateChannel(&recRunner{}, cfg, s, nil); err == nil {
		t.Error("a site declaring via.semp must not be built as a kubernetes channel")
	}
}

// sempSite is a site reached over SEMP whose password lives in a cluster -- the case
// ReadSecretKey exists for, and the one that has no via.kubernetes to borrow a command
// from.
func sempSite() config.ReplSite {
	s := replSite()
	s.Via = config.ReplVia{SEMP: &config.ReplViaSEMP{Host: "10.0.1.1", Port: 1943,
		PassSecret: &config.ReplPassSecret{Namespace: "solace-dr", Name: "dr-creds", Key: "username_admin_password"}}}
	return s
}

// TestReadSecretKeyArgvAndDecoding pins the read: the SITE's own cluster CLI, one key
// fetched by jsonpath, and the value base64-decoded with the trailing newline dropped.
//
// The newline matters more than it looks. `kubectl create secret --from-file` keeps the
// one at the end of the file and `echo` adds one; either way it is part of the stored
// value, and sending it as part of the password fails the login with nothing on screen
// to say why.
func TestReadSecretKeyArgvAndDecoding(t *testing.T) {
	// "s3cret\n" -- the stored value carries the trailing newline the read must drop.
	r := &recRunner{out: []byte("czNjcmV0Cg==\n")}
	site := replSite()
	ref := &config.ReplPassSecret{Namespace: "solace-dr", Name: "dr-creds", Key: "adminpass"}

	got, err := ReadSecretKey(r, replCfg(t), site, ref)
	if err != nil {
		t.Fatalf("ReadSecretKey: %v", err)
	}
	if got != "s3cret" {
		t.Errorf("password = %q, want the decoded value with no trailing newline", got)
	}
	if len(r.calls) != 1 {
		t.Fatalf("made %d calls, want exactly one read: %+v", len(r.calls), r.calls)
	}
	c := r.calls[0]
	want := []string{"--context", "dr", "get", "secret", "dr-creds", "-n", "solace-dr",
		"-o", "jsonpath={.data.adminpass}"}
	if c.name != "kubectl" || strings.Join(c.args, " ") != strings.Join(want, " ") {
		t.Errorf("read = %s %v,\nwant kubectl %v", c.name, c.args, want)
	}
	// `get -o yaml` would pull every key of the Secret through this process, and a
	// redirect of it into a file would leave them on disk.
	if strings.Contains(strings.Join(c.args, " "), "-o yaml") {
		t.Error("only the one key is fetched; -o yaml would carry every other one with it")
	}
}

// TestReadSecretKeyEscapesADottedKey is the silent-failure case: a jsonpath member with
// an unescaped dot is read as a PATH STEP, so `{.data.tls.key}` looks for `key` inside
// `tls` -- which does not exist -- and kubectl exits 0 printing nothing. The empty
// reading is then indistinguishable from an absent key.
func TestReadSecretKeyEscapesADottedKey(t *testing.T) {
	r := &recRunner{out: []byte("czNjcmV0Cg==")}
	ref := &config.ReplPassSecret{Namespace: "ns", Name: "n", Key: "tls.key"}
	if _, err := ReadSecretKey(r, replCfg(t), replSite(), ref); err != nil {
		t.Fatalf("ReadSecretKey: %v", err)
	}
	got := operandOf(r.calls[0].args, "-o")
	if got != `jsonpath={.data.tls\.key}` {
		t.Errorf("jsonpath = %q, want the dot escaped so it names one key", got)
	}
}

// TestReadSecretKeyFallsBackToThisClusterCLI proves a SEMP-reached site can still keep
// its password in a cluster: with no via.kubernetes to borrow a command from, the read
// uses THIS env file's kubernetes.command.
func TestReadSecretKeyFallsBackToThisClusterCLI(t *testing.T) {
	r := &recRunner{out: []byte("czNjcmV0Cg==")}
	site := sempSite()
	cfg := replCfg(t)
	cfg.Replication.Sites[1] = site

	if _, err := ReadSecretKey(r, cfg, site, site.Via.SEMP.PassSecret); err != nil {
		t.Fatalf("ReadSecretKey: %v", err)
	}
	if strings.Contains(strings.Join(r.calls[0].args, " "), "--context dr") {
		t.Error("a site with no via.kubernetes has no cluster command to borrow; " +
			"the read must fall back to this env file's runtime")
	}
	if r.calls[0].name != cfg.K8s.Command.Name() {
		t.Errorf("read ran %q, want this env file's command %q", r.calls[0].name, cfg.K8s.Command.Name())
	}
}

// TestReadSecretKeyRefusesAnEmptyReading is why the empty check exists at all: kubectl
// exits 0 and prints NOTHING for a key the Secret does not carry, so without this the
// mate would be dialled with an empty password and the failure would surface as a login
// refusal against a broker that is perfectly healthy.
func TestReadSecretKeyRefusesAnEmptyReading(t *testing.T) {
	r := &recRunner{out: []byte("\n")}
	ref := &config.ReplPassSecret{Namespace: "solace-dr", Name: "dr-creds", Key: "absent"}
	_, err := ReadSecretKey(r, replCfg(t), replSite(), ref)
	if err == nil {
		t.Fatal("an empty reading must fail loudly; kubectl exits 0 for a key that is not there")
	}
	for _, want := range []string{"dr-creds", "absent"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q, so it cannot be acted on", err, want)
		}
	}
}

// TestReadSecretKeyRefusesNonBase64 keeps a hand-made Secret from becoming a password
// of decoded garbage. `.data` is always base64; `.stringData` is write-only and never
// read back, so a value that does not decode means the wrong field was read.
func TestReadSecretKeyRefusesNonBase64(t *testing.T) {
	r := &recRunner{out: []byte("not base64 at all !!")}
	ref := &config.ReplPassSecret{Namespace: "ns", Name: "n", Key: "k"}
	if _, err := ReadSecretKey(r, replCfg(t), replSite(), ref); err == nil {
		t.Error("a value that is not base64 must be refused, not decoded to garbage")
	}
}

// TestReadSecretKeyGuardsTheCommand holds the execution guard on this path too: the read
// is built from a *config.Config and must not assume Validate ever ran.
func TestReadSecretKeyGuardsTheCommand(t *testing.T) {
	cfg := replCfg(t)
	bad := replSite()
	bad.Via.Kubernetes.Command = config.Command{"curl"}
	cfg.Replication.Sites[1] = bad

	r := &recRunner{out: []byte("czNjcmV0Cg==")}
	if _, err := ReadSecretKey(r, cfg, bad, &config.ReplPassSecret{Namespace: "n", Name: "n", Key: "k"}); err == nil {
		t.Fatal("a site command outside the kubernetes allowlist must be refused before the read")
	}
	if len(r.calls) != 0 {
		t.Errorf("the guard must stop before exec, but %d call(s) ran: %+v", len(r.calls), r.calls)
	}
}

func hasPrefixArgv(got, want []string) bool {
	if len(got) < len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
