package tools

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nais/cli/internal/mcp/client"
	"github.com/nais/cli/internal/naisapi/gql"
)

type suppressionTestClient struct {
	client.Client
	finding       client.SuppressionFinding
	writes        int
	input         client.SuppressionInput
	writeError    error
	lookupError   error
	indeterminate bool
}

type denySuppressionRequests struct{}

func (denySuppressionRequests) Allow() bool { return false }

func (c *suppressionTestClient) GetSuppressionFinding(context.Context, string) (client.SuppressionFinding, error) {
	return c.finding, c.lookupError
}

func (c *suppressionTestClient) ApplySuppression(_ context.Context, input client.SuppressionInput) (client.SuppressionResult, error) {
	c.writes++
	c.input = input
	result := client.SuppressionResult{
		Id: input.VulnerabilityID,
		Suppression: &gql.MCPApplySuppressionUpdateImageVulnerabilityUpdateImageVulnerabilityPayloadVulnerabilityImageVulnerabilitySuppression{
			State: *input.State, Reason: input.Reason,
		},
	}
	if c.indeterminate {
		result.Suppression.Reason = "different"
	}
	return result, c.writeError
}

func suppressionHandlerFixture() (*suppressionHandler, *suppressionTestClient, ConfirmSuppressionInput) {
	c := &suppressionTestClient{finding: client.SuppressionFinding{
		TypeName: "ImageVulnerability", ID: "VUL_finding", Identifier: "CVE-2026-1234",
		ImageName: "registry.example/app", Package: "pkg:golang/example@v1.2.3?type=module",
	}}
	h := &suppressionHandler{toolContext: &toolContext{client: c, rateLimiter: allowAll{}}}
	return h, c, ConfirmSuppressionInput{VulnerabilityID: c.finding.ID}
}

func acceptedSuppression() *mcp.ElicitationResult {
	return &mcp.ElicitationResult{ElicitationResponse: mcp.ElicitationResponse{
		Action: mcp.ElicitationResponseActionAccept,
		Content: map[string]any{
			"confirm": true, "reason": "  Reviewed parser usage.\nCustom \"reason\": https://example.test/review/1\n\u00c5rsak: package disabled.  ", "state": "NOT_AFFECTED",
		},
	}}
}

func mutationRequest(t *testing.T, input ConfirmSuppressionInput) mcp.CallToolRequest {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var proposal map[string]any
	if err := json.Unmarshal(raw, &proposal); err != nil {
		t.Fatal(err)
	}
	return mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: "execute_mutation",
		Arguments: map[string]any{
			"mutation": "updateImageVulnerability",
			"input":    proposal,
		},
	}}
}

