#!/usr/bin/env sh
set -eu

GO="${GO:-go}"
GIT="${GIT:-git}"
ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
WORKDIR="$(mktemp -d)"
BIN="$WORKDIR/cosha"
TEST_REPOS_DIR="${COSHA_TEST_REPOS_DIR:-$ROOT/.cache/test-repos}"
PETCLINIC_DIR="$TEST_REPOS_DIR/spring-petclinic"
PETCLINIC_URL="${COSHA_TEST_REPO_SPRING_PETCLINIC_URL:-https://github.com/spring-projects/spring-petclinic.git}"
PETCLINIC_REF="${COSHA_TEST_REPO_SPRING_PETCLINIC_REF:-3c06fbfc1e42eb40802e0d0ca989bc9226755804}"

cleanup() {
	rm -rf "$WORKDIR"
}
trap cleanup EXIT INT TERM

assert_contains() {
	value="$1"
	expected="$2"
	assert_output="$WORKDIR/assert-output"
	printf '%s' "$value" > "$assert_output"
	if ! grep -Fq -- "$expected" "$assert_output"; then
		printf 'expected output to contain: %s\nactual output:\n%s\n' "$expected" "$value" >&2
		exit 1
	fi
	rm -f "$assert_output"
}

assert_file() {
	if [ ! -f "$1" ]; then
		printf 'expected file to exist: %s\n' "$1" >&2
		exit 1
	fi
}

assert_valid_json() {
	file="$1"
	"$GO" run "$ROOT/scripts/validate_json.go" "$file"
}

prepare_petclinic() {
	mkdir -p "$TEST_REPOS_DIR"

	if [ ! -d "$PETCLINIC_DIR/.git" ]; then
		rm -rf "$PETCLINIC_DIR"
		mkdir -p "$PETCLINIC_DIR"
		"$GIT" -C "$PETCLINIC_DIR" init -q
		"$GIT" -C "$PETCLINIC_DIR" remote add origin "$PETCLINIC_URL"
	fi

	current_url="$("$GIT" -C "$PETCLINIC_DIR" remote get-url origin)"
	if [ "$current_url" != "$PETCLINIC_URL" ]; then
		"$GIT" -C "$PETCLINIC_DIR" remote set-url origin "$PETCLINIC_URL"
	fi

	"$GIT" -C "$PETCLINIC_DIR" fetch --depth 1 origin "$PETCLINIC_REF"
	"$GIT" -C "$PETCLINIC_DIR" checkout -q --detach FETCH_HEAD
	"$GIT" -C "$PETCLINIC_DIR" clean -ffdq
	rm -rf "$PETCLINIC_DIR/.cosha"
}

GOCACHE="${GOCACHE:-$ROOT/.cache/go-build}" GOMODCACHE="${GOMODCACHE:-$ROOT/.cache/gomod}" "$GO" build -o "$BIN" "$ROOT/cmd/cosha"

prepare_petclinic
cd "$PETCLINIC_DIR"

help_output="$("$BIN" --help)"
assert_contains "$help_output" "Usage:"
assert_contains "$help_output" "cosha [command]"

init_output="$("$BIN" init)"
assert_contains "$init_output" "Initialized .cosha/config.yaml"
assert_file "$PETCLINIC_DIR/.cosha/config.yaml"

detect_output="$("$BIN" detect)"
assert_contains "$detect_output" "Java repository: true"
assert_contains "$detect_output" "Contains Java source: true"
assert_contains "$detect_output" "pom.xml"

"$BIN" detect --json > "$WORKDIR/detect.json"
assert_valid_json "$WORKDIR/detect.json"
assert_contains "$(cat "$WORKDIR/detect.json")" '"isJavaRepository": true'

index_output="$("$BIN" index)"
assert_contains "$index_output" "Indexed "
assert_file "$PETCLINIC_DIR/.cosha/index/cosha.db"

if "$BIN" index > "$WORKDIR/reindex.out" 2> "$WORKDIR/reindex.err"; then
	printf 'expected repeated index without --rebuild to fail\n' >&2
	exit 1
fi
assert_contains "$(cat "$WORKDIR/reindex.err")" "index is already built"
assert_contains "$(cat "$WORKDIR/reindex.err")" "--rebuild or -r"

rebuild_output="$("$BIN" index -r)"
assert_contains "$rebuild_output" "Indexed "

stats_output="$("$BIN" stats)"
assert_contains "$stats_output" "Files:"
assert_contains "$stats_output" "Symbols:"
assert_contains "$stats_output" "Languages:"

"$BIN" stats --json > "$WORKDIR/stats.json"
assert_valid_json "$WORKDIR/stats.json"
assert_contains "$(cat "$WORKDIR/stats.json")" '"fileCount":'
assert_contains "$(cat "$WORKDIR/stats.json")" '"java":'

search_output="$("$BIN" search OwnerController)"
assert_contains "$search_output" "OwnerController.java"
assert_contains "$search_output" "OwnerController"

"$BIN" search Owner --kind class,interface --json > "$WORKDIR/search.json"
assert_valid_json "$WORKDIR/search.json"
assert_contains "$(cat "$WORKDIR/search.json")" '"OwnerController"'
assert_contains "$(cat "$WORKDIR/search.json")" '"OwnerRepository"'

symbols_output="$("$BIN" symbols OwnerRepository --kind interface)"
assert_contains "$symbols_output" "OwnerRepository.java"
assert_contains "$symbols_output" "OwnerRepository"

"$BIN" symbols Owner --json > "$WORKDIR/symbols.json"
assert_valid_json "$WORKDIR/symbols.json"
assert_contains "$(cat "$WORKDIR/symbols.json")" '"OwnerController"'

ui_output="$("$BIN" ui)"
assert_contains "$ui_output" "Wrote Cosha UI to"
assert_file "$PETCLINIC_DIR/.cosha/ui/index.html"
assert_contains "$(cat "$PETCLINIC_DIR/.cosha/ui/index.html")" "OwnerController"

printf 'integration tests passed\n'
