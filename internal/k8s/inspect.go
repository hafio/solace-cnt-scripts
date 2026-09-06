package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"solace/internal/output"
)

// This file is the read side of the cluster: `kubectl get -o json`, decoded into
// narrow per-kind structs, so both the status commands and the deploy plan can
// say what exists and whether it is healthy.
//
// Two rules make owning this rendering safe rather than fragile:
//
//  1. Every struct here decodes ONLY the fields it reads. A field the cluster
//     does not carry (an older kubectl, a CRD whose status shape differs) decodes
//     to its zero value and is reported as unknown -- never as a failure. The
//     tool must not invent an alarm out of a field it could not find.
//  2. A kind this tool cannot list, or a field it cannot read, is reported as
//     such and never as a failure. `kubectl get` remains the source of truth for
//     anyone who needs more than this report chooses to show.
//
// Health words are deliberately conservative: LevelOK means "in its expected
// steady state", LevelWarn means "exists, not there yet" (Pending, no address
// assigned, replicas short), LevelFail means "needs intervention".

// now returns the clock, a seam so age rendering is testable. nil means time.Now.
func (c *Cluster) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// getJSON runs `kubectl get <args...> -o json` and decodes into v. A get that
// fails is returned as-is: the caller decides whether a kind it cannot list is
// fatal (it is not, for a report) or worth reporting.
func (c *Cluster) getJSON(ctx context.Context, v any, args ...string) error {
	raw, err := c.output(ctx, append(append([]string{"get"}, args...), "-o", "json")...)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		// The Echo runner returns nothing. Not an error: a preview has no cluster
		// to answer, and the caller renders "skipped (preview)".
		return nil
	}
	raw = normalizeToList(raw)
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("could not read `kubectl get %s -o json` output: %w -- "+
			"run `kubectl get %s -o json` yourself to see what it returned", strings.Join(args, " "), err, strings.Join(args, " "))
	}
	return nil
}

// normalizeToList makes both shapes kubectl returns decode into the same
// list-shaped struct.
//
// This exists because of a bug worth remembering: `kubectl get deployment -n ns
// -o json` returns a List with an `items` array, but `kubectl get deployment
// NAME -n ns -o json` returns the single Deployment OBJECT, with no `items` at
// all. A decoder expecting `items` finds none and reports the object as absent --
// silently, and in the worst possible direction: `status operator` said "not
// found in namespace solace-operator" about an operator that was running there.
//
// The unit tests did not catch it because the fixture had been captured from an
// unnamed get, so it carried `items` while production asked by name. Normalising
// here fixes both callers and closes the class: any future named get decodes too.
func normalizeToList(raw []byte) []byte {
	var probe struct {
		Kind  string          `json:"kind"`
		Items json.RawMessage `json:"items"`
	}
	if json.Unmarshal(raw, &probe) != nil {
		return raw // not an object we recognise; let the real decode report it
	}
	if probe.Items != nil || probe.Kind == "" || strings.HasSuffix(probe.Kind, "List") {
		return raw
	}
	// A single object: wrap it as a one-item list. raw is already valid JSON, so
	// concatenation is safe and avoids a decode/re-encode round trip.
	out := make([]byte, 0, len(raw)+12)
	out = append(out, []byte(`{"items":[`)...)
	out = append(out, raw...)
	return append(out, []byte(`]}`)...)
}

// --- decode shapes -------------------------------------------------------------

// objectMeta is the metadata every kind carries.
type objectMeta struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

// condition is one entry of the broker CR's status.conditions.
//
// VERIFIED on 2026-08-21 against operator CRD pubsubplus.solace.com/v1beta1: the
// types this operator publishes are ServiceReady, NoWarnings and HAReady -- NOT
// the conventional "Ready". Reading only a "Ready" condition would therefore find
// nothing and report a perfectly healthy broker as unknown, so the type names are
// matched explicitly below.
type condition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// The condition types this operator publishes, in the order a report should read
// them: can clients reach it, is the redundancy group healthy, is it complaining.
const (
	condServiceReady = "ServiceReady"
	condHAReady      = "HAReady"
	condNoWarnings   = "NoWarnings"
)

// brokerStatus is the operator-populated status.broker block. Every field is a
// string, including the booleans (haDeployment: "true"), which is why none of
// these are typed bool.
type brokerStatus struct {
	BrokerImage  string   `json:"brokerImage"`
	ServiceName  string   `json:"serviceName"`
	ServiceType  string   `json:"serviceType"`
	HADeployment string   `json:"haDeployment"`
	TLSSupport   string   `json:"tlsSupport"`
	TLSSecret    string   `json:"tlsSecret"`
	StatefulSets []string `json:"statefulSets"`
}

