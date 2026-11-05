#!/bin/bash

## Generates a CA key+cert, then a server key+CSR, and signs the server cert.
## Always uses ECDSA (default curve P-521). Configuration comes from either:
##   - a YAML file:   gen-cert.sh -c config.yaml
##   - interactive prompts:   gen-cert.sh   (no arguments)

umask 077          # generated keys/passwords are private by default
set -o pipefail

US=$'\x1f'         # field separator used between this script and the YAML parser

#============================== DEFAULTS ======================================#
# Fallback values. Used (a) as defaults for any field a YAML file omits, and
# (b) as the pre-filled [defaults] shown in interactive mode.
# Single source of truth for defaults.
DEF_EC_CURVE="P-521"         # P-256 | P-384 | P-521 (strongest)

DEF_CA_GENERATE="true"
DEF_CA_PREFIX="CA"
DEF_CA_DAYS="3650"
DEF_CA_CN="HAFIO Services CA Authority"
DEF_CA_C="SG"
DEF_CA_ST=""
DEF_CA_L="Singapore"
DEF_CA_O="Solace"
DEF_CA_OU="IT"
DEF_CA_EMAIL=""
DEF_CA_SAN=("ca.hafio.services")

DEF_SVR_PREFIX="server"
DEF_SVR_DAYS="3650"
DEF_SVR_CN="holace"
DEF_SVR_C=""
DEF_SVR_ST=""
DEF_SVR_L=""
DEF_SVR_O=""
DEF_SVR_OU=""
DEF_SVR_UID=""
DEF_SVR_EMAIL=""
DEF_SVR_SAN=("*.hafio.services" "solace.hafio.services")
#==============================================================================#

echoUsage() {
	cat <<EOF
Usage:
  gen-cert.sh -c <config.yaml>   Generate certs from a YAML config file.
  gen-cert.sh                    Interactive mode (prompts for every field).
  gen-cert.sh -h                 Show this help.

Outputs (per run):
  <ca-prefix>.key  <ca-prefix>.crt   [<ca-prefix>.pass]
  <server-prefix>.key  <server-prefix>.csr  <server-prefix>.crt  [<server-prefix>.pass]

A .pass file is written only when that key has a password.
SAN entries are plain hostnames or IPs; the script adds DNS:/IP: automatically.

Docs: https://github.com/hafio/solace-psg/tree/main/ssl-certs
EOF
	exit "${1:-1}"
}

# --- small helpers ----------------------------------------------------------
concatCnIfNotBlank() { [[ -n "$2" ]] && printf '/%s=%s' "$1" "$2"; return 0; }

checkExit() {
	if [[ $1 -ne 0 ]] || [[ ! -f "$2" ]]; then
		echo "Failed"; exit 1
	else
		echo "Ok : $2"
	fi
}

# Classify a SAN entry and prefix it for openssl.
sanPrefix() {
	local v="$1"
	if [[ "$v" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]] || [[ "$v" == *:* ]]; then
		printf 'IP:%s' "$v"
	else
		printf 'DNS:%s' "$v"
	fi
}

# Join a SAN array into a comma list: DNS:a,DNS:b,IP:1.2.3.4
buildSanList() {
	local out="" e
	for e in "$@"; do
		out+="${out:+,}$(sanPrefix "$e")"
	done
	printf '%s' "$out"
}