func TestSuppressionRequiresHumanConfirmation(t *testing.T) {
	for _, mode := range []string{
		"custom reason", "state IN_TRIAGE", "state FALSE_POSITIVE", "state RESOLVED",
		"declined", "cancelled", "unsupported", "missing consent", "missing reason", "blank reason",
		"long reason", "invalid state", "malformed", "changed finding", "changed image",
		"changed package version", "changed suppression", "changed reason",
		"lookup error after confirmation", "write error", "indeterminate",
	} {
		t.Run(mode, func(t *testing.T) {
			h, c, input := suppressionHandlerFixture()
			if mode == "changed reason" {
				c.finding.Suppression = &client.SuppressionState{
					State:             gql.ImageVulnerabilitySuppressionStateInTriage,
					ReasonFingerprint: sha256.Sum256([]byte("previous private reason")),
				}
			}
			result := acceptedSuppression()
			content := result.Content.(map[string]any)
			h.elicit = func(_ context.Context, request mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				for _, expected := range []string{input.VulnerabilityID, c.finding.Identifier, c.finding.ImageName, c.finding.Package, "Other teams", "future matches", "authenticated Nais identity", "reason in activity history"} {
					if !strings.Contains(request.Params.Message, expected) {
						t.Errorf("human warning omitted %q", expected)
					}
				}
				if strings.Contains(request.Params.Message, "previous private reason") || strings.Contains(request.Params.Message, "ReasonFingerprint") {
					t.Fatal("previous reason or fingerprint leaked into confirmation")
				}
				if after, ok := strings.CutPrefix(mode, "state "); ok {
					content["state"] = after
				}
				schema := request.Params.RequestedSchema.(map[string]any)
				properties := schema["properties"].(map[string]any)
				if properties["reason"].(map[string]any)["type"] != "string" ||
					!slices.Contains(schema["required"].([]string), "reason") ||
					!slices.Contains(properties["state"].(map[string]any)["enum"].([]string), content["state"].(string)) {
					t.Fatal("confirmation must offer a custom reason and the selected state")
				}
				switch mode {
				case "declined":
					result.Action = mcp.ElicitationResponseActionDecline
				case "cancelled":
					result.Action = mcp.ElicitationResponseActionCancel
				case "unsupported":
					return nil, errors.New("client does not support elicitation")
				case "missing consent":
					delete(content, "confirm")
				case "missing reason":
					delete(content, "reason")
				case "blank reason":
					content["reason"] = " "
				case "long reason":
					content["reason"] = strings.Repeat("x", 4097)
				case "invalid state":
					content["state"] = "INVALID"
				case "malformed":
					result.Content = "yes"
				case "changed finding":
					c.finding.Identifier = "changed"
				case "changed image":
					c.finding.ImageName = "registry.example/other"
				case "changed package version":
					c.finding.Package = "pkg:golang/example@v2.0.0?type=module"
				case "changed suppression":
					c.finding.Suppression = &client.SuppressionState{State: gql.ImageVulnerabilitySuppressionStateInTriage}
				case "changed reason":
					c.finding.Suppression.ReasonFingerprint = sha256.Sum256([]byte("edited private reason"))
				case "lookup error after confirmation":
					c.lookupError = errors.New("private upstream error")
				case "write error":
					c.writeError = errors.New("sensitive upstream message")
				case "indeterminate":
					c.indeterminate = true
				}
				return result, nil
			}
			out, err := h.confirm(t.Context(), input, nil)
			persisted := mode == "custom reason" || strings.HasPrefix(mode, "state ")
			write := persisted || mode == "write error" || mode == "indeterminate"
			if (c.writes == 1) != write || c.writes > 1 {
				t.Fatalf("unexpected writes: %d", c.writes)
			}
			if persisted {
				if err != nil || out.Status != "persisted" || !c.input.Suppress ||
					c.input.VulnerabilityID != input.VulnerabilityID ||
					c.input.Reason != content["reason"] ||
					c.input.State == nil || string(*c.input.State) != content["state"] {
					t.Fatalf("out=%+v error=%v input=%+v", out, err, c.input)
				}
			} else if out.Status == "persisted" {
				t.Fatal("unconfirmed action reported as saved")
			}
			if mode == "write error" || mode == "indeterminate" {
				if err != nil || out.Status != "indeterminate" || strings.Contains(out.Message, "sensitive upstream") {
					t.Fatalf("out=%+v error=%v", out, err)
				}
			}
			raw, _ := json.Marshal(out)
			if strings.Contains(string(raw), "Reviewed parser usage") || strings.Contains(string(raw), "review/1") ||
				strings.Contains(string(raw), "private reason") || strings.Contains(string(raw), "ReasonFingerprint") {
				t.Fatal("decision text leaked into tool output")
			}
		})
	}
}

