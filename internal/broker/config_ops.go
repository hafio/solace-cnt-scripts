package broker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"solace/internal/config"
)

// showVPNName is the uploaded name of the VPN listing script. Both default-VPN
// directions run it as their confirmation and the default-user path runs it to learn
// which VPNs exist, so the name is written once: runCLIRead and removeCLI must agree on
// it or a script is left behind in the broker's cliscripts dir.
const showVPNName = "show-vpn"

// ServerCert loads the TLS server certificate into each of roles over the Solace
// CLI, porting the CLI branch of 051 (the $SOLBK_SVR_SECRET k8s-secret fast path
// is handled by the k8s platform, not here). It concatenates key + cert + CAs
// into the tls-<dt>.crt.key file the broker loads. The private key rides Upload's
// stdin, so it never appears in an argv or an echoed command (§3).
func (o *Ops) ServerCert(ctx context.Context, dt string, roles ...config.Role) error {
	// The key+cert pair comes from ServerCertBundle so its ORDER has one definition
	// shared with the bundle the container platforms mount. The CAs are appended
	// only here: they are not part of the certificate the broker presents, and
	// including them on this path is existing behaviour kept as-is rather than a
	// property of the bundle (see ServerCertBundle).
	bundle, err := ServerCertBundle(o.Cfg)
	if err != nil {
		return err
	}
	for _, ca := range o.Cfg.TLS.CAs {
		caBytes, err := os.ReadFile(ca)
		if err != nil {
			return fmt.Errorf("read tls CA %q: %w", ca, err)
		}
		bundle = append(bundle, caBytes...)
	}

	file := serverCertFile(dt)
	for _, role := range roles {
		o.logf("Loading server certificate into %q node...", role)
		if err := o.T.Upload(ctx, role, bundle, certPath(file)); err != nil {
			return fmt.Errorf("upload certificate into %q node: %w", role, err)
		}
		out, err := o.RunCLI(ctx, role, "apply-server-certs", serverCertScript(dt))
		if err != nil {
			return err
		}
		o.show(out)
	}
	return nil
}

// DomainCerts loads the given domain certificate authorities into the node,
// porting 052. files maps CA name -> certificate filename located under folder;
// each file is uploaded to the certs dir and referenced by the generated CLI.
func (o *Ops) DomainCerts(ctx context.Context, role config.Role, folder string, files map[string]string) error {
	if len(files) == 0 {
		o.logf("No domain certificate authorities configured -- skipping.")
		return nil
	}
	for _, ca := range sortedKeys(files) {
		file := files[ca]
		if err := validName("domain CA name", ca); err != nil {
			return err
		}
		if err := validName("domain certificate filename", file); err != nil {
			return err
		}
		if err := o.T.UploadFile(ctx, role, filepath.Join(folder, file), certPath(file)); err != nil {
			return fmt.Errorf("upload domain certificate %q: %w", file, err)
		}
	}
	out, err := o.RunCLI(ctx, role, "load-domain-certs", domainCertsScript(files))
	if err != nil {
		return err
	}
	o.show(out)
	return nil
}

// DisableDefaultVPN shuts the default message-VPN down on the node, porting 053,
// then shows the resulting VPN list.
func (o *Ops) DisableDefaultVPN(ctx context.Context, role config.Role) error {
	return o.defaultVPN(ctx, role, "disable-default-vpn", disableDefaultVPNScript())
}

// EnableDefaultVPN starts the default message-VPN back up, the inverse of
// DisableDefaultVPN. It reports the same way, so an operator who has met one has met
// both.
func (o *Ops) EnableDefaultVPN(ctx context.Context, role config.Role) error {
	return o.defaultVPN(ctx, role, "enable-default-vpn", enableDefaultVPNScript())
}

// defaultVPN runs one default-VPN script and then shows the VPN list. Both directions
// share it so the confirmation an operator reads afterwards comes from one `show`, not
// from two that could drift apart.
func (o *Ops) defaultVPN(ctx context.Context, role config.Role, script, body string) error {
	if _, err := o.RunCLI(ctx, role, script, body); err != nil {
		return err
	}
	out, err := o.runCLIRead(ctx, role, showVPNName, showVPNScript())
	if err != nil {
		return err
	}
	o.show(out)
	// Only showVPNName needs cleanup here: script ran through the wrapped RunCLI
	// above, which cleans up its own broker-side files on exit.
	o.removeCLI(ctx, role, showVPNName)
	return nil
}

// DisableDefaultUsers shuts down the "default" client-username in every VPN on
// the node, porting 054: list the VPNs, parse their names, then shut each down.
func (o *Ops) DisableDefaultUsers(ctx context.Context, role config.Role) error {
	return o.defaultUsers(ctx, role, "disable", "disable-default-usernames", disableDefaultUsersScript)
}

