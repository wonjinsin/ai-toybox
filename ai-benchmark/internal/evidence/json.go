package evidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

const maxJSONNesting = 1000

func parseLine(line int, raw []byte) (map[string]any, map[string]any) {
	failure := func(kind, message string) (map[string]any, map[string]any) {
		return nil, map[string]any{"line": line, "kind": kind, "message": message}
	}
	if !utf8.Valid(raw) {
		return failure("invalid_utf8", "Event contains invalid UTF-8")
	}
	if err := validateScalarEscapes(raw); err != nil {
		return failure("invalid_json", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := readValue(decoder, 0)
	if err != nil {
		return failure("invalid_json", err.Error())
	}
	if _, err := decoder.Token(); err != io.EOF {
		return failure("invalid_json", "Expected exactly one JSON value per line")
	}
	event, isObject := value.(map[string]any)
	_, hasType := event["type"].(string)
	if !isObject || !hasType {
		return failure("invalid_event", "Event must be an object with a string type")
	}
	return event, nil
}

func readValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maxJSONNesting {
		return nil, fmt.Errorf("JSON nesting exceeds %d levels", maxJSONNesting)
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch token {
	case json.Delim('{'):
		return readObject(decoder, depth)
	case json.Delim('['):
		return readArray(decoder, depth)
	default:
		return token, nil
	}
}

func readObject(decoder *json.Decoder, depth int) (map[string]any, error) {
	object := map[string]any{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, valid := token.(string)
		if !valid {
			return nil, fmt.Errorf("JSON object keys must be strings")
		}
		if _, exists := object[key]; exists {
			return nil, fmt.Errorf("JSON event contains duplicate key %q", key)
		}
		value, err := readValue(decoder, depth+1)
		if err != nil {
			return nil, err
		}
		object[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return object, nil
}

func readArray(decoder *json.Decoder, depth int) ([]any, error) {
	array := []any{}
	for decoder.More() {
		value, err := readValue(decoder, depth+1)
		if err != nil {
			return nil, err
		}
		array = append(array, value)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return array, nil
}
