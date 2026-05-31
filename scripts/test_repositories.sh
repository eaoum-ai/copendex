#!/usr/bin/env sh
set -eu

GO="${GO:-go}"
GIT="${GIT:-git}"
ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
TEST_REPOS_DIR="${COSHA_TEST_REPOS_DIR:-$ROOT/.cache/test-repos}"
REPORT_DIR="${COSHA_BENCHMARK_REPORT_DIR:-$ROOT/.cache/benchmark-reports}"
BIN="${COSHA_BIN:-$ROOT/cosha}"
ACTION="${1:-smoke}"
FILTER="${2:-all}"

repos() {
	cat <<'EOF'
spring-petclinic|https://github.com/spring-projects/spring-petclinic.git|OwnerController|OwnerRepository|interface
spring-boot|https://github.com/spring-projects/spring-boot.git|SpringApplication|SpringApplication|class
spring-framework|https://github.com/spring-projects/spring-framework.git|ApplicationContext|ApplicationContext|interface
kafka|https://github.com/apache/kafka.git|KafkaConsumer|KafkaConsumer|class
hadoop|https://github.com/apache/hadoop.git|Configuration|Configuration|class
dubbo|https://github.com/apache/dubbo.git|DubboBootstrap|DubboBootstrap|class
shardingsphere|https://github.com/apache/shardingsphere.git|ShardingSphere|ShardingSphere|class
elasticsearch|https://github.com/elastic/elasticsearch.git|Elasticsearch|Elasticsearch|class
EOF
}

usage() {
	cat <<'EOF'
usage: scripts/test_repositories.sh <clone|smoke|benchmark> [repo-id|all]

Clones public Java test repositories over HTTPS into .cache/test-repos,
runs Cosha smoke checks, and optionally records benchmark output.
EOF
}

