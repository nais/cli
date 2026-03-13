package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
)

func TestWriteCertFiles(t *testing.T) {
	creds := &gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials{
		AccessCert:     "-----BEGIN CERTIFICATE-----\ntest-cert\n-----END CERTIFICATE-----",
		AccessKey:      "-----BEGIN PRIVATE KEY-----\ntest-key\n-----END PRIVATE KEY-----",
		CaCert:         "-----BEGIN CERTIFICATE-----\ntest-ca\n-----END CERTIFICATE-----",
		Brokers:        "broker1:9092,broker2:9092",
		SchemaRegistry: "https://schema-registry:8081",
	}

	files, err := writeCertFiles(creds)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(filepath.Dir(files.cert))
	})

	for _, tc := range []struct {
		name    string
		path    string
		content string
	}{
		{"cert", files.cert, creds.AccessCert},
		{"key", files.key, creds.AccessKey},
		{"ca", files.ca, creds.CaCert},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("reading %s: %v", tc.path, err)
			}
			if string(got) != tc.content {
				t.Errorf("got %q, want %q", got, tc.content)
			}

			info, err := os.Stat(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			if perm := info.Mode().Perm(); perm != 0o600 {
				t.Errorf("permissions = %o, want 0600", perm)
			}
		})
	}
}

func TestWindowsify(t *testing.T) {
	// On non-Windows, windowsify should be a no-op
	input := "/tmp/nais-kafka-123/kafka-certificate.crt"
	got := windowsify(input)
	if got != input {
		t.Errorf("windowsify(%q) = %q, want %q", input, got, input)
	}
}