#======================= YAML PARSER (python3 + PyYAML) =======================#
# Emits records, fields separated by US (\x1f):
#   SCALAR <US> key <US> value   |   CASAN <US> value   |   SVRSAN <US> value
parseYaml() {
	local file="$1"
	if ! command -v python3 >/dev/null 2>&1; then
		echo "Error: python3 is required to read YAML config (PyYAML)." >&2
		echo "Install it, or run interactively with no arguments." >&2
		exit 1
	fi
	US="$US" python3 - "$file" <<'PY'
import os, sys, yaml
US = os.environ["US"]
path = sys.argv[1]
try:
    with open(path) as f:
        doc = yaml.safe_load(f) or {}
except FileNotFoundError:
    sys.exit(f"Error: config file not found: {path}")
except yaml.YAMLError as e:
    sys.exit(f"Error: invalid YAML in {path}: {e}")

def get(d, *keys, default=None):
    for k in keys:
        if not isinstance(d, dict): return default
        d = d.get(k, default if k == keys[-1] else {})
    return d

errs = []
curve = (get(doc, "algorithm", "ec_curve") or "").upper()
if curve and curve not in ("P-256", "P-384", "P-521"):
    errs.append(f"algorithm.ec_curve must be P-256, P-384, or P-521 (got '{curve}')")
ca_gen = get(doc, "ca", "generate")
if not get(doc, "server", "subject", "common_name"):
    errs.append("server.subject.common_name is required")
if ca_gen is not False and not get(doc, "ca", "subject", "common_name"):
    errs.append("ca.subject.common_name is required when ca.generate is true")
if errs:
    sys.exit("Error: invalid config:\n  - " + "\n  - ".join(errs))

def emit(key, val):
    if val is None: val = ""
    if isinstance(val, bool): val = "true" if val else "false"
    print(f"SCALAR{US}{key}{US}{val}")

emit("EC_CURVE", curve)
emit("CA_GENERATE", ca_gen)
emit("CA_PREFIX",   get(doc, "ca", "file_prefix"))
emit("CA_DAYS",     get(doc, "ca", "days"))
emit("CA_PASS",     get(doc, "ca", "password"))
emit("CA_CN",    get(doc, "ca", "subject", "common_name"))
emit("CA_C",     get(doc, "ca", "subject", "country"))
emit("CA_ST",    get(doc, "ca", "subject", "state"))
emit("CA_L",     get(doc, "ca", "subject", "locality"))
emit("CA_O",     get(doc, "ca", "subject", "organization"))
emit("CA_OU",    get(doc, "ca", "subject", "org_unit"))
emit("CA_EMAIL", get(doc, "ca", "subject", "email"))
emit("SVR_PREFIX", get(doc, "server", "file_prefix"))
emit("SVR_DAYS",   get(doc, "server", "days"))
emit("SVR_PASS",   get(doc, "server", "password"))
emit("SVR_CN",    get(doc, "server", "subject", "common_name"))
emit("SVR_C",     get(doc, "server", "subject", "country"))
emit("SVR_ST",    get(doc, "server", "subject", "state"))
emit("SVR_L",     get(doc, "server", "subject", "locality"))
emit("SVR_O",     get(doc, "server", "subject", "organization"))
emit("SVR_OU",    get(doc, "server", "subject", "org_unit"))
emit("SVR_UID",   get(doc, "server", "subject", "uid"))
emit("SVR_EMAIL", get(doc, "server", "subject", "email"))
for s in (get(doc, "ca", "san") or []):
    print(f"CASAN{US}{s}")
for s in (get(doc, "server", "san") or []):
    print(f"SVRSAN{US}{s}")
PY
}

