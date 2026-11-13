package webmcp

// Schema is a small fluent builder for the JSON Schema of a tool's input.
// It covers the shapes WebMCP tools actually use: objects with typed
// properties and a required list.

// ObjectSchema starts a schema for a tool input object.
func ObjectSchema(props ...Property) map[string]any {
	required := []string{}
	p := map[string]any{}
	for _, prop := range props {
		p[prop.Name] = prop.schema
		if prop.Required {
			required = append(required, prop.Name)
		}
	}
	out := map[string]any{"type": "object", "properties": p}
	if len(required) > 0 {
		out["required"] = required
	}
	return out
}

// Property is one named, typed input field.
type Property struct {
	Name     string
	Required bool
	schema   map[string]any
}

// String adds a string property.
func String(name, description string) Property {
	return Property{Name: name, schema: map[string]any{"type": "string", "description": description}}
}

// Number adds a number property.
func Number(name, description string) Property {
	return Property{Name: name, schema: map[string]any{"type": "number", "description": description}}
}

// Integer adds an integer property.
func Integer(name, description string) Property {
	return Property{Name: name, schema: map[string]any{"type": "integer", "description": description}}
}

// Boolean adds a boolean property.
func Boolean(name, description string) Property {
	return Property{Name: name, schema: map[string]any{"type": "boolean", "description": description}}
}

// Enum adds a string property restricted to the given values.
func Enum(name, description string, values ...string) Property {
	return Property{Name: name, schema: map[string]any{"type": "string", "description": description, "enum": values}}
}

// Req marks the property required.
func (p Property) Req() Property {
	p.Required = true
	return p
}
