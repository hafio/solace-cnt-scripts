# What `import-config` applies

This file is GENERATED from `internal/broker/sections.go` by
`internal/broker/importdoc_test.go`. Do not edit it by hand: run the `regen` task.

`broker perform import-config` decides what to do with a captured configuration
section by section, using the broker's own `! Configure X` / `! Create X` comments
as the unit. The classification is compiled in -- there is no config key and no
flag -- because an artifact that chose what to execute would be untrusted input
doing what `internal/config/execguard.go` exists to prevent.

A section marked **skip** is REMOVED by `export-config` -- the artifact never contains
it, and records the omission as a `! solace-util-omitted:` line -- and cannot be opted
into. `import-config` ignores such a section if an older or hand-edited artifact still
carries one, and reports it only then. Recover a node by deploying it from its own
env file and letting config-sync populate it.

## Applied

| Section | Why |
|---|---|
| `Configure SMF Service` | Bounces SMF (semp/all.cli:77, 82) to apply its listen ports. |
| `Configure SSL Service` | Connection-count thresholds only (semp/all.cli:96); no port change. |
| `Configure WEB Service` | Bounces web-transport (semp/all.cli:99, 103) to apply its ports. |
| `Configure REST Service` | Enables REST incoming/outgoing and their thresholds (semp/all.cli:108-110). |
| `Configure MQTT Service` | Starts MQTT back up (semp/all.cli:113). |
| `Configure AMQP Service` | Bounces AMQP (semp/all.cli:116, 118) to apply its port. |
| `Configure Virtual Hostnames` | Virtual hostnames and the message-VPN each is bound to, created inside the `service` scope (semp/all-cfg-tps-org.cli:93-103). Every virtual hostname already on the target is removed first, so the target ends up matching the artifact rather than carrying the union: a virtual hostname is a DNS name pointing at this broker, and a stale one left behind resolves somewhere it should not. |
| `Configure HealthCheck Service` | Bounces the health-check service (semp/all.cli:121, 124); port 5550 (semp/all.cli:122) is what render.go's healthPort targets, so a capture with a different port fails every pod's readiness. |
| `Configure Routing MNR` | CSPF SSL validation and queue thresholds (semp/all.cli:146-152); no shutdown. |
| `Configure Replication` | Carries a config-sync bridge pre-shared key (semp/all.cli:212) and `mate virtual-router-name` (semp/all.cli:205), so the replication mate must be re-imported to match. |
| `Configure Memory Event` | Event thresholds only (semp/all.cli:226-228). |
| `Configure Service` | Starts msg-backbone back up (semp/all.cli:232) after its thresholds. |
| `Configure MQTT` | `mqtt retain max-memory 0` (semp/all.cli:235). |
| `Create Syslog` | Syslog destinations: facilities and remote hosts with their transports (semp/all.cli:239-247). Applied, and uniquely among these it needs the TARGET consulted first -- a `create syslog` for an entry that already exists is rejected, so ClearExistingSyslogs emits `no syslog <name>` ahead of it, but only for a name the target actually has, because the removal is itself an error when the entry does not exist. |
| `Configure Schedule` | Disables the autobackup schedule (semp/all.cli:240). |
| `Create Compression` | `compression mode optimize-for-size` (semp/all.cli:243). |
| `Create console` | Console timeout and login banner (semp/all.cli:248-249). |
| `Create logging` | Logging changes end the CLI session (semp/all.cli:251-260), so any later command in the same invocation would never run; applying it alone in its own RunCLI removes the hazard, since each RunCLI execs a fresh cli process. |
| `Configure SSL` | Cipher-suite and TLS-version changes (semp/all.cli:264-267) can drop sessions already using the settings being replaced. |
| `Create Usernames` | Overwrites the target's CLI admin password with the source's (semp/all.cli:284, 286); harmless to this tool, since the CLI channel needs no broker credentials, but on Kubernetes adminCredentialsSecret still holds the old value, so pods go out of readiness until semp.adminPass is updated to match. |
| `Create LDAP Profile` | LDAP profiles, created inside the `authentication` scope and carrying an aes-password bind credential (semp/all-cfg-tps-org.cli:494). A `create ldap-profile` for a name the target already has is rejected, so ClearExistingNested emits `no ldap-profile <name>` first -- but only for names the target carries AND this artifact re-creates, since a removal for an absent profile is itself an error. |
| `Create Radius Profile` | Named object, shut down in this capture (semp/all.cli:325). |
| `Create Domain Certificate Authority` | Domain CAs, created inside the `ssl` scope (semp/all.cli:334-337). A `create domain-certificate-authority` for a name the target already has is rejected, so ClearExistingNested emits `no domain-certificate-authority <name>` first, for names the target carries AND this artifact re-creates. Overlaps `broker configure domain-certs`, which also owns this surface. |
| `Create Client Certificate Authority` | Client CA trust material -- certificate, CRL and OCSP revocation settings -- created inside the `authentication` scope (semp/all-cfg-tps-org.cli:410-500). A `create` for a name the target already has is rejected, so ClearExistingNested emits `no client-certificate-authority <name>` first, but ONLY for names the target carries AND this artifact re-creates: trust material the target's own operator added is not this import's to remove. |
| `Configure Web Manager` | Wizard/customization flags and the HTTP redirect port (semp/all.cli:346-348). |
| `Create Authentication` | Sets default access levels and the basic auth type (semp/all.cli:365, 371); a wrong value can lock an operator out of the broker, so review it before applying. |
| `Configure Management Message Vpn` | `no management-message-vpn` (semp/all.cli:377); cheap and reversible. |

