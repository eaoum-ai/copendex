package java

import "testing"

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
	seen := map[string]bool{}
	for _, sym := range symbols {
		seen[sym.Kind+":"+sym.Name] = true
		if sym.Name == "AuthorizationService" {
			if sym.Line != 7 {
				t.Fatalf("class line = %d, want 7", sym.Line)
			}
			if len(sym.Annotations) != 1 || sym.Annotations[0] != "Service" {
				t.Fatalf("class annotations = %#v", sym.Annotations)
			}
		}
	}
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
}
