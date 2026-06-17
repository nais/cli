package resource

import "testing"

func TestKindSupport_Supports(t *testing.T) {
	// Available as a stripped manifest and as a CRD.
	both := kindSupport{kind: "Application", strippedVersion: "v1", apiVersion: "nais.io/v1alpha1"}
	// Available only as a stripped manifest (no CRD form).
	strippedOnly := kindSupport{kind: "Valkey", strippedVersion: "v1"}

	for name, tc := range map[string]struct {
		support         kindSupport
		strippedVersion string
		apiVersion      string
		want            bool
	}{
		"matching stripped version":            {both, "v1", "", true},
		"mismatching stripped version":         {both, "v2", "", false},
		"matching apiVersion":                  {both, "", "nais.io/v1alpha1", true},
		"mismatching apiVersion":               {both, "", "nais.io/v1", false},
		"stripped-only matches stripped":       {strippedOnly, "v1", "", true},
		"stripped-only rejects any apiVersion": {strippedOnly, "", "aiven.io/v1alpha1", false},
		"neither set is never supported":       {both, "", "", false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tc.support.Supports(tc.strippedVersion, tc.apiVersion); got != tc.want {
				t.Errorf("Supports(%q, %q) = %v, want %v", tc.strippedVersion, tc.apiVersion, got, tc.want)
			}
		})
	}
}

func TestForManifest(t *testing.T) {
	for name, tc := range map[string]struct {
		kind        string
		version     string
		wantFound   bool
		wantApplier bool
		wantWaiter  bool
	}{
		"valkey v1 resolves to a mutation": {
			kind: "Valkey", version: "v1", wantFound: true, wantApplier: true, wantWaiter: false,
		},
		"opensearch v1 resolves to a mutation": {
			kind: "OpenSearch", version: "v1", wantFound: true, wantApplier: true, wantWaiter: false,
		},
		"application is not handled as a stripped manifest": {
			kind: "Application", version: "v1", wantFound: false,
		},
		"unsupported version does not resolve": {
			kind: "Valkey", version: "v2", wantFound: false,
		},
		"unknown kind does not resolve": {
			kind: "Bogus", version: "v1", wantFound: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, found := ForManifest(Manifest{Kind: tc.kind, Version: tc.version})
			if found != tc.wantFound {
				t.Fatalf("ForManifest found = %v, want %v", found, tc.wantFound)
			}
			if !found {
				return
			}
			if _, ok := r.(Applier); ok != tc.wantApplier {
				t.Errorf("Applier = %v, want %v", ok, tc.wantApplier)
			}
			if _, ok := r.(Waiter); ok != tc.wantWaiter {
				t.Errorf("Waiter = %v, want %v", ok, tc.wantWaiter)
			}
		})
	}
}

func TestForCRD(t *testing.T) {
	for name, tc := range map[string]struct {
		apiVersion string
		kind       string
		wantFound  bool
		wantWaiter bool
	}{
		"application with correct apiVersion resolves to a waiter": {
			apiVersion: "nais.io/v1alpha1", kind: "Application", wantFound: true, wantWaiter: true,
		},
		"application with wrong apiVersion does not resolve": {
			apiVersion: "nais.io/v1", kind: "Application", wantFound: false,
		},
		"valkey CRD is not handled (stripped-only)": {
			apiVersion: "aiven.io/v1alpha1", kind: "Valkey", wantFound: false,
		},
		"unknown kind does not resolve": {
			apiVersion: "nais.io/v1alpha1", kind: "Bogus", wantFound: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, found := ForCRD(tc.apiVersion, tc.kind)
			if found != tc.wantFound {
				t.Fatalf("ForCRD found = %v, want %v", found, tc.wantFound)
			}
			if !found {
				return
			}
			if _, ok := r.(Waiter); ok != tc.wantWaiter {
				t.Errorf("Waiter = %v, want %v", ok, tc.wantWaiter)
			}
		})
	}
}

func TestApplicationAPIVersion(t *testing.T) {
	r, found := ForCRD("nais.io/v1alpha1", "Application")
	if !found {
		t.Fatal("expected Application to resolve")
	}
	if got, want := r.APIVersion(), "nais.io/v1alpha1"; got != want {
		t.Errorf("APIVersion = %q, want %q", got, want)
	}
}