loadFromYaml() {
	local file="$1" kind a b out
	declare -gA Y=()
	CA_SAN_IN=(); SVR_SAN_IN=()
	if ! out="$(parseYaml "$file")"; then
		exit 1   # parseYaml already printed the reason
	fi
	while IFS="$US" read -r kind a b; do
		case "$kind" in
			SCALAR) Y["$a"]="$b" ;;
			CASAN)  CA_SAN_IN+=("$a") ;;
			SVRSAN) SVR_SAN_IN+=("$a") ;;
		esac
	done <<< "$out"

	EC_CURVE="${Y[EC_CURVE]:-$DEF_EC_CURVE}"
	CA_GENERATE="${Y[CA_GENERATE]:-$DEF_CA_GENERATE}"
	CA_PREFIX="${Y[CA_PREFIX]:-$DEF_CA_PREFIX}"
	CA_DAYS="${Y[CA_DAYS]:-$DEF_CA_DAYS}"
	CA_PASS="${Y[CA_PASS]}"
	CA_CN="${Y[CA_CN]:-$DEF_CA_CN}"
	CA_C="${Y[CA_C]:-$DEF_CA_C}"; CA_ST="${Y[CA_ST]}"; CA_L="${Y[CA_L]:-$DEF_CA_L}"
	CA_O="${Y[CA_O]:-$DEF_CA_O}"; CA_OU="${Y[CA_OU]:-$DEF_CA_OU}"; CA_EMAIL="${Y[CA_EMAIL]}"
	SVR_PREFIX="${Y[SVR_PREFIX]:-$DEF_SVR_PREFIX}"
	SVR_DAYS="${Y[SVR_DAYS]:-$DEF_SVR_DAYS}"
	SVR_PASS="${Y[SVR_PASS]}"
	SVR_CN="${Y[SVR_CN]:-$DEF_SVR_CN}"
	SVR_C="${Y[SVR_C]:-$DEF_SVR_C}"; SVR_ST="${Y[SVR_ST]}"; SVR_L="${Y[SVR_L]:-$DEF_SVR_L}"
	SVR_O="${Y[SVR_O]:-$DEF_SVR_O}"; SVR_OU="${Y[SVR_OU]:-$DEF_SVR_OU}"
	SVR_UID="${Y[SVR_UID]}"; SVR_EMAIL="${Y[SVR_EMAIL]}"

	if [[ ${#CA_SAN_IN[@]}  -gt 0 ]]; then CA_SAN=("${CA_SAN_IN[@]}");  else CA_SAN=("${DEF_CA_SAN[@]}");  fi
	if [[ ${#SVR_SAN_IN[@]} -gt 0 ]]; then SVR_SAN=("${SVR_SAN_IN[@]}"); else SVR_SAN=("${DEF_SVR_SAN[@]}"); fi
}

#============================== INTERACTIVE ===================================#
ask() {
	local __var="$1" __prompt="$2" __def="$3" __in
	if [[ -n "$__def" ]]; then
		read -r -p "$__prompt [$__def]: " __in
	else
		read -r -p "$__prompt: " __in
	fi
	printf -v "$__var" '%s' "${__in:-$__def}"
}

askYesNo() {
	local __var="$1" __prompt="$2" __def="$3" __in __dl
	[[ "$__def" == "true" ]] && __dl="Y/n" || __dl="y/N"
	read -r -p "$__prompt [$__dl]: " __in
	__in="${__in:-$__def}"
	case "${__in,,}" in
		y|yes|true) printf -v "$__var" 'true' ;;
		*)          printf -v "$__var" 'false' ;;
	esac
}

askPass() {
	local __var="$1" __prompt="$2" __in
	read -r -s -p "$__prompt (blank = no password): " __in; echo
	printf -v "$__var" '%s' "$__in"
}

askSan() {
	local __arr="$1"; shift
	local defaults=("$@") entries=() line
	echo "Enter SAN entries one per line (hostname or IP). Blank line to finish."
	echo "  defaults: ${defaults[*]}"
	while true; do
		read -r -p "  SAN> " line
		[[ -z "$line" ]] && break
		entries+=("$line")
	done
	if [[ ${#entries[@]} -eq 0 ]]; then
		eval "$__arr=(\"\${defaults[@]}\")"
	else
		eval "$__arr=(\"\${entries[@]}\")"
	fi
}

runInteractive() {
	echo "=== gen-cert interactive setup (press Enter to accept [default]) ==="
	echo
	echo "-- Algorithm (ECDSA) --"
	ask EC_CURVE "EC curve (P-256/P-384/P-521)" "$DEF_EC_CURVE"

	echo; echo "-- Certificate Authority --"
	askYesNo CA_GENERATE "Generate a new CA? (No = reuse existing files)" "$DEF_CA_GENERATE"
	ask CA_PREFIX "CA file prefix" "$DEF_CA_PREFIX"
	if [[ "$CA_GENERATE" == "true" ]]; then
		ask CA_CN    "CA common name"   "$DEF_CA_CN"
		ask CA_C     "CA country"       "$DEF_CA_C"
		ask CA_ST    "CA state"         "$DEF_CA_ST"
		ask CA_L     "CA locality"      "$DEF_CA_L"
		ask CA_O     "CA organization"  "$DEF_CA_O"
		ask CA_OU    "CA org unit"      "$DEF_CA_OU"
		ask CA_EMAIL "CA email"         "$DEF_CA_EMAIL"
		ask CA_DAYS  "CA validity days" "$DEF_CA_DAYS"
		askSan CA_SAN "${DEF_CA_SAN[@]}"
		askPass CA_PASS "CA key password"
	else
		CA_CN="$DEF_CA_CN"; CA_C="$DEF_CA_C"; CA_ST="$DEF_CA_ST"; CA_L="$DEF_CA_L"
		CA_O="$DEF_CA_O"; CA_OU="$DEF_CA_OU"; CA_EMAIL="$DEF_CA_EMAIL"; CA_DAYS="$DEF_CA_DAYS"
		CA_SAN=("${DEF_CA_SAN[@]}")
		askPass CA_PASS "Existing CA key password"
	fi

	echo; echo "-- Server --"
	ask SVR_PREFIX "Server file prefix" "$DEF_SVR_PREFIX"
	ask SVR_CN    "Server common name"  "$DEF_SVR_CN"
	ask SVR_C     "Server country"      "$DEF_SVR_C"
	ask SVR_ST    "Server state"        "$DEF_SVR_ST"
	ask SVR_L     "Server locality"     "$DEF_SVR_L"
	ask SVR_O     "Server organization" "$DEF_SVR_O"
	ask SVR_OU    "Server org unit"     "$DEF_SVR_OU"
	ask SVR_UID   "Server UID"          "$DEF_SVR_UID"
	ask SVR_EMAIL "Server email"        "$DEF_SVR_EMAIL"
	ask SVR_DAYS  "Server validity days" "$DEF_SVR_DAYS"
	askSan SVR_SAN "${DEF_SVR_SAN[@]}"
	askPass SVR_PASS "Server key password"
	echo
}

#============================== ARG PARSING ===================================#
CONFIG_FILE=""
case "${1:-}" in
	-h|--help) echoUsage 0 ;;
	-c|--config)
		[[ -z "${2:-}" ]] && { echo "$1: missing file argument"; exit 1; }
		CONFIG_FILE="$2" ;;
	"" ) : ;;
	* ) echoUsage ;;
esac

if [[ -n "$CONFIG_FILE" ]]; then
	loadFromYaml "$CONFIG_FILE"
else
	runInteractive
fi

#======================= ALGORITHM + KEY USAGE ================================#
EC_CURVE="${EC_CURVE^^}"   # normalize p-521 -> P-521
case "$EC_CURVE" in
	P-256) DIGEST="sha256" ;;
	P-384) DIGEST="sha384" ;;
	P-521) DIGEST="sha512" ;;
	*) echo "Error: unsupported EC curve '$EC_CURVE' (use P-256, P-384, or P-521)"; exit 1 ;;
