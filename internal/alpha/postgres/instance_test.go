package postgres_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/nais/cli/internal/naisapi/gql"
)

type deleteClient struct {
	t        *testing.T
	accepted bool
}

func (c deleteClient) MakeRequest(_ context.Context, req *graphql.Request, resp *graphql.Response) error {
	c.t.Helper()
	if req.OpName != "DeletePostgres" || !strings.Contains(req.Query, "deletionRequested") || strings.Contains(req.Query, "deletePostgresBranch") {
		c.t.Errorf("unexpected operation: %q / %q", req.OpName, req.Query)
	}
	data, err := json.Marshal(req.Variables)
	if err != nil {
		return err
	}
	var variables struct {
		Input gql.DeletePostgresInput `json:"input"`
	}
	if err := json.Unmarshal(data, &variables); err != nil {
		return err
	}
	if got, want := variables.Input, (gql.DeletePostgresInput{Name: "db", TeamSlug: "team", EnvironmentName: "dev"}); got != want {
		c.t.Errorf("input = %+v, want %+v", got, want)
	}
	if c.accepted {
		return json.Unmarshal([]byte(`{"deletePostgres":{"deletionRequested":true}}`), resp.Data)
	}
	return json.Unmarshal([]byte(`{"deletePostgres":{"deletionRequested":false}}`), resp.Data)
}

type changeClient struct {
	t         *testing.T
	operation string
	wantInput string
}

func (c changeClient) MakeRequest(_ context.Context, req *graphql.Request, resp *graphql.Response) error {
	c.t.Helper()
	if req.OpName != c.operation {
		c.t.Errorf("operation = %q, want %q", req.OpName, c.operation)
	}
	variables, err := json.Marshal(req.Variables)
	if err != nil {
		return err
	}
	if !strings.Contains(string(variables), c.wantInput) {
		c.t.Errorf("variables = %s, want %s", variables, c.wantInput)
	}
	if c.operation == "UpdatePostgres" {
		for _, field := range []string{`"cpu"`, `"memory"`, `"diskSize"`} {
			if strings.Contains(string(variables), field) {
				c.t.Errorf("unchanged field %s was sent: %s", field, variables)
			}
		}
		return json.Unmarshal([]byte(`{"updatePostgres":{"postgres":{"name":"db"}}}`), resp.Data)
	}
	return json.Unmarshal([]byte(`{"createPostgres":{"postgres":{"name":"db"}}}`), resp.Data)
}

func TestPostgresChangeMutationContract(t *testing.T) {
	created, err := gql.CreatePostgres(context.Background(), changeClient{t, "CreatePostgres", `"majorVersion":"18"`}, gql.CreatePostgresInput{
		Name: "db", TeamSlug: "team", EnvironmentName: "dev", MajorVersion: "18",
	})
	if err != nil || created.CreatePostgres.Postgres.Name != "db" {
		t.Errorf("create = %+v, %v", created, err)
	}
	value := false
	updated, err := gql.UpdatePostgres(context.Background(), changeClient{t, "UpdatePostgres", `"highAvailability":false`}, gql.UpdatePostgresInput{
		Name: "db", TeamSlug: "team", EnvironmentName: "dev", HighAvailability: &value,
	})
	if err != nil || updated.UpdatePostgres.Postgres.Name != "db" {
		t.Errorf("update = %+v, %v", updated, err)
	}
}

func TestDeletePostgresMutationContract(t *testing.T) {
	for _, accepted := range []bool{true, false} {
		result, err := gql.DeletePostgres(context.Background(), deleteClient{t: t, accepted: accepted}, gql.DeletePostgresInput{
			Name: "db", TeamSlug: "team", EnvironmentName: "dev",
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.DeletePostgres.DeletionRequested != accepted {
			t.Errorf("deletionRequested = %v, want %v", result.DeletePostgres.DeletionRequested, accepted)
		}
	}
}