func TestSuppressionRejectsInvalidFinding(t *testing.T) {
	for _, mode := range []string{"wrong id", "missing advisory", "api error", "empty id", "rate limit", "cancelled context"} {
		t.Run(mode, func(t *testing.T) {
			h, c, input := suppressionHandlerFixture()
			h.elicit = func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				t.Fatal("unsafe proposal reached human confirmation")
				return nil, nil
			}
			ctx := t.Context()
			switch mode {
			case "wrong id":
				c.finding.ID = "different"
			case "missing advisory":
				c.finding.Identifier = ""
			case "api error":
				c.lookupError = errors.New("private upstream error")
			case "empty id":
				input.VulnerabilityID = ""
			case "rate limit":
				h.rateLimiter = denySuppressionRequests{}
			case "cancelled context":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			out, err := h.confirm(ctx, input, nil)
			if c.writes != 0 || out.Status == "persisted" || err == nil {
				t.Fatalf("out=%+v error=%v writes=%d", out, err, c.writes)
			}
		})
	}
}

func TestSuppressionRequiresResolvedTarget(t *testing.T) {
	for _, target := range []struct {
		name, image, purl string
	}{
		{"missing image", "", "pkg:golang/example@v1.2.3"},
		{"blank image", " ", "pkg:golang/example@v1.2.3"},
		{"missing package", "registry.example/app", ""},
		{"package name only", "registry.example/app", "example"},
		{"unversioned purl", "registry.example/app", "pkg:golang/example"},
		{"missing package name", "registry.example/app", "pkg:golang/@v1.2.3"},
		{"missing version", "registry.example/app", "pkg:golang/example@"},
		{"blank version", "registry.example/app", "pkg:golang/example@%20"},
		{"invalid encoding", "registry.example/app", "pkg:golang/example@%invalid"},
		{"wrong scheme", "registry.example/app", "https:golang/example@v1.2.3"},
	} {
		t.Run(target.name, func(t *testing.T) {
			h, c, input := suppressionHandlerFixture()
			c.finding.ImageName, c.finding.Package = target.image, target.purl
			h.elicit = func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				t.Fatal("unresolved target reached confirmation")
				return nil, nil
			}
			out, err := h.confirm(t.Context(), input, nil)
			if err != nil || c.writes != 0 || out.Status != "proposed" ||
				!strings.Contains(out.Message, "no write attempted") {
				t.Fatalf("out=%+v error=%v writes=%d", out, err, c.writes)
			}
		})
	}
}

type suppressionElicitor struct{ calls int }

func (e *suppressionElicitor) Elicit(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
	e.calls++
	return acceptedSuppression(), nil
}

func TestSuppressionProtocolConfirmation(t *testing.T) {
	for _, tc := range []struct {
		name, protocol string
		human          bool
	}{
		{"legacy unsupported", "2025-11-25", false},
		{"legacy human", "2025-11-25", true},
		{"modern unsupported", mcp.LATEST_PROTOCOL_VERSION, false},
		{"modern human", mcp.LATEST_PROTOCOL_VERSION, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			human := tc.human
			h, backend, input := suppressionHandlerFixture()
			s := server.NewMCPServer("test", "1", server.WithToolCapabilities(false))
			registerMutationTool(s, h.toolContext)
			elicitor := &suppressionElicitor{}
			var options []mcpclient.ClientOption
			if human {
				options = append(options, mcpclient.WithElicitationHandler(elicitor))
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			serverRead, clientWrite := io.Pipe()
			clientRead, serverWrite := io.Pipe()
			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = server.NewStdioServer(s).Listen(ctx, serverRead, serverWrite)
			}()
			t.Cleanup(func() {
				cancel()
				_ = clientWrite.Close()
				_ = serverWrite.Close()
				_ = serverRead.Close()
				_ = clientRead.Close()
				<-done
			})
			c := mcpclient.NewClient(transport.NewIO(clientRead, clientWrite, nil), options...)
			if err := c.Start(ctx); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = c.Close() }()
			_, err := c.Initialize(ctx, mcp.InitializeRequest{Params: mcp.InitializeParams{
				ProtocolVersion: tc.protocol, ClientInfo: mcp.Implementation{Name: "test", Version: "1"},
			}})
			if err != nil {
				t.Fatal(err)
			}
			result, err := c.CallTool(ctx, mutationRequest(t, input))
			if err != nil || result.IsError {
				t.Fatalf("result=%+v error=%v", result, err)
			}
			raw, _ := json.Marshal(result.StructuredContent)
			var out SuppressionOutput
			if err := json.Unmarshal(raw, &out); err != nil {
				t.Fatal(err)
			}
			if human {
				if out.Status != "persisted" || backend.writes != 1 || elicitor.calls != 1 {
					t.Fatalf("status=%s message=%s writes=%d human prompts=%d needsInput=%v", out.Status, out.Message, backend.writes, elicitor.calls, result.NeedsInput())
				}
			} else if out.Status != "proposed" || backend.writes != 0 {
				t.Fatalf("model-supplied confirmation authorized a write: %+v", out)
			}
		})
	}
}

