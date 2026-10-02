package k8s

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"solace/internal/config"
)

// generatorPage is the browser page that builds an env file and previews what
// `operator generate` and `broker generate` render from it. Its logic is hand-written
// JavaScript, but every value it would otherwise have to copy from Go -- the operator
// bundle and the defaults it shows -- sits in one generated block, written by
// TestGeneratorPageEmbedsTheCLI under -update (the `regen` task).
const generatorPage = "../../solace-yaml-generator.html"

// The block's fences. Everything between them is rewritten wholesale under -update.
const (
	genBegin = "// BEGIN GENERATED -- regen writes this block from internal/k8s; never hand-edit\n"
	genEnd   = "// END GENERATED\n"
)

// generatorTiers are the maxConnections values the page offers. The tier table itself
// is unexported, so each one is resolved through ApplyDefaults rather than read.
var generatorTiers = []int{100, 1000, 10000, 100000, 200000}

// pageOperatorActions are the template actions the page's operator renderer
// implements (renderOperatorTemplate). NOT regenerated: a new action in the bundle has
// to fail this test until the JavaScript handles it, or regen would hide the gap.
var pageOperatorActions = []string{
	"{{.Namespace}}", "{{.WatchNamespace}}", "{{.Image}}", "{{.CPU}}", "{{.Mem}}",
	"{{- if .PullSecret}}", "{{- end}}",
}

type genTier struct {
	CPU          string `json:"cpu"`
	K8sMem       string `json:"k8sMem"`
	ContainerMem string `json:"containerMem"`
	CPUSet       string `json:"cpuset"`
}

type genScaling struct {
	MaxConnections      int `json:"maxConnections"`
	MaxQueueMessages    int `json:"maxQueueMessages"`
	MaxSpoolUsageMB     int `json:"maxSpoolUsageMB"`
	MaxKafkaBridge      int `json:"maxKafkaBridge"`
	MaxKafkaConnections int `json:"maxKafkaConnections"`
	MaxBridges          int `json:"maxBridges"`
	MaxSubscriptions    int `json:"maxSubscriptions"`
	MaxGuaranteedMsgMB  int `json:"maxGuaranteedMsgMB"`
}

type genHealthCheck struct {
	Interval    string `json:"interval"`
	Timeout     string `json:"timeout"`
	Retries     int    `json:"retries"`
	StartPeriod string `json:"startPeriod"`
}

type genDefaults struct {
	OperatorImage         string             `json:"operatorImage"`
	OperatorNamespace     string             `json:"operatorNamespace"`
	OperatorCPU           string             `json:"operatorCpu"`
	OperatorMem           string             `json:"operatorMem"`
	UpdateStrategy        string             `json:"updateStrategy"`
	MonNodeSize           string             `json:"monNodeSize"`
	KubernetesCommand     string             `json:"kubernetesCommand"`
	CLIScriptsDir         string             `json:"cliScriptsDir"`
	HostDiagnosticDir     string             `json:"hostDiagnosticDir"`
	ScalingKubernetes     genScaling         `json:"scalingKubernetes"`
	ScalingContainer      genScaling         `json:"scalingContainer"`
	Tiers                 map[string]genTier `json:"tiers"`
	DockerCommand         string             `json:"dockerCommand"`
	DockerCompose         string             `json:"dockerCompose"`
	PodmanCommand         string             `json:"podmanCommand"`
	PodmanQuadletDir      string             `json:"podmanQuadletDir"`
	RunUserDocker         string             `json:"runUserDocker"`
	RunUserPodman         string             `json:"runUserPodman"`
	RunUserPodmanRootless string             `json:"runUserPodmanRootless"`
	ContainerName         string             `json:"containerName"`
	DataDir               string             `json:"dataDir"`
	MonitorCPUSet         string             `json:"monitorCpuset"`
	UlimitsCore           string             `json:"ulimitsCore"`
	NetworkMode           string             `json:"networkMode"`
	HealthCheck           genHealthCheck     `json:"healthCheck"`
}

// genBlock is the page's GEN constant.
type genBlock struct {
	// OperatorTemplate is operatorBundle, gzipped and base64-encoded. Compared by its
	// decompressed text, so a compressor change between Go releases is not a diff.
	OperatorTemplate string      `json:"operatorTemplate"`
	Defaults         genDefaults `json:"defaults"`
}

func scalingOf(s config.Scaling) genScaling {
	return genScaling{s.MaxConnections, s.MaxQueueMessages, s.MaxSpoolUsageMB, s.MaxKafkaBridge,
		s.MaxKafkaConnections, s.MaxBridges, s.MaxSubscriptions, s.MaxGuaranteedMsgMB}
}

