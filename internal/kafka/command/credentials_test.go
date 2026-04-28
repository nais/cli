package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func newTestOutputWriter(buf *bytes.Buffer) *naistrix.OutputWriter {
	level := naistrix.OutputVerbosityLevelNormal
	return naistrix.NewOutputWriter(buf, &level)
}

func testKafkaCreds() *gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials {
	return &gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials{
		Username:       "myteam_alice_deadbeef_99",
		AccessCert:     "-----BEGIN CERTIFICATE-----\ntest-cert\n-----END CERTIFICATE-----",
		AccessKey:      "-----BEGIN PRIVATE KEY-----\ntest-key\n-----END PRIVATE KEY-----",
		CaCert:         "-----BEGIN CERTIFICATE-----\ntest-ca\n-----END CERTIFICATE-----",
		Brokers:        "broker1:9092,broker2:9092",
		SchemaRegistry: "https://schema-registry:8081",
	}
}

func lineValue(output, prefix string) string {
	for line := range strings.SplitSeq(output, "\n") {
		if after, ok := strings.CutPrefix(line, prefix); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

func TestWriteCertFiles(t *testing.T) {
	creds := testKafkaCreds()

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
		{"keystore", files.keystore, strings.TrimSpace(creds.AccessCert) + "\n" + strings.TrimSpace(creds.AccessKey) + "\n"},
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

func TestWriteKafkaEnv(t *testing.T) {
	var buf bytes.Buffer
	out := newTestOutputWriter(&buf)
	creds := testKafkaCreds()

	if err := writeKafkaEnv(out, creds); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	for _, want := range []string{
		`KAFKA_BROKERS="broker1:9092,broker2:9092"`,
		`KAFKA_USERNAME="myteam_alice_deadbeef_99"`,
		`KAFKA_SCHEMA_REGISTRY="https://schema-registry:8081"`,
		`KAFKA_SCHEMA_REGISTRY_USER="myteam_alice_deadbeef_99"`,
		"KAFKA_CERTIFICATE=$(cat <<'NAIS_KAFKA_CERT_EOF'",
		"KAFKA_PRIVATE_KEY=$(cat <<'NAIS_KAFKA_KEY_EOF'",
		"KAFKA_CA=$(cat <<'NAIS_KAFKA_CA_EOF'",
		creds.AccessCert,
		creds.AccessKey,
		creds.CaCert,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("env output missing %q\n%s", want, got)
		}
	}
}

func TestWriteKafkaKcat(t *testing.T) {
	var buf bytes.Buffer
	out := newTestOutputWriter(&buf)
	creds := testKafkaCreds()

	if err := writeKafkaKcat(out, creds); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	dir := lineValue(got, "Kafka kcat configuration written to: ")
	if dir == "" {
		t.Fatalf("missing output dir in:\n%s", got)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	if !strings.Contains(got, "Warning: ") {
		t.Fatalf("missing warning in output:\n%s", got)
	}

	configFile := filepath.Join(dir, "kcat.conf")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("reading kcat config: %v", err)
	}

	text := string(content)
	for _, want := range []string{
		"# nais-cli ",
		"bootstrap.servers=broker1:9092,broker2:9092",
		`# sasl.username=myteam_alice_deadbeef_99 (use "alice" with 'nais kafka grant-access')`,
		"security.protocol=ssl",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("kcat config missing %q\n%s", want, text)
		}
	}
}

func TestWriteKafkaJava(t *testing.T) {
	var buf bytes.Buffer
	out := newTestOutputWriter(&buf)
	creds := testKafkaCreds()

	if err := writeKafkaJava(out, creds); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	dir := lineValue(got, "Kafka Java configuration written to: ")
	if dir == "" {
		t.Fatalf("missing output dir in:\n%s", got)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	if !strings.Contains(got, "Warning: ") {
		t.Fatalf("missing warning in output:\n%s", got)
	}

	configFile := filepath.Join(dir, "kafka.properties")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("reading Java config: %v", err)
	}

	text := string(content)
	for _, want := range []string{
		"# nais-cli ",
		`# sasl.username=myteam_alice_deadbeef_99 (use "alice" with 'nais kafka grant-access')`,
		"security.protocol=SSL",
		"ssl.truststore.type=PEM",
		"ssl.keystore.type=PEM",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("Java config missing %q\n%s", want, text)
		}
	}
}

func TestKafkaApplicationName(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"tsm_tmp-kafka-topic-4cea36_50d5a466_pF5", "tmp-kafka-topic-4cea36"},
		{"redundant-team_application_18515795_99", "application"},
		{"my-app", "my-app"},
		{"", ""},
	} {
		if got := kafkaApplicationName(tc.in); got != tc.want {
			t.Errorf("kafkaApplicationName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