func TestSuppressionModernContinuation(t *testing.T) {
	for _, mode := range []string{"success and replay", "expired", "changed input", "forged token", "missing response", "changed finding", "changed image", "changed package version", "changed reason", "lost capability", "argument consent", "concurrent"} {
		t.Run(mode, func(t *testing.T) {
			h, backend, input := suppressionHandlerFixture()
			if mode == "changed reason" {
				backend.finding.Suppression = &client.SuppressionState{
					State:             gql.ImageVulnerabilitySuppressionStateInTriage,
					ReasonFingerprint: sha256.Sum256([]byte("previous private reason")),
				}
			}
			ctx := server.WithRequestProtocolInfo(t.Context(), &server.RequestProtocolInfo{
				Modern: true, ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientCapabilities: &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapability{}},
			})
			request := mutationRequest(t, input)
			first, err := h.handleExecuteMutation(ctx, request)
			if err != nil || !first.NeedsInput() || backend.writes != 0 {
				t.Fatalf("first=%+v error=%v writes=%d", first, err, backend.writes)
			}
			raw, err := json.Marshal(first)
			if err != nil || strings.Contains(string(raw), "private reason") || strings.Contains(string(raw), "ReasonFingerprint") {
				t.Fatal("previous reason or fingerprint leaked into continuation")
			}
			token := first.RequestState
			request.Params.MultiRoundTripParams = mcp.MultiRoundTripParams{
				RequestState: token, InputResponses: mcp.InputResponses{"suppression": mcp.NewElicitationInputResponse(*acceptedSuppression())},
			}
			switch mode {
			case "expired":
				pending := h.pending[token]
				pending.Expires = time.Now().Add(-time.Minute)
				h.pending[token] = pending
			case "changed input":
				proposal := request.Params.Arguments.(map[string]any)["input"].(map[string]any)
				proposal["vulnerabilityID"] = "other"
			case "forged token":
				request.Params.RequestState = "forged"
			case "missing response":
				request.Params.InputResponses = nil
			case "changed finding":
				backend.finding.Identifier = "changed"
			case "changed image":
				backend.finding.ImageName = "registry.example/other"
			case "changed package version":
				backend.finding.Package = "pkg:golang/example@v2.0.0?type=module"
			case "changed reason":
				backend.finding.Suppression.ReasonFingerprint = sha256.Sum256([]byte("edited private reason"))
			case "lost capability":
				ctx = server.WithRequestProtocolInfo(t.Context(), &server.RequestProtocolInfo{Modern: true, ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION})
			case "argument consent":
				request.Params.MultiRoundTripParams = mcp.MultiRoundTripParams{}
				args := request.Params.Arguments.(map[string]any)
				args["requestState"], args["inputResponses"] = token, acceptedSuppression()
			}
			if mode == "concurrent" {
				done := make(chan *mcp.CallToolResult, 1)
				go func() { result, _ := h.handleExecuteMutation(ctx, request); done <- result }()
				result, err := h.handleExecuteMutation(ctx, request)
				other := <-done
				if err != nil || result == nil || other == nil || result.IsError == other.IsError || backend.writes != 1 {
					t.Fatalf("concurrent confirmation was not single-use: writes=%d error=%v", backend.writes, err)
				}
				return
			}
			result, err := h.handleExecuteMutation(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			wantWrites := 0
			if mode == "success and replay" {
				wantWrites = 1
				replay, err := h.handleExecuteMutation(ctx, request)
				if err != nil || !replay.IsError {
					t.Fatalf("continuation reused: %+v error=%v", replay, err)
				}
			}
			if backend.writes != wantWrites {
				t.Fatalf("unsafe continuation writes=%d want=%d result=%+v", backend.writes, wantWrites, result)
			}
		})
	}
}