// EnableDefaultUsers starts the "default" client-username back up in every VPN on the
// node, the inverse of DisableDefaultUsers.
func (o *Ops) EnableDefaultUsers(ctx context.Context, role config.Role) error {
	return o.defaultUsers(ctx, role, "enable", "enable-default-usernames", enableDefaultUsersScript)
}

// defaultUsers is the list-parse-apply shape both default-user directions share. The
// VPN list has to be read from the broker either way -- neither direction can be
// rendered from config, because which VPNs exist is broker state -- so the parse and
// its "nothing to act on" branch live here once. verb names the direction in that
// warning; build turns the parsed VPN names into the script.
func (o *Ops) defaultUsers(ctx context.Context, role config.Role, verb, script string,
	build func([]string) string) error {
	list, err := o.runCLIRead(ctx, role, showVPNName, showVPNBareScript())
	if err != nil {
		return err
	}
	o.removeCLI(ctx, role, showVPNName)
	vpns := parseVPNNames(string(list))
	if len(vpns) == 0 {
		o.progress().Warn("no message-VPNs parsed from broker output -- nothing to %s.", verb)
		return nil
	}
	// script runs through the wrapped RunCLI, which cleans up its own broker-side
	// files on exit and now fails loud if the broker rejects a line -- this used to
	// have no detection at all.
	out, err := o.RunCLI(ctx, role, script, build(vpns))
	if err != nil {
		return err
	}
	o.show(out)
	return nil
}

// ProductKeys applies keys to each of roles (Primary, plus Backup in HA),
// porting 057. It fails loud if the broker rejects a key.
//
// The rejection scan used to be a whole-transcript containsAnyFold("error",
// "fail") run once over every role's combined output. That is gone: RunCLI's own
// stop-on-error wrapper now scans each role's own transcript as it runs (against
// failKeywords, the vetted list -- a bare "error"/"fail" false-positives on an
// object legitimately named e.g. "error-events") and returns an error immediately,
// so a rejection on the Primary now stops before the Backup is ever touched
// instead of being merged into a combined buffer and checked at the end.
func (o *Ops) ProductKeys(ctx context.Context, keys []string, roles ...config.Role) error {
	if len(keys) == 0 {
		return fmt.Errorf("no product keys configured")
	}
	// Each key is interpolated into a line of a CLI script that runs with admin
	// already enabled, so it is checked before anything is uploaded -- the sibling
	// DomainCerts does the same for CA names and filenames.
	for _, k := range keys {
		if err := validCLILine("product key", k); err != nil {
			return err
		}
	}
	for _, role := range roles {
		o.logf("Applying product key(s) to %q node...", role)
		out, err := o.RunCLI(ctx, role, "product-keys", productKeysScript(keys))
		if err != nil {
			return err
		}
		o.show(out)
	}
	return nil
}

// AdditionalUsers, the op that created these users over the broker CLI, is GONE.
//
// It was kept unwired through the command-tree overhaul as a placeholder, on the note that
// the replacement would deliver the same schema as a Secret surfaced to the broker as
// environment variables. That is what k8s.AdditionalUsersSecret now does, so the
// placeholder has served its purpose. Both properties that made the CLI route awkward go
// with it: it was not re-runnable (`create username` fails on a user that exists) and its
// transcript repeated every password, so it could never show its own output.

// ExecCLI uploads a local Solace CLI script and runs it in the node, porting 059.
// The remote name is the file's basename, validated to keep it out of shell/CLI
// injection range.
//
// Like every other CLI execution in this package, the script now runs through the
// broker's own `source script ... stop-on-error no-prompt` wrapper (see RunCLI),
// via the same runCLISkeleton driver.go's chunks and RunCLI's writes share. That is
// a real behaviour change from what this ported: the bash original, and this
// port until now, ran every line of the script regardless of an earlier rejection
// and only reported afterwards (by counting suspect-looking lines) that something
// had failed. Now the BROKER stops at the first rejected line, so a script relying
// on a later line to recover from an earlier failure will stop instead of running
// to the end -- the full transcript (shown either way) says where.
func (o *Ops) ExecCLI(ctx context.Context, role config.Role, localPath string) error {
	// config.BaseName, not filepath.Base: filepath.Base is OS-dependent, so an env
	// file written on Windows with `cli\setup.cli` yielded "setup.cli" there and the
	// whole string on Linux. That name becomes the in-broker filename, so the two
	// answers named two different files for one env file.
	name := config.BaseName(localPath)
	if err := validName("cli script filename", name); err != nil {
		return err
	}
	// Unlike RunCLI's own writes, the body is already a local file: UploadFile
	// streams it from disk rather than buffering the whole script in memory, so it
	// is uploaded directly to the BARE path `source script` will run rather than
	// carried on OutputInput's stdin. writeBody=false in the skeleton below is what
	// skips the `cat > body` step RunCLI needs and this does not.
	bodyName, wrapName, outName := cliRunNames(name)
	dest := CLIScriptsDir + "/" + bodyName
	if err := o.T.UploadFile(ctx, role, localPath, dest); err != nil {
		return fmt.Errorf("upload cli script %q: %w", name, err)
	}
	skeleton := runCLISkeleton(bodyName, wrapName, outName, false)
	out, err := o.T.Output(ctx, role, "sh", "-c", skeleton)
	o.show(out)
	if err != nil {
		return fmt.Errorf("run cli script %q: %w", name, err)
	}
	if bad := rejectionIn(out); bad != "" {
		o.progress().Warn("errors detected in CLI output.")
		// The keyword only, never the line: like the removed additional-users op, a
		// CLI transcript can carry passwords.
		return fmt.Errorf("cli script %q rejected: the transcript carries %q; because of "+
			"stop-on-error the rest of the script did not run -- see the output above for detail",
			name, bad)
	}
	return nil
}

