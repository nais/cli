package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func credentials(parentFlags *flag.Kafka) *naistrix.Command {
	flags := &flag.Credentials{Kafka: parentFlags}
	return &naistrix.Command{
		Name:        "credentials",
		Title:       "Create temporary credentials for Kafka.",
		Description: "Creates temporary credentials for accessing Kafka. Output format can be env (default), kcat, or java. The env format prints environment variables to stdout. The kcat and java formats write configuration files to a temporary directory.",
		Flags:       flags,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if len(flags.Environment) != 1 {
				return fmt.Errorf("exactly one environment is required, set using --environment/-e flag")
			}
			if flags.TTL == "" {
				return fmt.Errorf("ttl is required, set using --ttl flag (e.g. '1d', '7d')")
			}
			output := string(flags.Output)
			if output != "" && output != "env" && output != "kcat" && output != "java" {
				return fmt.Errorf("invalid output format %q, must be one of: env, kcat, java", output)
			}
			return nil
		},
		Examples: []naistrix.Example{
			{
				Description: "Create Kafka credentials in environment dev, valid for 1 day, output as environment variables.",
				Command:     "--environment dev --ttl 1d",
			},
			{
				Description: "Create Kafka credentials and output kcat configuration files.",
				Command:     "--environment dev --ttl 1d --output kcat",
			},
			{
				Description: "Create Kafka credentials and output Java configuration files.",
				Command:     "--environment dev --ttl 1d --output java",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			creds, err := kafka.CreateCredentials(
				ctx,
				flags.Team,
				flags.Environment[0],
				flags.TTL,
			)
			if err != nil {
				return fmt.Errorf("creating Kafka credentials: %w", err)
			}

			output := string(flags.Output)
			if output == "" {
				output = "env"
			}

			switch output {
			case "env":
				return writeKafkaEnv(out, creds)
			case "kcat":
				return writeKafkaKcat(out, creds)
			case "java":
				return writeKafkaJava(out, creds)
			default:
				return fmt.Errorf("unknown output format: %s", output)
			}
		},
	}
}

// kafkaCertFiles holds the paths to the TLS certificate files written to disk.
type kafkaCertFiles struct {
	cert, key, ca, keystore string
}

// writeCertFiles writes the Kafka TLS credentials to a temporary directory
// and returns the file paths.
func writeCertFiles(creds *gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials) (kafkaCertFiles, error) {
	dir, err := os.MkdirTemp("", "nais-kafka-*")
	if err != nil {
		return kafkaCertFiles{}, fmt.Errorf("creating temp directory: %w", err)
	}

	cleanupDir := func(cause error) (kafkaCertFiles, error) {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			return kafkaCertFiles{}, fmt.Errorf("%w (failed cleaning temp directory: %v)", cause, removeErr)
		}
		return kafkaCertFiles{}, cause
	}

	files := kafkaCertFiles{
		cert:     filepath.Join(dir, "kafka-certificate.crt"),
		key:      filepath.Join(dir, "kafka-private-key.pem"),
		ca:       filepath.Join(dir, "kafka-ca.pem"),
		keystore: filepath.Join(dir, "kafka-keystore.pem"),
	}

	for _, f := range []struct {
		path    string
		content string
	}{
		{files.cert, creds.AccessCert},
		{files.key, creds.AccessKey},
		{files.ca, creds.CaCert},
		{files.keystore, strings.TrimSpace(creds.AccessCert) + "\n" + strings.TrimSpace(creds.AccessKey) + "\n"},
	} {
		if err := os.WriteFile(f.path, []byte(f.content), 0o600); err != nil {
			return cleanupDir(fmt.Errorf("writing %s: %w", filepath.Base(f.path), err))
		}
	}

	return files, nil
}

func writeKafkaEnv(out *naistrix.OutputWriter, creds *gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials) error {
	out.Println(fmt.Sprintf("KAFKA_BROKERS=%q", creds.Brokers))
	out.Println(fmt.Sprintf("KAFKA_SCHEMA_REGISTRY=%q", creds.SchemaRegistry))
	out.Println(fmt.Sprintf("KAFKA_CERTIFICATE=%q", creds.AccessCert))
	out.Println(fmt.Sprintf("KAFKA_PRIVATE_KEY=%q", creds.AccessKey))
	out.Println(fmt.Sprintf("KAFKA_CA=%q", creds.CaCert))
	return nil
}

func writeKafkaKcat(out *naistrix.OutputWriter, creds *gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials) error {
	files, err := writeCertFiles(creds)
	if err != nil {
		return err
	}

	configFile := filepath.Join(filepath.Dir(files.cert), "kcat.conf")

	var config strings.Builder
	config.WriteString(fmt.Sprintf("# nais %s\n# kcat -F %s -t your.topic\n",
		time.Now().Truncate(time.Minute), configFile))
	config.WriteString(fmt.Sprintf("bootstrap.servers=%s\n", creds.Brokers))
	config.WriteString("security.protocol=ssl\n")
	config.WriteString(fmt.Sprintf("ssl.certificate.location=%s\n", files.cert))
	config.WriteString(fmt.Sprintf("ssl.key.location=%s\n", files.key))
	config.WriteString(fmt.Sprintf("ssl.ca.location=%s\n", files.ca))

	if err := os.WriteFile(configFile, []byte(config.String()), 0o600); err != nil {
		return fmt.Errorf("writing kcat config: %w", err)
	}

	out.Println(fmt.Sprintf("Kafka kcat configuration written to: %s", filepath.Dir(files.cert)))
	out.Println(fmt.Sprintf("Usage: kcat -F %s -t your.topic", configFile))
	return nil
}

func writeKafkaJava(out *naistrix.OutputWriter, creds *gql.CreateKafkaCredentialsCreateKafkaCredentialsCreateKafkaCredentialsPayloadCredentialsKafkaCredentials) error {
	files, err := writeCertFiles(creds)
	if err != nil {
		return err
	}

	configFile := filepath.Join(filepath.Dir(files.cert), "kafka.properties")

	var properties strings.Builder
	properties.WriteString(fmt.Sprintf("# nais-cli %s\n", time.Now().Truncate(time.Minute)))
	properties.WriteString(fmt.Sprintf("# Usage: kafka-console-consumer.sh --topic your.topic --bootstrap-server %s --consumer.config %s\n",
		creds.Brokers, configFile))
	properties.WriteString("security.protocol=SSL\n")
	properties.WriteString("ssl.protocol=TLS\n")
	properties.WriteString(fmt.Sprintf("ssl.truststore.location=%s\n", windowsify(files.ca)))
	properties.WriteString("ssl.truststore.type=PEM\n")
	properties.WriteString(fmt.Sprintf("ssl.keystore.location=%s\n", windowsify(files.keystore)))
	properties.WriteString("ssl.keystore.type=PEM\n")

	if err := os.WriteFile(configFile, []byte(properties.String()), 0o600); err != nil {
		return fmt.Errorf("writing Java config: %w", err)
	}

	out.Println(fmt.Sprintf("Kafka Java configuration written to: %s", filepath.Dir(files.cert)))
	out.Println(fmt.Sprintf("Usage: kafka-console-consumer.sh --topic your.topic --bootstrap-server %s --consumer.config %s", creds.Brokers, configFile))
	return nil
}

func windowsify(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "/", "\\")
	}
	return path
}
