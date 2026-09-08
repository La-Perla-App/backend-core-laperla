package jsonparser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type jsonNode map[string]any

func ConvertToJSON(data any) any {
	// Si el dato es un tipo primitivo, lo retornamos tal cual
	switch reflect.TypeOf(data).Kind() {
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
		return data
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var result any
	json.Unmarshal(jsonData, &result)

	return result
}

func getJSONPathValue(path string, node jsonNode) any {
	var currentNode = node
	var value any
	keys := strings.Split(path, ".")
	for _, key := range keys {
		tempNode := getNodeValue(key, -1, currentNode)
		if tempNode == nil {
			value = currentNode[key]
			continue
		}
		currentNode = tempNode
		value = currentNode
	}
	return value
}

func getNodeValue(key string, arrayIndex int, node jsonNode) jsonNode {
	tempNode := getNode(node[key], arrayIndex)
	if tempNode == nil {
		tempKey, index, err := formatKey(key)
		if err != nil {
			return nil
		}
		return getNodeValue(tempKey, index, node)
	}
	return tempNode
}

func getNode(v any, i int) jsonNode {
	switch node := v.(type) {
	case map[string]any:
		configNode := jsonNode(node)
		return configNode
	case []any:
		if len(node) <= i {
			return nil
		}
		if i == -1 {
			return nil
		}
		configNode := jsonNode(getNode(node[i], 0))
		return configNode
	default:
		return nil
	}
}

func formatKey(key string) (string, int, error) {
	i := strings.Index(key, "[")
	k := strings.Index(key, "]")
	if i == -1 {
		return "", 0, fmt.Errorf("Not format key for index")
	}
	formatKey := key[:i]
	if formatKey == "" {
		return "", 0, fmt.Errorf("Not format key")
	}
	index, err := strconv.Atoi(key[i+1 : k])
	if err != nil {
		return "", 0, err
	}
	return formatKey, index, nil
}
