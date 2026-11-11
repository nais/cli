package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi/gql"
)

type suppressionTransport func(*http.Request) (*http.Response, error)

func (f suppressionTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func suppressionHTTPClient(t *testing.T, response map[string]any, inspect func(*http.Request)) graphql.Client {
	t.Helper()
	body, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return graphql.NewClient("https://nais.example.test/graphql", &http.Client{Transport: suppressionTransport(func(r *http.Request) (*http.Response, error) {
		if inspect != nil {
			inspect(r)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
	})})
}

func TestSuppressionTypedOperations(t *testing.T) {
	target := SuppressionTarget{
		FindingID: "finding", AdvisoryID: "CVE-2026-1234", PackagePurl: "pkg:npm/example@1",
		WorkloadID: "app", Team: "team", Environment: "dev", Workload: "app", WorkloadType: "application",
		ImageID: "image", ImageName: "registry.example/app", ImageTag: "v1",
	}
	for _, write := range []bool{false, true} {
		t.Run(map[bool]string{false: "preview", true: "write"}[write], func(t *testing.T) {
			root := "vulnerabilitySuppressionPreview"
			response := map[string]any{"target": target, "affectedWorkloads": []any{}, "scopeFingerprint": "fingerprint", "includesFutureMatches": true, "warning": "shared scope"}
			if write {
				root = "applyVulnerabilitySuppression"
				response = map[string]any{"persisted": true, "indeterminate": false, "state": "IN_TRIAGE", "scopeFingerprint": "fingerprint"}
			}
			calls := 0
			c := suppressionHTTPClient(t, map[string]any{"data": map[string]any{root: response}}, func(r *http.Request) {
				calls++
				var request struct {
					Query     string                     `json:"query"`
					Variables map[string]json.RawMessage `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(request.Query, target.PackagePurl) || strings.Contains(request.Query, "updateImageVulnerability") {
					t.Fatal("untyped or legacy mutation")
				}
				if write {
					var input SuppressionInput
					if err := json.Unmarshal(request.Variables["input"], &input); err != nil {
						t.Fatal(err)
					}
					if !input.Confirmed || input.Target.FindingID != target.FindingID || input.Decision.EvidenceReference == nil {
						t.Fatal("decision missing from typed input")
					}
				} else {
					var input SuppressionTarget
					if err := json.Unmarshal(request.Variables["target"], &input); err != nil || input != target {
						t.Fatalf("incorrect target: %+v error=%v", input, err)
					}
				}
			})
			if write {
				out, err := applySuppression(t.Context(), c, SuppressionInput{
					Target:    gql.VulnerabilitySuppressionTargetInput(target),
					Confirmed: true, ExpectedScopeFingerprint: "fingerprint", Decision: gql.VulnerabilitySuppressionDecisionInput{
						Justification: "Reviewed", EvidenceReference: new("https://example.test/evidence"), State: gql.ImageVulnerabilitySuppressionStateInTriage,
					},
				})
				if err != nil || !out.Persisted {
					t.Fatalf("out=%+v error=%v", out, err)
				}
			} else {
				out, err := fetchSuppressionPreview(t.Context(), c, target)
				if err != nil || SuppressionTarget(out.Target) != target {
					t.Fatalf("out=%+v error=%v", out, err)
				}
			}
			if calls != 1 {
				t.Fatalf("unexpected calls/retries: %d", calls)
			}
		})
	}
}

func TestSuppressionErrorsDoNotLeakOrRetry(t *testing.T) {
	calls := 0
	c := suppressionHTTPClient(t, map[string]any{"errors": []any{map[string]any{"message": "PRIVATE JUSTIFICATION"}}}, func(*http.Request) { calls++ })
	if _, err := fetchSuppressionPreview(t.Context(), c, SuppressionTarget{}); err == nil || strings.Contains(err.Error(), "PRIVATE") {
		t.Fatalf("unsafe preview error: %v", err)
	}
	if _, err := applySuppression(t.Context(), c, SuppressionInput{}); err == nil || strings.Contains(err.Error(), "PRIVATE") {
		t.Fatalf("unsafe write error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("automatic retries: %d", calls)
	}
}
