package naisapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func SSEQuery[T any](ctx context.Context, graphqlRequest graphql.Request, onData func(T), onError func(gqlerror.Error)) error {
	user, err := GetAuthenticatedUser(ctx)
	if err != nil {
		return err
	}

	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(graphqlRequest); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, user.APIURL(), &body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	resp, err := user.HTTPClient(ctx).Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	scanner := bufio.NewScanner(resp.Body)

	var data []byte
	for scanner.Scan() {
		line := scanner.Bytes()

		// empty line indicates end of event
		if len(line) == 0 {
			if len(data) == 0 {
				continue
			}

			var decoded graphql.BaseResponse[*T]
			if err := json.Unmarshal(data, &decoded); err != nil {
				return err
			}

			if decoded.Data != nil {
				onData(*decoded.Data)
			}

			for _, e := range decoded.Errors {
				onError(*e)
			}

			data = nil
		}

		if after, ok := bytes.CutPrefix(line, []byte("data:")); ok {
			data = append(data, bytes.TrimSpace(after)...)
			data = append(data, '\n')
		}
	}

	return scanner.Err()
}
