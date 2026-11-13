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
	id := "VUL_finding"
	reason := "Custom \"reason\".\nReviewed usage: https://example.test/review/1"
	state := gql.ImageVulnerabilitySuppressionStateNotAffected
	for _, write := range []bool{false, true} {
		t.Run(map[bool]string{false: "finding", true: "write"}[write], func(t *testing.T) {
			root := "node"
			response := map[string]any{
				"__typename": "ImageVulnerability", "id": id, "identifier": "CVE-2026-1234",
				"imageName": "registry.example/app", "package": "pkg:golang/example@v1.2.3",
			}
			if write {
				root = "updateImageVulnerability"
				response = map[string]any{"vulnerability": map[string]any{"id": id, "suppression": map[string]any{"state": state, "reason": reason}}}
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
				if strings.Contains(request.Query, id) || strings.Contains(request.Query, reason) ||
					!strings.Contains(request.Query, root+"(") {
					t.Fatal("request must use variables and an existing API operation")
				}
				if write {
					var input SuppressionInput
					if err := json.Unmarshal(request.Variables["input"], &input); err != nil {
						t.Fatal(err)
					}
					if !input.Suppress || input.VulnerabilityID != id || input.Reason != reason || input.State == nil || *input.State != state {
						t.Fatalf("incorrect suppression input: %+v", input)
					}
				} else {
					if !strings.Contains(request.Query, "imageName") || !strings.Contains(request.Query, "reason") {
						t.Fatal("finding lookup omitted target identity or suppression reason")
					}
					var input string
					if err := json.Unmarshal(request.Variables["id"], &input); err != nil || input != id {
						t.Fatalf("incorrect finding ID: %q error=%v", input, err)
					}
				}
			})
			if write {
				out, err := applySuppression(t.Context(), c, SuppressionInput{
					VulnerabilityID: id, Suppress: true, Reason: reason, State: &state,
				})
				if err != nil || out.Id != id || out.Suppression == nil || out.Suppression.Reason != reason {
					t.Fatalf("out=%+v error=%v", out, err)
				}
			} else {
				out, err := fetchSuppressionFinding(t.Context(), c, id)
				if err != nil || out.ID != id || out.Identifier != "CVE-2026-1234" ||
					out.ImageName != "registry.example/app" || out.Package != "pkg:golang/example@v1.2.3" {
					t.Fatalf("out=%+v error=%v", out, err)
				}
			}
			if calls != 1 {
				t.Fatalf("unexpected calls/retries: %d", calls)
			}
		})
	}
}

func TestSuppressionPrivateReasonSnapshot(t *testing.T) {
	var previous *SuppressionState
	for _, reason := range []any{"", "original private reason", "edited private reason", nil, true} {
		c := suppressionHTTPClient(t, map[string]any{"data": map[string]any{"node": map[string]any{
			"__typename": "ImageVulnerability", "id": "VUL_finding",
			"suppression": map[string]any{"state": "IN_TRIAGE", "reason": reason},
		}}}, nil)
		out, err := fetchSuppressionFinding(t.Context(), c, "VUL_finding")
		if _, ok := reason.(string); !ok {
			if err == nil {
				t.Fatal("accepted a missing or invalid suppression reason")
			}
			continue
		}
		if err != nil || out.Suppression == nil {
			t.Fatalf("finding=%+v error=%v", out, err)
		}
		if previous != nil && previous.ReasonFingerprint == out.Suppression.ReasonFingerprint {
			t.Fatal("reason-only edit did not change the snapshot")
		}
		raw, err := json.Marshal(out)
		if err != nil || strings.Contains(string(raw), "reason") || strings.Contains(string(raw), "Fingerprint") {
			t.Fatalf("private reason snapshot leaked: %s error=%v", raw, err)
		}
		previous = out.Suppression
	}
}

func TestSuppressionRejectsMissingOrWrongNode(t *testing.T) {
	for _, node := range []any{nil, map[string]any{"__typename": "Application", "id": "APP_app"}} {
		c := suppressionHTTPClient(t, map[string]any{"data": map[string]any{"node": node}}, nil)
		if _, err := fetchSuppressionFinding(t.Context(), c, "VUL_finding"); err == nil {
			t.Fatalf("accepted invalid finding: %+v", node)
		}
	}
	c := suppressionHTTPClient(t, map[string]any{"data": map[string]any{"updateImageVulnerability": map[string]any{"vulnerability": nil}}}, nil)
	if _, err := applySuppression(t.Context(), c, SuppressionInput{}); err == nil {
		t.Fatal("missing mutation result reported as saved")
	}
}

func TestSuppressionErrorsDoNotLeakOrRetry(t *testing.T) {
	calls := 0
	c := suppressionHTTPClient(t, map[string]any{"errors": []any{map[string]any{"message": "PRIVATE JUSTIFICATION"}}}, func(*http.Request) { calls++ })
	if _, err := fetchSuppressionFinding(t.Context(), c, "VUL_finding"); err == nil || strings.Contains(err.Error(), "PRIVATE") {
		t.Fatalf("unsafe preview error: %v", err)
	}
	if _, err := applySuppression(t.Context(), c, SuppressionInput{}); err == nil || strings.Contains(err.Error(), "PRIVATE") {
		t.Fatalf("unsafe write error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("automatic retries: %d", calls)
	}
}