esac
# Pair the signature hash to the curve (P-256/sha256, P-384/sha384, P-521/sha512).
KEY_ARGS=(-newkey ec -pkeyopt "ec_paramgen_curve:${EC_CURVE}")
# EC keys can only sign, so the server keyUsage has no keyEncipherment.
SVR_KEY_USAGE="critical,digitalSignature"

CA_PASS_ARGS=(-noenc);  [[ -n "$CA_PASS"  ]] && { export CA_PASS;  CA_PASS_ARGS=(-passout env:CA_PASS); }
SVR_PASS_ARGS=(-noenc); [[ -n "$SVR_PASS" ]] && { export SVR_PASS; SVR_PASS_ARGS=(-passout env:SVR_PASS); }
CASIGN_PASS_ARGS=();    [[ -n "$CA_PASS"  ]] && CASIGN_PASS_ARGS=(-passin env:CA_PASS)

CA_SAN_LIST="$(buildSanList "${CA_SAN[@]}")"
SVR_SAN_LIST="$(buildSanList "${SVR_SAN[@]}")"
CA_SAN_ARGS=();  [[ -n "$CA_SAN_LIST"  ]] && CA_SAN_ARGS=(-addext "subjectAltName=${CA_SAN_LIST}")
SVR_SAN_ARGS=(); [[ -n "$SVR_SAN_LIST" ]] && SVR_SAN_ARGS=(-addext "subjectAltName=${SVR_SAN_LIST}")

