package naisapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ApplyResponse struct {
	Results []ResourceResult `json:"results"`
}

type ResourceResult struct {
	Resource        string `json:"resource"`
	EnvironmentName string `json:"environmentName"`
	Status          string `json:"status"`
	Error           string `json:"error,omitempty"`
}

func ApplyManifests(ctx context.Context, teamSlug, environmentName string, manifests []unstructured.Unstructured) (*ApplyResponse, error) {
	const url = "%v/api/v1/teams/%v/environments/%v/apply"

	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}

	bodyData := struct {
		Resources []unstructured.Unstructured `json:"resources"`
	}{
		Resources: manifests,
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(bodyData); err != nil {
		return nil, fmt.Errorf("failed to encode manifests: %w", err)
	}

	uri := fmt.Sprintf(url, strings.TrimSuffix(user.APIURL(), "/graphql"), teamSlug, environmentName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := user.HTTPClient(ctx).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("apply failed with HTTP status %d", resp.StatusCode)
	}

	var applyResp ApplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&applyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &applyResp, nil
}
