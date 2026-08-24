package helpers

import "encoding/json"

func DeepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	data, _ := json.Marshal(src)
	var dst map[string]interface{}
	_ = json.Unmarshal(data, &dst)
	return dst
}