// ExecShellScript uploads a local shell script and runs it with bash inside the node.
//
// It has no bash ancestor: 059 only ever ran Solace CLI scripts. It is the sibling of
// ExecCLI and deliberately the same shape -- same config.BaseName rule, same validName
// gate, cleanup guaranteed on every exit -- so the two behave alike where they can.
// Four things differ, and each is a property of running arbitrary host code rather
// than a list of CLI commands:
//
//   - The script lands at the jail root, not in cliscripts (see shellScriptPath).
//   - Failure is the process's own exit status. Unlike a CLI script, whose lines run
//     independently so a rejected line has to be read out of the transcript, bash
//     reports one status for the whole run and that is what is returned. There is
//     therefore no stop-on-error wrapper and no rejectionIn scan here: the exit
//     status already says what those exist to reconstruct.
//   - The output IS shown, in full. A script that echoes a secret will therefore print
//     it; the command's help text says so, because the alternative -- withholding the
//     output the way AdditionalUsers does -- would make an operator's own script
//     useless to them.
//   - Cleanup is a deferred Go call rather than the broker-side `trap` ExecCLI's
//     skeleton carries, because there is no skeleton to hang one on: bash runs the
//     uploaded file directly.
//
// Either way the script must not be left behind in the broker if it fails partway,
// since whatever it carries is as sensitive as whatever the operator put in it.
func (o *Ops) ExecShellScript(ctx context.Context, role config.Role, localPath string) error {
	// config.BaseName, not filepath.Base, for the reason ExecCLI records: filepath.Base
	// is OS-dependent, so one env file would name two different in-broker files
	// depending on which host drove the run.
	name := config.BaseName(localPath)
	if err := validName("shell script filename", name); err != nil {
		return err
	}
	dest := shellScriptPath(name)
	if err := o.T.UploadFile(ctx, role, localPath, dest); err != nil {
		return fmt.Errorf("upload shell script %q: %w", name, err)
	}
	defer o.removeFiles(ctx, role, dest)

	o.logf("Running shell script %q on the %q node...", name, role)
	out, err := o.T.Output(ctx, role, "bash", dest)
	o.show(out)
	if err != nil {
		return fmt.Errorf("run shell script %q: %w", name, err)
	}
	return nil
}

// RemoveDomainCerts deletes each domain certificate authority from the node,
// porting 150 (k8s teardown domain-certs). cas are the CA names to remove; they
// are sorted for deterministic output and validated to keep them out of CLI
// injection range (§3). Empty input is a no-op with a log line.
func (o *Ops) RemoveDomainCerts(ctx context.Context, role config.Role, cas []string) error {
	if len(cas) == 0 {
		o.logf("No domain certificate authorities configured -- nothing to remove.")
		return nil
	}
	sorted := append([]string(nil), cas...)
	sort.Strings(sorted)
	for _, ca := range sorted {
		if err := validName("domain CA name", ca); err != nil {
			return err
		}
	}
	// RunCLI cleans up its own broker-side files on exit; no separate removeCLI.
	out, err := o.RunCLI(ctx, role, "remove-domain-certs", removeDomainCertsScript(sorted))
	if err != nil {
		return err
	}
	o.show(out)
	return nil
}

