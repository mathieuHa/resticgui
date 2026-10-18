package main

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The settings table of the configuration reference lists every key of the
// sample configuration, with its value there as the default, and the binary
// reads exactly those keys.
func TestConfigurationReference(t *testing.T) {
	doc, err := os.ReadFile("../../docs/reference/configuration.md")
	if err != nil {
		t.Fatal(err)
	}
	_, table, ok := strings.Cut(string(doc), "<!-- settings:")
	if !ok {
		t.Fatal("no settings marker in the configuration reference")
	}
	table, _, ok = strings.Cut(table, "<!-- end of settings -->")
	if !ok {
		t.Fatal("no end of settings marker in the configuration reference")
	}
	row := regexp.MustCompile("(?m)^\\| `([a-z_.]+)` \\| `([^`]*)` \\|")
	documented := map[string]string{}
	for _, m := range row.FindAllStringSubmatch(table, -1) {
		documented[m[1]] = m[2]
	}
	if len(documented) == 0 {
		t.Fatal("no settings in the table")
	}

	sample := flatten(t, "config.yaml")
	if !mapsKeysEqual(documented, sample) {
		t.Errorf("documented keys %v, config.yaml has %v", sorted(documented), sorted(sample))
	}
	for key, value := range sample {
		if documented[key] != value {
			t.Errorf("%s: documented default %q, config.yaml has %q", key, documented[key], value)
		}
	}
	for key := range flatten(t, "config-docker.yaml") {
		if _, ok := documented[key]; !ok {
			t.Errorf("config-docker.yaml has %s, which is not documented", key)
		}
	}

	read := map[string]string{}
	get := regexp.MustCompile(`viper\.(?:Get\w*|SetDefault)\("([a-z_.]+)"`)
	for _, file := range []string{"main.go", "../../internal/logs/logs.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range get.FindAllStringSubmatch(string(src), -1) {
			read[m[1]] = ""
		}
	}
	if !mapsKeysEqual(documented, read) {
		t.Errorf("documented keys %v, the binary reads %v", sorted(documented), sorted(read))
	}
}

// flatten returns the keys of a configuration file as dotted paths, with
// their values written as in the file.
func flatten(t *testing.T, name string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	keys := map[string]string{}
	var walk func(prefix string, n *yaml.Node)
	walk = func(prefix string, n *yaml.Node) {
		switch n.Kind {
		case yaml.DocumentNode:
			walk(prefix, n.Content[0])
		case yaml.MappingNode:
			for i := 0; i < len(n.Content); i += 2 {
				key := n.Content[i].Value
				if prefix != "" {
					key = prefix + "." + key
				}
				walk(key, n.Content[i+1])
			}
		case yaml.SequenceNode:
			items := make([]string, len(n.Content))
			for i, c := range n.Content {
				items[i] = c.Value
			}
			keys[prefix] = "[" + strings.Join(items, ", ") + "]"
		default:
			value := n.Value
			if n.Style == yaml.DoubleQuotedStyle && value == "" {
				value = `""`
			}
			keys[prefix] = value
		}
	}
	walk("", &root)
	return keys
}

func mapsKeysEqual(a, b map[string]string) bool {
	return slices.Equal(sorted(a), sorted(b))
}

func sorted(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