// generatorDefaults derives every default the page shows from the code that applies
// it, so the page cannot offer a default the CLI no longer has.
func generatorDefaults(t *testing.T) genDefaults {
	t.Helper()
	k := &config.Config{}
	k.ApplyDefaults(config.K8s)
	d := &config.Config{}
	d.ApplyDefaults(config.Docker)
	p := &config.Config{}
	p.ApplyDefaults(config.Podman)
	r := &config.Config{}
	r.Podman.Rootless = true
	r.ApplyDefaults(config.Podman)

	tiers := map[string]genTier{}
	for _, n := range generatorTiers {
		kt := &config.Config{}
		kt.Scaling.MaxConnections = n
		kt.ApplyDefaults(config.K8s)
		ct := &config.Config{}
		ct.Scaling.MaxConnections = n
		ct.ApplyDefaults(config.Docker)
		if kt.Scaling.CPU == "" {
			t.Fatalf("maxConnections %d is not a scaling tier any more: drop it from generatorTiers and "+
				"the page follows on regen", n)
		}
		tiers[strconv.Itoa(n)] = genTier{CPU: kt.Scaling.CPU, K8sMem: kt.K8s.MsgNode.Mem,
			ContainerMem: ct.Docker.Container.Mem, CPUSet: ct.Docker.Container.CPUSet}
	}
	hc := d.Docker.Container.HealthCheck
	return genDefaults{
		OperatorImage:         k.K8s.Operator.Image,
		OperatorNamespace:     defaultOperatorNS,
		OperatorCPU:           k.K8s.Operator.CPU,
		OperatorMem:           k.K8s.Operator.Mem,
		UpdateStrategy:        k.K8s.UpdateStrategy,
		MonNodeSize:           k.K8s.Storage.MonNodeSize,
		KubernetesCommand:     k.K8s.Command.String(),
		CLIScriptsDir:         k.Broker.CLIScriptsDir,
		HostDiagnosticDir:     k.Broker.HostDiagnosticDir,
		ScalingKubernetes:     scalingOf(k.Scaling),
		ScalingContainer:      scalingOf(d.Scaling),
		Tiers:                 tiers,
		DockerCommand:         d.Docker.Command.String(),
		DockerCompose:         d.Docker.Compose.String(),
		PodmanCommand:         p.Podman.Command.String(),
		PodmanQuadletDir:      p.Podman.QuadletDir,
		RunUserDocker:         d.Docker.Container.RunUser,
		RunUserPodman:         p.Podman.Container.RunUser,
		RunUserPodmanRootless: r.Podman.Container.RunUser,
		ContainerName:         d.Docker.Container.Name,
		DataDir:               d.Docker.Container.DataDir,
		MonitorCPUSet:         d.Docker.Container.MonitorCPUSet,
		UlimitsCore:           d.Docker.Container.Ulimits.Core,
		NetworkMode:           d.Docker.Network.Mode,
		HealthCheck:           genHealthCheck{hc.Interval, hc.Timeout, hc.Retries, hc.StartPeriod},
	}
}

func gzipBase64(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zw.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func gunzipBase64(t *testing.T, s string) string {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("the page's operatorTemplate is not base64: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("the page's operatorTemplate is not gzip: %v", err)
	}
	out, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("the page's operatorTemplate does not decompress: %v", err)
	}
	return string(out)
}

// templateActions lists the distinct {{...}} actions in the bundle, in order of first use.
func templateActions(tmpl string) []string {
	var out []string
	seen := map[string]bool{}
	for _, a := range regexp.MustCompile(`\{\{[^}]*\}\}`).FindAllString(tmpl, -1) {
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}

// TestGeneratorPageEmbedsTheCLI pins solace-yaml-generator.html to the CLI it previews.
// The page renders the operator bundle from the copy in its GEN block and fills every
// default it shows from there, so a bundle bump or a changed default reaches it only
// through this block: the test fails while the block is stale, naming what moved, and
// -update rewrites it. It also fails when the bundle gains a template action the page's
// renderer does not implement -- that one is a JavaScript change, which regen cannot make.
func TestGeneratorPageEmbedsTheCLI(t *testing.T) {
	if got := templateActions(operatorBundle); !reflect.DeepEqual(got, pageOperatorActions) {
		t.Errorf("the operator bundle uses the template actions %q, but the page's renderOperatorTemplate "+
			"implements only %q: extend it in solace-yaml-generator.html, then pageOperatorActions",
			got, pageOperatorActions)
	}

	page, err := os.ReadFile(generatorPage)
	if err != nil {
		t.Fatalf("read %s: %v", generatorPage, err)
	}
	start := bytes.Index(page, []byte(genBegin))
	end := bytes.Index(page, []byte(genEnd))
	if start < 0 || end < start {
		t.Fatalf("%s has no generated block: it must hold the line %q and, after it, %q",
			generatorPage, strings.TrimSpace(genBegin), strings.TrimSpace(genEnd))
	}
	want := genBlock{Defaults: generatorDefaults(t)}

	if *update {
		want.OperatorTemplate = gzipBase64(t, operatorBundle)
		js, err := json.MarshalIndent(want, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		out.Write(page[:start+len(genBegin)])
		out.WriteString("const GEN = " + string(js) + ";\n")
		out.Write(page[end:])
		if err := os.WriteFile(generatorPage, out.Bytes(), 0o644); err != nil {
			t.Fatalf("write %s: %v", generatorPage, err)
		}
		return
	}

	body := strings.TrimSpace(string(page[start+len(genBegin) : end]))
	body = strings.TrimSuffix(strings.TrimPrefix(body, "const GEN = "), ";")
	var got genBlock
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("%s's generated block is not `const GEN = <json>;` (regenerate: go test ./internal/k8s "+
			"-update): %v", generatorPage, err)
	}
	if gunzipBase64(t, got.OperatorTemplate) != operatorBundle {
		t.Errorf("%s embeds a different operator bundle than assets/ -- regenerate it with the regen task",
			generatorPage)
	}
	gv, wv := reflect.ValueOf(got.Defaults), reflect.ValueOf(want.Defaults)
	for i := 0; i < gv.NumField(); i++ {
		if !reflect.DeepEqual(gv.Field(i).Interface(), wv.Field(i).Interface()) {
			t.Errorf("%s shows %s = %v, but the CLI's default is %v -- regenerate it with the regen task",
				generatorPage, gv.Type().Field(i).Name, gv.Field(i).Interface(), wv.Field(i).Interface())
		}
	}
}
