package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderYAMLJSONAndRaw(t *testing.T) {
	value := map[string]string{"schemaVersion": SchemaVersion, "key": "PROJ-1"}
	var yamlOut bytes.Buffer
	if err := Render(&yamlOut, "yaml", value); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yamlOut.String(), "schemaVersion: \"1\"") {
		t.Fatalf("missing schema version in yaml: %s", yamlOut.String())
	}
	var jsonOut bytes.Buffer
	if err := Render(&jsonOut, "json", value); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsonOut.String(), `"schemaVersion":"1"`) {
		t.Fatalf("missing schema version in json: %s", jsonOut.String())
	}
	var rawOut bytes.Buffer
	if err := Render(&rawOut, "raw", RawBody(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rawOut.String(), `{"ok":true}`) {
		t.Fatalf("unexpected raw output: %s", rawOut.String())
	}
}