// RemoveServerCerts removes the TLS server certificate each node presents, the inverse
// of ServerCert. It runs on every role it is given, because a group with the certificate
// gone from one node and still loaded on another is a half state nobody asked for -- the
// apply path spans the whole group too.
//
// It TAKES TLS DOWN on each node it touches: every listener configured to present a
// certificate stops being able to, immediately. That is the point of the command rather
// than a hazard to be defended against here -- rotating a certificate out has to be
// possible, not only overwriting one -- so the confirmation lives where every other
// destructive command's does, in internal/cli, and this stays the mechanism.
//
// No cluster-side counterpart: on a Kubernetes deployment whose certificate comes from
// kubernetes.tlsServerSecret, the operator owns the mount and would put the certificate
// straight back. The CLI refuses that combination rather than starting a fight it loses.
func (o *Ops) RemoveServerCerts(ctx context.Context, roles ...config.Role) error {
	for _, role := range roles {
		o.logf("Removing the server certificate from %q node...", role)
		// RunCLI cleans up its own broker-side files on exit; no separate removeCLI.
		out, err := o.RunCLI(ctx, role, "remove-server-certs", removeServerCertScript())
		if err != nil {
			return err
		}
		o.show(out)
	}
	return nil
}

// RemoveProductKeys revokes each configured product key (`no product-key <key>`), the
// exact inverse of ProductKeys and deliberately its mirror image throughout: the same
// empty-list refusal, the same per-key validation before anything is uploaded, the same
// roles, and the same detection of the broker's own rejection.
//
// That last one carries more weight here than anywhere else in this package. Revoking a
// key the broker does not hold, or naming one it will not parse, is the kind of thing a
// CLI reports in prose and returns zero for -- so a removal that silently did nothing
// would read as success and the operator would believe an entitlement was gone when it is
// not. RunCLI's own stop-on-error wrapper and failKeywords scan (the vetted list, not the
// bare "error"/"fail" this used to scan a combined transcript for) is what turns that
// into a failed command, per role, as it happens -- rather than merging every role's
// output and checking it once at the end.
//
// Removing every key can leave the broker UNLICENSED, which is an outage whose cause
// points nowhere near this command. The gate for that lives in internal/cli with every
// other destructive confirmation; this stays the mechanism.
func (o *Ops) RemoveProductKeys(ctx context.Context, keys []string, roles ...config.Role) error {
	if len(keys) == 0 {
		return fmt.Errorf("no product keys configured")
	}
	for _, k := range keys {
		if err := validCLILine("product key", k); err != nil {
			return err
		}
	}
	for _, role := range roles {
		o.logf("Removing product key(s) from %q node...", role)
		out, err := o.RunCLI(ctx, role, "remove-product-keys", removeProductKeysScript(keys))
		if err != nil {
			return err
		}
		o.show(out)
	}
	return nil
}

// ServerCertBundle is the broker's server certificate as one PEM: the private KEY
// followed by the CERTIFICATE, which is the form Solace's
// tls_servercertificate_filepath expects and the same order the CLI path has always
// written (hence serverCertFile's .crt.key extension, and the bash ancestor's
// `cat CERTKEY CERT`).
//
// It deliberately does NOT append tls.cas. Trusted CAs reach the broker through
// `config apply domain-certs`, which installs them into the broker's own trust
// store; they are not part of the server certificate the broker presents. Ops.ServerCert
// still appends them for now, which is why it adds them itself rather than this
// returning them -- that difference is deliberate and is on the list to reconcile
// when the domain-certificate path is revisited.
//
// Order lives in this one expression so changing it is a one-line edit. What the
// file must contain has never been verified against a live broker; it is recorded
// as an unverified claim here.
func ServerCertBundle(cfg *config.Config) ([]byte, error) {
	if cfg.TLS.Cert == "" || cfg.TLS.CertKey == "" {
		return nil, fmt.Errorf("tls.cert and tls.certKey must both be set to build a server certificate")
	}
	certBytes, err := os.ReadFile(cfg.TLS.Cert)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", cfg.TLS.Cert, err)
	}
	// A cert file that ALREADY carries its private key is the one misconfiguration
	// this concatenation turns into a silently wrong file: the key would appear
	// twice, once from tls.certKey and once inside tls.cert. Refuse it and say which
	// of the two fields to change, because the operator has one file where this
	// schema expects two -- and a pre-chained file is exactly what a deployment
	// predating this tool is likely to have.
	if keyPEMRE.Match(certBytes) {
		return nil, fmt.Errorf("tls.cert %q contains a PRIVATE KEY block as well as the certificate, and "+
			"tls.certKey is also set, so the bundle would carry the key twice.\n"+
			"  Either point tls.cert at a certificate-only file (with the key in tls.certKey), or split the "+
			"combined file into its two halves", cfg.TLS.Cert)
	}
	keyBytes, err := os.ReadFile(cfg.TLS.CertKey)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", cfg.TLS.CertKey, err)
	}
	return append(keyBytes, certBytes...), nil
}

// keyPEMRE matches a PEM private-key header in any of the spellings openssl and its
// relatives emit: PRIVATE KEY, RSA PRIVATE KEY, EC PRIVATE KEY, ENCRYPTED PRIVATE
// KEY. Matching the header alone is enough -- this is a misconfiguration check, not
// a parser.
var keyPEMRE = regexp.MustCompile(`(?m)^-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----`)