RSUBJ="$(concatCnIfNotBlank CN "$CA_CN")$(concatCnIfNotBlank C "$CA_C")$(concatCnIfNotBlank ST "$CA_ST")$(concatCnIfNotBlank L "$CA_L")$(concatCnIfNotBlank O "$CA_O")$(concatCnIfNotBlank OU "$CA_OU")$(concatCnIfNotBlank emailAddress "$CA_EMAIL")"
SSUBJ="$(concatCnIfNotBlank CN "$SVR_CN")$(concatCnIfNotBlank C "$SVR_C")$(concatCnIfNotBlank ST "$SVR_ST")$(concatCnIfNotBlank L "$SVR_L")$(concatCnIfNotBlank O "$SVR_O")$(concatCnIfNotBlank OU "$SVR_OU")$(concatCnIfNotBlank UID "$SVR_UID")$(concatCnIfNotBlank emailAddress "$SVR_EMAIL")"

#============================== GENERATION ====================================#
if [[ "$CA_GENERATE" == "true" ]]; then
	echo -n "Generating Root Certificate and Key..."
	openssl req -x509 "${KEY_ARGS[@]}" -"$DIGEST" -days "$CA_DAYS" \
		-keyout "${CA_PREFIX}.key" -out "${CA_PREFIX}.crt" \
		"${CA_PASS_ARGS[@]}" -subj "$RSUBJ" \
		"${CA_SAN_ARGS[@]}" \
		-addext "basicConstraints=critical,CA:TRUE,pathlen:1" \
		-addext "keyUsage=critical,keyCertSign,cRLSign,digitalSignature" 2>/dev/null
	checkExit $? "${CA_PREFIX}.crt"
	chmod 600 "${CA_PREFIX}.key"
	[[ -n "$CA_PASS" ]] && { echo "$CA_PASS" > "${CA_PREFIX}.pass"; chmod 600 "${CA_PREFIX}.pass"; }
fi

if [[ -f "${CA_PREFIX}.crt" && -f "${CA_PREFIX}.key" ]]; then
	echo -n "Generating Server CSR and Key..."
	openssl req "${KEY_ARGS[@]}" "${SVR_PASS_ARGS[@]}" \
		-keyout "${SVR_PREFIX}.key" -out "${SVR_PREFIX}.csr" \
		-subj "$SSUBJ" "${SVR_SAN_ARGS[@]}" \
		-addext "basicConstraints=critical,CA:FALSE" \
		-addext "keyUsage=${SVR_KEY_USAGE}" \
		-addext "extendedKeyUsage=critical,serverAuth" 2>/dev/null
	checkExit $? "${SVR_PREFIX}.csr"
	chmod 600 "${SVR_PREFIX}.key"
	[[ -n "$SVR_PASS" ]] && { echo "$SVR_PASS" > "${SVR_PREFIX}.pass"; chmod 600 "${SVR_PREFIX}.pass"; }

	echo -n "Generating Server Certificate..."
	# -copy_extensions copy carries SAN/keyUsage from our own CSR. Safe because we
	# created the CSR. Never copy extensions blindly when signing third-party CSRs.
	openssl x509 -req -in "${SVR_PREFIX}.csr" \
		-CA "${CA_PREFIX}.crt" -CAkey "${CA_PREFIX}.key" -CAcreateserial \
		"${CASIGN_PASS_ARGS[@]}" -"$DIGEST" -days "$SVR_DAYS" \
		-copy_extensions copy -out "${SVR_PREFIX}.crt" 2>/dev/null
	checkExit $? "${SVR_PREFIX}.crt"
	echo "--- Issued certificate ---"
	openssl x509 -noout -in "${SVR_PREFIX}.crt" -issuer -subject -ext subjectAltName
else
	echo "Root file(s) (${CA_PREFIX}.crt and/or ${CA_PREFIX}.key) is/are missing"
	exit 1
fi
