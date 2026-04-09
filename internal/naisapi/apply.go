package naisapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ApplyManifests(ctx context.Context, teamSlug, environmentName string, manifests []unstructured.Unstructured) error {
	const url = "%v/api/v1/teams/%v/environments/%v/apply"

	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated user: %w", err)
	}

	bodyData := struct {
		Resources []unstructured.Unstructured `json:"resources"`
	}{
		Resources: manifests,
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(bodyData); err != nil {
		return fmt.Errorf("failed to encode manifests to YAML: %w", err)
	}

	uri := fmt.Sprintf(url, strings.TrimSuffix(user.APIURL(), "/graphql"), teamSlug, environmentName)
	fmt.Println(uri)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := user.HTTPClient(ctx).Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	for k, v := range req.Header {
		fmt.Printf("%s: %s\n", k, v)
	}

	io.Copy(os.Stdout, resp.Body)

	return nil
}