## Applied, minus some lines

| Section | Why |
|---|---|
| `Configure Routing` | Drops `interface "intf0"` (semp/all.cli:135): that name is host hardware, and the sections that create it are skip-appliance, so the target may have no interface by that name. |
| `Create hardware` | Empty on a software broker; classified with Message Spool for the case it is not. |
| `Message Spool` | Applies the event thresholds and defragment schedule; drops `max-spool-usage`, `spool-sync mode` and the shutdown/no-shutdown pair (semp/all.cli:186, 187, 192, 200) that bounce the message spool -- the guaranteed-messaging data path -- and that on Kubernetes are sized by the CR and the PVC instead. |

## Never applied

| Section | Why |
|---|---|
| `Configure System` | On Kubernetes the CR sets connection/subscription scaling at first boot (semp/all.cli:32) and this is not durable: it reverts on the next pod recreate. |
| `Configure SEMP Service` | Lines 65-70 shut SEMP down to change its own port, and SEMP is the channel import runs over, so applying this section would sever the connection the import needs to finish. |
| `Configure Matelink Service` | Bounces the HA mate link (semp/all.cli:86, 88) to apply its port. |
| `Configure Redundancy Service` | Bounces HA (semp/all.cli:91, 93) to apply its listen port. |
| `Configure Router Name` | Node identity (semp/all.cli:131); overwriting the target's router-name breaks HA pairing, DMR and MNR. |
| `Configure Hostname` | The broker's own hostname (`hostname "<name>" defer`); node identity like Router Name above, so copying the source's onto a target renames that broker to its source. |
| `Configure Redundancy` | Ends at `redundancy shutdown` (semp/all.cli:168) with no matching un-shutdown, so a standalone capture applied to an HA pair leaves redundancy down; it also sets `active-standby-role primary` (semp/all.cli:159), which can split the group on the wrong node. |
| `Configure Config Sync` | Line 279 shuts config-sync down, which is what propagates VPN configuration to the HA mate. |
| `Create Redundancy PSK` | The HA pre-shared key (semp/all.cli:342); the Solace operator ignores preSharedAuthKeySecret updates once a group exists, so this breaks the group with no automatic repair. |

## Appliance only (skipped on a software broker)

| Section | Why |
|---|---|
| `Configure Ethernet Interfaces` | Physical Ethernet interface configuration; empty on a software broker. |
| `Configure Lag Interfaces` | Link-aggregation interface configuration; empty on a software broker. |
| `Configure SolOsPhy Interfaces` | Host NIC identity (semp/all.cli:45-48); on Kubernetes the interface belongs to the pod, not to an imported artifact. |
| `Configure ip vrf` | Host addressing under the management/matelink/msg-backbone VRFs (semp/all.cli:50-58, 60, 62); matched by prefix since the broker names the VRF in its own section comment. |
| `Configure DNS` | DNS server configuration; empty on a software broker. |
| `Disk` | Physical disk configuration; empty on a software broker. |
| `Power Redundancy` | Physical power-supply redundancy (`hardware power-redundancy "1+1"`); an appliance chassis property with nothing to configure on a software broker. |
| `Topic Routing` | `hardware topic-routing acl-topic-matching-mode`; issued as a hardware command and present on no software capture, so a software target has nowhere to apply it. |
| `Configure Clock` | Clock/timezone configuration; empty on a software broker. |
| `Configure Clock Synchronization` | NTP/clock-sync configuration; empty on a software broker. |
| `Configure SNMP` | SNMP agent and trap configuration; empty on a software broker. |

## Sections that interrupt a service

These are applied, and applying them is service-affecting. `import-config` does
not ask before doing so at broker scope.

| Section | Interrupts |
|---|---|
| `Configure SMF Service` | all messaging |
| `Configure HealthCheck Service` | the kubelet readiness probe |
| `Configure Service` | msg-backbone, the data path |
| `Create logging` | this CLI session |
| `Configure SSL` | live TLS sessions |

