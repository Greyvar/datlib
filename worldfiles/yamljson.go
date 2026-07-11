package worldfiles

import (
	"encoding/json"
	"fmt"
)

func marshalYAMLValue(v interface{}) ([]byte, error) {
	return json.Marshal(toJSONCompatible(v))
}

func toJSONCompatible(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(val))
		for key, value := range val {
			out[fmt.Sprint(key)] = toJSONCompatible(value)
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for key, value := range val {
			out[key] = toJSONCompatible(value)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, value := range val {
			out[i] = toJSONCompatible(value)
		}
		return out
	default:
		return v
	}
}