type podList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			// NodeName and PodIP are --detail only. NodeName earns its place there
			// because it is the one field that shows whether anti-affinity actually
			// spread the redundancy group across machines; PodIP rides along with it.
			NodeName string `json:"nodeName"`
		} `json:"spec"`
		Status struct {
			Phase             string `json:"phase"`
			PodIP             string `json:"podIP"`
			ContainerStatuses []struct {
				Ready        bool `json:"ready"`
				RestartCount int  `json:"restartCount"`
			} `json:"containerStatuses"`
		} `json:"status"`
	} `json:"items"`
}

type brokerList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Status   struct {
			Broker     brokerStatus `json:"broker"`
			Conditions []condition  `json:"conditions"`
			// PodsList is the operator's own list of the pods belonging to this
			// broker. Preferred over filtering all pods by a name infix: the CR
			// says which pods are its own, so a co-tenant workload whose name
			// happens to contain the infix cannot be mistaken for a broker pod.
			PodsList []string `json:"podsList"`
		} `json:"status"`
	} `json:"items"`
}

type deploymentList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			Replicas *int `json:"replicas"`
			// The running image, so the report can say which operator version is
			// actually installed. Worth reading rather than assuming: a cluster can
			// easily be running a newer operator than this tool renders, and an
			// operator debugging reconciliation needs to know which one answered.
			Template struct {
				Spec struct {
					Containers []struct {
						Image string `json:"image"`
					} `json:"containers"`
				} `json:"spec"`
			} `json:"template"`
		} `json:"spec"`
		Status struct {
			ReadyReplicas int `json:"readyReplicas"`
			Replicas      int `json:"replicas"`
		} `json:"status"`
	} `json:"items"`
}

type serviceList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			Type      string `json:"type"`
			ClusterIP string `json:"clusterIP"`
		} `json:"spec"`
		Status struct {
			LoadBalancer struct {
				Ingress []struct {
					IP       string `json:"ip"`
					Hostname string `json:"hostname"`
				} `json:"ingress"`
			} `json:"loadBalancer"`
		} `json:"status"`
	} `json:"items"`
}

type statefulSetList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			Replicas *int `json:"replicas"`
		} `json:"spec"`
		Status struct {
			ReadyReplicas int `json:"readyReplicas"`
		} `json:"status"`
	} `json:"items"`
}

type pvcList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
		Spec     struct {
			StorageClassName string `json:"storageClassName"`
			Resources        struct {
				Requests struct {
					Storage string `json:"storage"`
				} `json:"requests"`
			} `json:"resources"`
		} `json:"spec"`
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
}

// --- health rules --------------------------------------------------------------

// podHealth classifies one pod. Ready containers against total is the signal that
// matters; the phase is what distinguishes "not yet" from "needs looking at".
func podHealth(phase string, ready, total, restarts int) output.Level {
	switch {
	case phase == "Failed" || phase == "Unknown":
		return output.LevelFail
	case phase == "Succeeded":
		return output.LevelInfo // a completed pod is not a broker problem
	case total > 0 && ready == total && phase == "Running":
		if restarts > 0 {
			// Serving, but it has crashed before. Not a failure now; worth seeing.
			return output.LevelWarn
		}
		return output.LevelOK
	default:
		return output.LevelWarn
	}
}

// readyCount renders "ready/total" and counts restarts across containers.
func (p podList) at(i int) (ready, total, restarts int) {
	for _, cs := range p.Items[i].Status.ContainerStatuses {
		total++
		if cs.Ready {
			ready++
		}
		restarts += cs.RestartCount
	}
	return ready, total, restarts
}

// findCondition returns the named condition, if the operator published it.
func findCondition(conds []condition, want string) (condition, bool) {
	for _, c := range conds {
		if strings.EqualFold(c.Type, want) {
			return c, true
		}
	}
	return condition{}, false
}

// conditionLevel classifies one named condition. A True condition is the steady
// state; a False one carries the operator's own message, which is more useful
// than anything this tool could infer. A condition the operator did NOT publish
// is LevelInfo with no health claim -- a CRD version that drops or renames one
// must degrade to "not reported", never to a false alarm or a false pass.
func conditionLevel(conds []condition, want string) (output.Level, string) {
	c, ok := findCondition(conds, want)
	if !ok {
		return output.LevelInfo, "not reported by the operator"
	}
	msg := c.Message
	if msg == "" {
		msg = c.Reason
	}
	switch strings.ToLower(c.Status) {
	case "true":
		return output.LevelOK, msg
	case "false":
		// NoWarnings inverts: "there ARE warnings" is the operator telling us to
		// look, which is a warning rather than a failure of the deployment.
		if strings.EqualFold(want, condNoWarnings) {
			return output.LevelWarn, msg
		}
		return output.LevelFail, msg
	default:
		return output.LevelInfo, fmt.Sprintf("status %q: %s", c.Status, msg)
	}
}

