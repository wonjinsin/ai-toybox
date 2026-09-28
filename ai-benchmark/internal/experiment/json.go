package experiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"
)

func strictJSON(raw []byte) (any, error) {
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("experiment must contain valid UTF-8 JSON")
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("experiment must contain valid JSON within the parser nesting limit")
	}
	if err := validateUnicodeEscapes(raw); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	return decodeValue(decoder, 0)
}

func validateUnicodeEscapes(raw []byte) error {
	inString := false
	for index := 0; index < len(raw); index++ {
		if raw[index] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[index] != '\\' {
			continue
		}
		if raw[index+1] != 'u' {
			index++
			continue
		}
		value, _ := strconv.ParseUint(string(raw[index+2:index+6]), 16, 16)
		if value >= 0xDC00 && value <= 0xDFFF {
			return fmt.Errorf("experiment JSON must encode valid UTF-8 Unicode scalars; unpaired low surrogate")
		}
		if value >= 0xD800 && value <= 0xDBFF {
			if index+12 > len(raw) || raw[index+6] != '\\' || raw[index+7] != 'u' {
				return fmt.Errorf("experiment JSON must encode valid UTF-8 Unicode scalars; unpaired high surrogate")
			}
			low, err := strconv.ParseUint(string(raw[index+8:index+12]), 16, 16)
			if err != nil || low < 0xDC00 || low > 0xDFFF {
				return fmt.Errorf("experiment JSON must encode valid UTF-8 Unicode scalars; invalid surrogate pair")
			}
			index += 11
		} else {
			index += 5
		}
	}
	return nil
}

func decodeValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > 1000 {
		return nil, fmt.Errorf("experiment JSON exceeds the 1000-level parser nesting limit")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid experiment JSON: %w", err)
	}
	switch token {
	case json.Delim('{'):
		object := map[string]any{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			name := key.(string)
			if _, exists := object[name]; exists {
				return nil, fmt.Errorf("JSON contains duplicate keys: %s", name)
			}
			value, err := decodeValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			object[name] = value
		}
		_, err := decoder.Token()
		return object, err
	case json.Delim('['):
		array := []any{}
		for decoder.More() {
			value, err := decodeValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		_, err := decoder.Token()
		return array, err
	default:
		return token, nil
	}
}
