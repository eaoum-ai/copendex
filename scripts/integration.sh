#!/usr/bin/env sh
set -eu

GO="${GO:-go}"
ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
WORKDIR="$(mktemp -d)"
BIN="$WORKDIR/cosha"
REPO="$WORKDIR/repo"

cleanup() {
	rm -rf "$WORKDIR"
}
trap cleanup EXIT INT TERM

assert_contains() {
	value="$1"
	expected="$2"
	if ! printf '%s' "$value" | grep -Fq -- "$expected"; then
		printf 'expected output to contain: %s\nactual output:\n%s\n' "$expected" "$value" >&2
		exit 1
	fi
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

GOCACHE="${GOCACHE:-$ROOT/.cache/go-build}" GOMODCACHE="${GOMODCACHE:-$ROOT/.cache/gomod}" "$GO" build -o "$BIN" "$ROOT/cmd/cosha"

mkdir -p "$REPO/src/main/java/com/example/web"
mkdir -p "$REPO/src/main/java/com/example/service"
mkdir -p "$REPO/src/main/java/com/example/repository"
mkdir -p "$REPO/src/main/java/com/example/model"
mkdir -p "$REPO/src/test/java/com/example/web"

cat > "$REPO/pom.xml" <<'XML'
<project>
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.example</groupId>
  <artifactId>cosha-integration</artifactId>
  <version>1.0.0</version>
</project>
XML

cat > "$REPO/src/main/java/com/example/web/OwnerController.java" <<'JAVA'
package com.example.web;

import com.example.model.Owner;
import com.example.service.OwnerService;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;

@Controller
public class OwnerController {
    private final OwnerService ownerService;

    public OwnerController(OwnerService ownerService) {
        this.ownerService = ownerService;
    }

    @GetMapping("/owners")
    public Owner findOwner(String id) {
        return ownerService.findOwner(id);
    }
}
JAVA

cat > "$REPO/src/main/java/com/example/service/OwnerService.java" <<'JAVA'
package com.example.service;

import com.example.model.Owner;
import com.example.repository.OwnerRepository;
import org.springframework.stereotype.Service;

@Service
public class OwnerService {
    private final OwnerRepository ownerRepository;

    public OwnerService(OwnerRepository ownerRepository) {
        this.ownerRepository = ownerRepository;
    }

    public Owner findOwner(String id) {
        return ownerRepository.findById(id);
    }
}
JAVA

cat > "$REPO/src/main/java/com/example/repository/OwnerRepository.java" <<'JAVA'
package com.example.repository;

import com.example.model.Owner;
import org.springframework.stereotype.Repository;

@Repository
public interface OwnerRepository {
    Owner findById(String id);
}
JAVA

cat > "$REPO/src/main/java/com/example/model/Owner.java" <<'JAVA'
package com.example.model;

public record Owner(String id, String name) {
}
JAVA

cat > "$REPO/src/test/java/com/example/web/OwnerControllerTest.java" <<'JAVA'
package com.example.web;

class OwnerControllerTest {
    void findsOwner() {
    }
}
JAVA

cd "$REPO"

help_output="$("$BIN" --help)"
assert_contains "$help_output" "Usage:"
assert_contains "$help_output" "cosha [command]"

init_output="$("$BIN" init)"
assert_contains "$init_output" "Initialized .cosha/config.yaml"
assert_file "$REPO/.cosha/config.yaml"

detect_output="$("$BIN" detect)"
assert_contains "$detect_output" "Java repository: true"
assert_contains "$detect_output" "Contains Java source: true"
assert_contains "$detect_output" "Java source files: 5"
assert_contains "$detect_output" "pom.xml"

"$BIN" detect --json > "$WORKDIR/detect.json"
assert_valid_json "$WORKDIR/detect.json"
assert_contains "$(cat "$WORKDIR/detect.json")" '"isJavaRepository": true'

index_output="$("$BIN" index)"
assert_contains "$index_output" "Indexed 5 files"
assert_file "$REPO/.cosha/index/cosha.db"

if "$BIN" index > "$WORKDIR/reindex.out" 2> "$WORKDIR/reindex.err"; then
	printf 'expected repeated index without --rebuild to fail\n' >&2
	exit 1
fi
assert_contains "$(cat "$WORKDIR/reindex.err")" "index is already built"
assert_contains "$(cat "$WORKDIR/reindex.err")" "--rebuild or -r"

rebuild_output="$("$BIN" index -r)"
assert_contains "$rebuild_output" "Indexed 5 files"

stats_output="$("$BIN" stats)"
assert_contains "$stats_output" "Files: 5"
assert_contains "$stats_output" "Languages: 1"

"$BIN" stats --json > "$WORKDIR/stats.json"
assert_valid_json "$WORKDIR/stats.json"
assert_contains "$(cat "$WORKDIR/stats.json")" '"fileCount": 5'
assert_contains "$(cat "$WORKDIR/stats.json")" '"java": 5'

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
assert_contains "$(cat "$WORKDIR/symbols.json")" '"OwnerService"'

ui_output="$("$BIN" ui)"
assert_contains "$ui_output" "Wrote Cosha UI to"
assert_file "$REPO/.cosha/ui/index.html"
assert_contains "$(cat "$REPO/.cosha/ui/index.html")" "OwnerController"

printf 'integration tests passed\n'
