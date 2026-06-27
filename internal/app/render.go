package app

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = "1"

type RawBody []byte

func Render(w io.Writer, format string, value any) error {
	switch format {
	case "", "yaml":
		data, err := yaml.Marshal(value)
		if err != nil {
			return WrapError("Could not render YAML output", "Try -o json to inspect the same data.", err)
		}
		_, err = w.Write(data)
		return err
	case "json":
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		return enc.Encode(value)
	case "raw":
		raw, ok := value.(RawBody)
		if !ok {
			return NewError("Raw output is not available for this command", "Use -o yaml or -o json for structured output.")
		}
		_, err := w.Write(raw)
		if len(raw) == 0 || raw[len(raw)-1] != '\n' {
			_, _ = fmt.Fprintln(w)
		}
		return err
	default:
		return NewError("Unsupported output format: "+format, "Use -o yaml, -o json, or -o raw.")
	}
}