// brokerConditionRows renders the three conditions this operator publishes, in
// the order that answers an operator's questions: can clients reach it, is the
// redundancy group whole, is it complaining about anything. Standalone
// deployments get no HAReady row, since there is no redundancy group to report.
func brokerConditionRows(conds []condition, ha bool) []struct {
	Key   string
	Level output.Level
	Value string
} {
	wanted := []struct{ key, cond string }{
		{"service", condServiceReady},
		{"ha group", condHAReady},
		{"warnings", condNoWarnings},
	}
	var rows []struct {
		Key   string
		Level output.Level
		Value string
	}
	for _, w := range wanted {
		if w.cond == condHAReady && !ha {
			continue
		}
		lvl, msg := conditionLevel(conds, w.cond)
		rows = append(rows, struct {
			Key   string
			Level output.Level
			Value string
		}{w.key, lvl, msg})
	}
	return rows
}

// replicaLevel classifies a controller by ready-vs-wanted replicas.
func replicaLevel(ready int, want *int) (output.Level, string) {
	wanted := 1
	if want != nil {
		wanted = *want
	}
	text := fmt.Sprintf("%d/%d", ready, wanted)
	switch {
	case wanted == 0:
		return output.LevelInfo, "scaled to zero"
	case ready == wanted:
		return output.LevelOK, text
	case ready == 0:
		return output.LevelFail, text
	default:
		return output.LevelWarn, text
	}
}

// serviceAddress renders what a client would connect to, and flags the one case
// worth flagging: a LoadBalancer with no address assigned yet.
func serviceAddress(kind, clusterIP string, ingress []struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}) (output.Level, string) {
	if kind != "LoadBalancer" {
		if clusterIP == "None" {
			return output.LevelInfo, "headless"
		}
		return output.LevelInfo, clusterIP
	}
	for _, in := range ingress {
		if in.IP != "" {
			return output.LevelOK, in.IP
		}
		if in.Hostname != "" {
			return output.LevelOK, in.Hostname
		}
	}
	return output.LevelWarn, "pending"
}

// pvcLevel classifies a claim: Bound is the steady state, Pending is the one an
// operator chases (usually a StorageClass that cannot provision).
func pvcLevel(phase string) output.Level {
	switch phase {
	case "Bound":
		return output.LevelOK
	case "Lost", "Failed":
		return output.LevelFail
	default:
		return output.LevelWarn
	}
}

// age renders a duration in kubectl's own AGE shape, deliberately: this column
// REPLACES the one kubectl printed, and an operator comparing the two should not
// have to wonder why "6m40s" became "6m". The bands mirror
// k8s.io/apimachinery/pkg/util/duration.HumanDuration -- two units while the
// number is small enough for the second to matter, one unit after that.
//
// A zero timestamp renders "-" rather than a computed age: the object did not
// tell us when it was created, and inventing "0s" would read as brand new.
func age(from, now time.Time) string {
	if from.IsZero() {
		return "-"
	}
	d := now.Sub(from)
	switch {
	case d < 0:
		return "0s"
	case d < time.Second:
		return "<1s"
	case d < 2*time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < 10*time.Minute:
		m, s := int(d/time.Minute), int(d/time.Second)%60
		if s == 0 {
			return fmt.Sprintf("%dm", m)
		}
		return fmt.Sprintf("%dm%ds", m, s)
	case d < 3*time.Hour:
		return fmt.Sprintf("%dm", int(d/time.Minute))
	case d < 8*time.Hour:
		h, m := int(d/time.Hour), int(d/time.Minute)%60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh%dm", h, m)
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d/time.Hour))
	default:
		days, h := int(d/(24*time.Hour)), int(d/time.Hour)%24
		if h == 0 || days >= 8 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd%dh", days, h)
	}
}

// roleRank orders a broker pod or StatefulSet by its place in the redundancy
// group -- primary, then backup, then monitor -- read off the
// `-pubsubplus-<p|b|m>` infix the operator's own naming guarantees (podName
// builds the same shape).
//
// It exists for ORDERING only. There is deliberately no ROLE column: the name
// already ends in -p-0/-b-0/-m-0, so a column repeating that would spend width
// saying what the reader can already see. What the reader cannot get from
// kubectl is the ORDER -- kubectl sorts alphabetically, which puts backup first
// and reads as though the group were listed arbitrarily.
//
// An unrecognised name sorts last rather than being dropped: a pod we cannot
// classify is still a pod worth showing.
func roleRank(name string) int {
	for i, infix := range []string{"-pubsubplus-p", "-pubsubplus-b", "-pubsubplus-m"} {
		if strings.Contains(name, infix) {
			return i
		}
	}
	return 3
}
