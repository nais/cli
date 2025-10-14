package naisapi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Khan/genqlient/graphql"
	"github.com/sirupsen/logrus"
)

func DoSSEQuery[T any](u *url.URL, client *http.Client, graphqlRequest graphql.Request, c chan T, log logrus.FieldLogger) error {
	body := bytes.Buffer{}
	err := json.NewEncoder(&body).Encode(graphqlRequest)
	if err != nil {
		log.WithError(err).Error("Error encoding GraphQL request")
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), &body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("Error connecting to SSE endpoint")
	}

	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	var event, data []byte
	for scanner.Scan() {
		line := scanner.Bytes()
		// empty line indicates end of event
		if len(line) == 0 {
			// Event ended, print if any data was received
			if len(data) > 0 {
				fmt.Printf("Event: %s\nData: %s\n\n", event, data)

				var decoded T
				err := json.Unmarshal(data, &decoded)
				if err != nil {
					return err
				}

				c <- decoded
				event, data = nil, nil
			}
			continue
		}
		if after, ok := bytes.CutPrefix(line, []byte("event:")); ok {
			event = bytes.TrimSpace(after)
		}
		if after, ok := bytes.CutPrefix(line, []byte("data:")); ok {
			data = append(data, bytes.TrimSpace(after)...)
			data = append(data, '\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
