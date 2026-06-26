package apply

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nais/naistrix"
	"gopkg.in/yaml.v3"
)

// render resolves the manifest to apply by applying mixin and --set overrides on
// top of the base file.
//
// When neither a mixin nor any --set overrides are in play, the base file is
// returned unchanged (preserving multi-document files). Otherwise the base must
// contain exactly one YAML document, onto which the mixin is deep-merged and the
// --set overrides applied, in that order (base < mixin < set).
//
// If mixinPath is empty, an adjacent "<base>.<env>.yaml" file is auto-loaded when
// it exists.
func render(basePath, mixinPath, environment string, sets []string, out *naistrix.OutputWriter) ([]byte, error) {
	if basePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}
	if err := requireYAMLExtension(basePath); err != nil {
		return nil, err
	}

	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", basePath, err)
	}

	if mixinPath == "" {
		if auto := autoMixinPath(basePath, environment); auto != "" {
			if _, err := os.Stat(auto); err == nil {
				mixinPath = auto
				out.Printf("✓ auto-loaded mixin from %s\n", auto)
			}
		}
	}

	if mixinPath == "" && len(sets) == 0 {
		return baseData, nil
	}

	base, err := decodeSingleDocument(baseData, basePath)
	if err != nil {
		return nil, err
	}

	merged := base
	if mixinPath != "" {
		mixinData, err := os.ReadFile(mixinPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read mixin %s: %w", mixinPath, err)
		}
		mixin, err := decodeSingleDocument(mixinData, mixinPath)
		if err != nil {
			return nil, err
		}
		m, ok := deepMerge(base, mixin).(map[string]any)
		if !ok {
			return nil, fmt.Errorf("merge of %s and %s did not produce a mapping", basePath, mixinPath)
		}
		merged = m
	}

	for _, s := range sets {
		key, value, ok := splitSet(s)
		if !ok {
			return nil, fmt.Errorf("invalid --set %q: expected KEY=VALUE", s)
		}
		if err := applySet(merged, key, value); err != nil {
			return nil, err
		}
	}

	rendered, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to render merged manifest: %w", err)
	}
	return rendered, nil
}

// requireYAMLExtension returns an error unless path has a .yaml or .yml extension.
func requireYAMLExtension(path string) error {
	switch strings.TrimLeft(filepath.Ext(path), ".") {
	case "yaml", "yml":
		return nil
	default:
		return fmt.Errorf("unsupported file extension for file %s; expected .yaml or .yml", path)
	}
}

// autoMixinPath returns the conventional per-environment mixin path for a base
// file, e.g. ("nais.yaml", "dev") -> "nais.dev.yaml". It returns "" when there is
// no environment.
func autoMixinPath(basePath, environment string) string {
	if environment == "" {
		return ""
	}
	ext := filepath.Ext(basePath)
	stem := basePath[:len(basePath)-len(ext)]
	return stem + "." + environment + ext
}

// splitSet splits a "key=value" string on the first "=".
func splitSet(s string) (key, value string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}

// decodeSingleDocument decodes exactly one YAML document into a map. It is an
// error if the file is empty or contains more than one document.
func decodeSingleDocument(data []byte, path string) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s contains no YAML document", path)
		}
		return nil, fmt.Errorf("failed to decode YAML from %s: %w", path, err)
	}

	var extra map[string]any
	if err := decoder.Decode(&extra); err == nil {
		return nil, fmt.Errorf("%s contains multiple YAML documents; mixins and --set require a single document", path)
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to decode YAML from %s: %w", path, err)
	}

	return doc, nil
}

// renderDir collects all YAML base files in a directory (excluding mixin files),
// renders each with its environment-specific mixin (if present), and returns the
// concatenated YAML output.
//
// knownEnvs is an optional list of known Nais environment names used to identify
// mixin files: a file "<base>.<env>.yaml" is treated as a mixin when the
// corresponding base file exists and env is a known environment. When knownEnvs
// is empty or nil, any "<base>.<suffix>.yaml" file with a matching base is
// treated as a mixin (heuristic fallback).
func renderDir(dirPath, environment string, knownEnvs []string, out *naistrix.OutputWriter) ([]byte, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Collect all YAML files and identify which are mixins.
	var yamlFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		yamlFiles = append(yamlFiles, name)
	}

	mixins := mixinSet(yamlFiles, knownEnvs)

	// Collect base files (non-mixins) sorted alphabetically.
	var baseFiles []string
	for _, name := range yamlFiles {
		if mixins[name] {
			continue
		}
		baseFiles = append(baseFiles, name)
	}
	sort.Strings(baseFiles)

	if len(baseFiles) == 0 {
		return nil, fmt.Errorf("no YAML resource files found in %s", dirPath)
	}

	var combined []byte
	for _, name := range baseFiles {
		basePath := filepath.Join(dirPath, name)
		data, err := render(basePath, "", environment, nil, out)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if len(combined) > 0 {
			combined = append(combined, []byte("---\n")...)
		}
		combined = append(combined, data...)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			combined = append(combined, '\n')
		}
	}

	return combined, nil
}

// mixinSet returns the set of filenames that are environment-specific mixins.
// A file is considered a mixin when it matches the pattern "<base>.<env>.<ext>"
// and the corresponding base file "<base>.<ext>" exists in the directory.
//
// When knownEnvs is non-empty, only suffixes present in the list are considered
// environment names; otherwise any suffix is accepted (heuristic fallback for
// when the environment list could not be fetched).
func mixinSet(yamlFiles []string, knownEnvs []string) map[string]bool {
	fileSet := make(map[string]bool, len(yamlFiles))
	for _, f := range yamlFiles {
		fileSet[f] = true
	}

	envSet := make(map[string]bool, len(knownEnvs))
	for _, e := range knownEnvs {
		envSet[e] = true
	}

	mixins := make(map[string]bool)

	for _, name := range yamlFiles {
		ext := filepath.Ext(name)
		stem := name[:len(name)-len(ext)]

		before, after, ok := strings.Cut(stem, ".")
		if !ok {
			continue
		}

		baseName := before + ext
		if !fileSet[baseName] {
			continue
		}

		suffix := after
		if len(envSet) > 0 && !envSet[suffix] {
			continue
		}

		mixins[name] = true
	}

	return mixins
}
