// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies Tree-sitter-backed Java symbol extraction.
package java

import (
	"reflect"
	"testing"

	"github.com/eaoum-ai/copendex/internal/index"
)

func TestExtractJavaSymbols(t *testing.T) {
	content := []byte(`package com.example.security;

import java.util.List;
import org.springframework.stereotype.Service;

@Service
public class AuthorizationService {
  @Deprecated
  public boolean canAccess(String user) {
    return true;
  }
}
`)
	symbols := Extract("src/main/java/com/example/security/AuthorizationService.java", content)
	seen := symbolSet(symbols)
	for _, want := range []string{
		"package:com.example.security",
		"import:java.util.List",
		"import:org.springframework.stereotype.Service",
		"annotation:Service",
		"class:AuthorizationService",
		"annotation:Deprecated",
		"method:canAccess",
	} {
		if !seen[want] {
			t.Fatalf("missing symbol %s in %#v", want, seen)
		}
	}
	assertSymbol(t, symbols, "class", "AuthorizationService", 7, []string{"Service"})
	assertSymbol(t, symbols, "method", "canAccess", 9, []string{"Deprecated"})
}

func TestExtractConstructorsNestedTypesInterfacesAndEnums(t *testing.T) {
	content := []byte(`package com.example.orders;

import java.util.List;

@Aggregate
public class OrderService {
  private final Repository repository;

  @Inject
  public OrderService(Repository repository) {
    this.repository = repository;
  }

  public Order find(String id) {
    return repository.find(id);
  }

  public Order find(String id, boolean includeDeleted) {
    return repository.find(id);
  }

  interface Repository {
    Order find(String id);
  }

  enum Status {
    OPEN,
    CLOSED;
  }

  static class Audit {
    void record(Order order) {}
  }
}
`)
	symbols := Extract("src/main/java/com/example/orders/OrderService.java", content)
	for _, want := range []string{
		"package:com.example.orders",
		"import:java.util.List",
		"annotation:Aggregate",
		"class:OrderService",
		"annotation:Inject",
		"constructor:OrderService",
		"method:find",
		"interface:Repository",
		"enum:Status",
		"enumConstant:OPEN",
		"enumConstant:CLOSED",
		"class:Audit",
		"method:record",
	} {
		if !symbolSet(symbols)[want] {
			t.Fatalf("missing symbol %s in %#v", want, symbols)
		}
	}
	assertSymbol(t, symbols, "class", "OrderService", 6, []string{"Aggregate"})
	assertSymbol(t, symbols, "constructor", "OrderService", 10, []string{"Inject"})
	if got := countSymbols(symbols, "method", "find"); got != 3 {
		t.Fatalf("method find count = %d, want 3", got)
	}
}

func symbolSet(symbols []index.Symbol) map[string]bool {
	seen := map[string]bool{}
	for _, sym := range symbols {
		seen[sym.Kind+":"+sym.Name] = true
	}
	return seen
}

func assertSymbol(t *testing.T, symbols []index.Symbol, kind, name string, line int, annotations []string) {
	t.Helper()
	for _, sym := range symbols {
		if sym.Kind != kind || sym.Name != name {
			continue
		}
		if sym.Line != line {
			t.Fatalf("%s:%s line = %d, want %d", kind, name, sym.Line, line)
		}
		if !reflect.DeepEqual(sym.Annotations, annotations) {
			t.Fatalf("%s:%s annotations = %#v, want %#v", kind, name, sym.Annotations, annotations)
		}
		return
	}
	t.Fatalf("missing symbol %s:%s", kind, name)
}

func countSymbols(symbols []index.Symbol, kind, name string) int {
	var count int
	for _, sym := range symbols {
		if sym.Kind == kind && sym.Name == name {
			count++
		}
	}
	return count
}
