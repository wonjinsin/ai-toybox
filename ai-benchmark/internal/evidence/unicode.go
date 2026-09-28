package evidence

import (
	"fmt"
	"strconv"
)

// encoding/json replaces unpaired surrogates; validate before losing evidence.
func validateScalarEscapes(raw []byte) error {
	inString := false
	for index := 0; index < len(raw); index++ {
		if raw[index] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[index] != '\\' || index+1 >= len(raw) {
			continue
		}
		if raw[index+1] != 'u' {
			index++
			continue
		}
		code, valid := unicodeEscape(raw, index)
		if !valid {
			return fmt.Errorf("JSON string contains an invalid Unicode escape")
		}
		switch {
		case code >= 0xd800 && code <= 0xdbff:
			low, valid := unicodeEscape(raw, index+6)
			if !valid || low < 0xdc00 || low > 0xdfff {
				return fmt.Errorf("JSON string contains an unpaired Unicode surrogate")
			}
			index += 11
		case code >= 0xdc00 && code <= 0xdfff:
			return fmt.Errorf("JSON string contains an unpaired Unicode surrogate")
		default:
			index += 5
		}
	}
	return nil
}

func unicodeEscape(raw []byte, offset int) (uint64, bool) {
	if offset+6 > len(raw) || raw[offset] != '\\' || raw[offset+1] != 'u' {
		return 0, false
	}
	value, err := strconv.ParseUint(string(raw[offset+2:offset+6]), 16, 16)
	return value, err == nil
}