assert_contains() {
	value="$1"
	expected="$2"
	file="$WORKDIR/assert-output"
	printf '%s' "$value" > "$file"
	if ! grep -Fq -- "$expected" "$file"; then
		printf 'expected output to contain: %s\nactual output:\n%s\n' "$expected" "$value" >&2
		exit 1
	fi
	rm -f "$file"
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

repo_dir() {
	printf '%s/%s\n' "$TEST_REPOS_DIR" "$1"
}

clone_repo() {
	id="$1"
	url="$2"
	dir="$(repo_dir "$id")"

	mkdir -p "$TEST_REPOS_DIR"
	if [ ! -d "$dir/.git" ]; then
		rm -rf "$dir"
		"$GIT" clone --depth 1 "$url" "$dir"
	else
		current_url="$("$GIT" -C "$dir" remote get-url origin)"
		if [ "$current_url" != "$url" ]; then
			"$GIT" -C "$dir" remote set-url origin "$url"
		fi
		"$GIT" -C "$dir" fetch --depth 1 origin
		default_branch="$("$GIT" -C "$dir" remote show origin | sed -n 's/.*HEAD branch: //p')"
		"$GIT" -C "$dir" checkout -q "$default_branch"
		"$GIT" -C "$dir" reset --hard -q "origin/$default_branch"
	fi

	"$GIT" -C "$dir" clean -ffdq
	rm -rf "$dir/.cosha"
	printf '%s %s\n' "$id" "$("$GIT" -C "$dir" rev-parse HEAD)"
}

smoke_repo() {
	id="$1"
	query="$2"
	symbol="$3"
	kind="$4"
	dir="$(repo_dir "$id")"

	WORKDIR="$(mktemp -d)"
	trap 'rm -rf "$WORKDIR"' EXIT INT TERM

	rm -rf "$dir/.cosha"
	cd "$dir"

	init_output="$("$BIN" init)"
	assert_contains "$init_output" "Initialized .cosha/config.yaml"
	assert_file "$dir/.cosha/config.yaml"

	detect_output="$("$BIN" detect)"
	assert_contains "$detect_output" "Java repository: true"
	assert_contains "$detect_output" "Contains Java source: true"

	"$BIN" detect --json > "$WORKDIR/detect.json"
	assert_valid_json "$WORKDIR/detect.json"
	assert_contains "$(cat "$WORKDIR/detect.json")" '"isJavaRepository": true'

	index_output="$("$BIN" index)"
	assert_contains "$index_output" "Indexed "
	assert_file "$dir/.cosha/index/cosha.db"

	"$BIN" stats --json > "$WORKDIR/stats.json"
	assert_valid_json "$WORKDIR/stats.json"
	assert_contains "$(cat "$WORKDIR/stats.json")" '"fileCount":'
	assert_contains "$(cat "$WORKDIR/stats.json")" '"java":'

	"$BIN" search "$query" --json > "$WORKDIR/search.json"
	assert_valid_json "$WORKDIR/search.json"
	assert_contains "$(cat "$WORKDIR/search.json")" "\"$query\""

	"$BIN" symbols "$symbol" --kind "$kind" --json > "$WORKDIR/symbols.json"
	assert_valid_json "$WORKDIR/symbols.json"
	assert_contains "$(cat "$WORKDIR/symbols.json")" "\"$symbol\""

	rm -rf "$WORKDIR"
	trap - EXIT INT TERM
	printf 'smoke passed: %s\n' "$id"
}

benchmark_repo() {
	id="$1"
	query="$2"
	symbol="$3"
	dir="$(repo_dir "$id")"
	report="$REPORT_DIR/$id.md"

	mkdir -p "$REPORT_DIR"
	rm -rf "$dir/.cosha"

	if command -v hyperfine >/dev/null 2>&1; then
		{
			printf '# %s\n\n' "$id"
			printf -- '- Repository: `%s`\n' "$dir"
			printf -- '- Commit: `%s`\n\n' "$("$GIT" -C "$dir" rev-parse HEAD)"
			printf '```text\n'
			(
				cd "$dir"
				hyperfine --warmup 1 --runs "${COSHA_BENCHMARK_INDEX_RUNS:-1}" \
					"$BIN index --rebuild"
				hyperfine --warmup 1 --runs "${COSHA_BENCHMARK_QUERY_RUNS:-3}" \
					"$BIN detect >/dev/null" \
					"$BIN stats >/dev/null" \
					"$BIN stats --json >/dev/null" \
					"$BIN search $query >/dev/null" \
					"$BIN symbols $symbol >/dev/null"
			)
			printf '```\n'
		} > "$report"
	else
		{
			printf '# %s\n\n' "$id"
			printf -- '- Repository: `%s`\n' "$dir"
			printf -- '- Commit: `%s`\n\n' "$("$GIT" -C "$dir" rev-parse HEAD)"
			printf 'hyperfine is not installed; benchmark skipped.\n'
		} > "$report"
	fi

	printf 'benchmark report: %s\n' "$report"
}

run_selected() {
	while IFS='|' read -r id url query symbol kind; do
		if [ "$FILTER" != "all" ] && [ "$FILTER" != "$id" ]; then
			continue
		fi

		case "$ACTION" in
			clone)
				clone_repo "$id" "$url"
				;;
			smoke)
				clone_repo "$id" "$url" >/dev/null
				smoke_repo "$id" "$query" "$symbol" "$kind"
				;;
			benchmark)
				clone_repo "$id" "$url" >/dev/null
				smoke_repo "$id" "$query" "$symbol" "$kind"
				benchmark_repo "$id" "$query" "$symbol"
				;;
			*)
				usage >&2
				exit 2
				;;
		esac
	done <<EOF
$(repos)
EOF
}

case "$ACTION" in
	clone|smoke|benchmark)
		run_selected
		;;
	-h|--help|help)
		usage
		;;
	*)
		usage >&2
		exit 2
		;;
esac