func TestMutationToolDiscovery(t *testing.T) {
	s := server.NewMCPServer("test", "1", server.WithToolCapabilities(false))
	h, _, _ := suppressionHandlerFixture()
	RegisterTools(s, h.client, allowAll{}, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	raw, _ := json.Marshal(s.HandleMessage(t.Context(), request))
	var response struct {
		Result struct {
			Tools []struct {
				Name        string         `json:"name"`
				InputSchema map[string]any `json:"inputSchema"`
				Annotations map[string]any `json:"annotations"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	found := 0
	readOnly := false
	for _, tool := range response.Result.Tools {
		if tool.Name == "execute_graphql" {
			readOnly = tool.Annotations["readOnlyHint"] == true
		}
		if tool.Name != "execute_mutation" {
			continue
		}
		found++
		props := tool.InputSchema["properties"].(map[string]any)
		if len(props) != 2 || props["mutation"] == nil || props["input"] == nil || len(tool.InputSchema["required"].([]any)) != 2 {
			t.Fatalf("unexpected mutation schema: %v", tool.InputSchema)
		}
		if tool.Annotations["readOnlyHint"] != false || tool.Annotations["destructiveHint"] != true || tool.Annotations["idempotentHint"] != false {
			t.Fatalf("unsafe write annotations: %v", tool.Annotations)
		}
	}
	if found != 1 || !readOnly {
		t.Fatalf("found %d mutation tools; read-only query tool: %v", found, readOnly)
	}
}

func TestExecuteMutationRejectsUnsafeInput(t *testing.T) {
	for _, mode := range []string{
		"unknown mutation", "nonexistent mutation", "graphql document", "multiple mutations",
		"missing mutation", "missing input", "invalid input type", "confirmed", "reason", "state",
		"suppress", "actor", "unknown field", "missing id", "invalid id",
	} {
		t.Run(mode, func(t *testing.T) {
			h, backend, input := suppressionHandlerFixture()
			h.elicit = func(context.Context, mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
				t.Fatal("invalid mutation reached human confirmation")
				return nil, nil
			}
			request := mutationRequest(t, input)
			args := request.Params.Arguments.(map[string]any)
			proposal := args["input"].(map[string]any)
			switch mode {
			case "unknown mutation":
				args["mutation"] = "restartApplication"
			case "nonexistent mutation":
				args["mutation"] = "applyVulnerabilitySuppression"
			case "graphql document":
				args["mutation"] = "mutation { updateImageVulnerability { vulnerability { id } } }"
			case "multiple mutations":
				args["mutation"] = []string{"updateImageVulnerability", "restartApplication"}
			case "missing mutation":
				delete(args, "mutation")
			case "missing input":
				delete(args, "input")
			case "invalid input type":
				args["input"] = "confirmed"
			case "confirmed":
				proposal["confirmed"] = true
			case "reason":
				proposal["reason"] = "model text"
			case "state":
				proposal["state"] = "NOT_AFFECTED"
			case "suppress":
				proposal["suppress"] = true
			case "actor":
				proposal["actor"] = "someone else"
			case "unknown field":
				proposal["extra"] = true
			case "missing id":
				delete(proposal, "vulnerabilityID")
			case "invalid id":
				proposal["vulnerabilityID"] = ""
			}
			result, err := h.handleExecuteMutation(t.Context(), request)
			if err != nil || result == nil || !result.IsError || backend.writes != 0 {
				t.Fatalf("result=%+v error=%v writes=%d", result, err, backend.writes)
			}
		})
	}
}
