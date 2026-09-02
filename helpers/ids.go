package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func InterfaceToInt(value interface{}) (int, bool) {
	switch typedValue := value.(type) {
	case int:
		return typedValue, true
	case int64:
		return int(typedValue), true
	case float64:
		return int(typedValue), true
	case float32:
		return int(typedValue), true
	case json.Number:
		parsedValue, err := strconv.Atoi(typedValue.String())
		return parsedValue, err == nil
	case string:
		parsedValue, err := strconv.Atoi(typedValue)
		return parsedValue, err == nil
	default:
		return 0, false
	}
}

func IDToString(value interface{}) (string, bool) {
	id, ok := InterfaceToInt(value)
	if !ok {
		return "", false
	}
	return strconv.Itoa(id), true
}

func IDFromRelation(value interface{}) (int, bool) {
	if value == nil {
		return 0, false
	}
	if id, ok := InterfaceToInt(value); ok {
		return id, true
	}
	relation, ok := value.(map[string]interface{})
	if !ok || relation == nil {
		return 0, false
	}
	return InterfaceToInt(relation["Id"])
}

func RelationIDToString(value interface{}) (string, error) {
	id, ok := IDFromRelation(value)
	if !ok || id <= 0 {
		return "", fmt.Errorf("relación inválida")
	}
	return strconv.Itoa(id), nil
}

func ExtractID(value interface{}) (int, bool) {
	if id, ok := IDFromRelation(value); ok {
		return id, true
	}
	if item, ok := value.(map[string]interface{}); ok {
		return IDFromRelation(item["Id"])
	}
	return 0, false
}

func DocumentoIDOrNil(value interface{}) interface{} {
	id, ok := InterfaceToInt(value)
	if !ok {
		return nil
	}
	return id
}
