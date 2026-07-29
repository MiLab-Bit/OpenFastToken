package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MiLab-Bit/OpenFastToken/common"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)
func tryParseOperations(paramOverride map[string]interface{}) ([]ParamOperation, bool) {
	// 检查是否包含 "operations" 字段
	opsValue, exists := paramOverride["operations"]
	if !exists {
		return nil, false
	}

	var opMaps []map[string]interface{}
	switch ops := opsValue.(type) {
	case []interface{}:
		opMaps = make([]map[string]interface{}, 0, len(ops))
		for _, op := range ops {
			opMap, ok := op.(map[string]interface{})
			if !ok {
				return nil, false
			}
			opMaps = append(opMaps, opMap)
		}
	case []map[string]interface{}:
		opMaps = ops
	default:
		return nil, false
	}

	operations := make([]ParamOperation, 0, len(opMaps))
	for _, opMap := range opMaps {
		operation := ParamOperation{}

		// 断言必要字段
		if path, ok := opMap["path"].(string); ok {
			operation.Path = path
		}
		if mode, ok := opMap["mode"].(string); ok {
			operation.Mode = mode
		} else {
			return nil, false // mode 是必需的
		}

		// 可选字段
		if value, exists := opMap["value"]; exists {
			operation.Value = value
		}
		if keepOrigin, ok := opMap["keep_origin"].(bool); ok {
			operation.KeepOrigin = keepOrigin
		}
		if from, ok := opMap["from"].(string); ok {
			operation.From = from
		}
		if to, ok := opMap["to"].(string); ok {
			operation.To = to
		}
		if logic, ok := opMap["logic"].(string); ok {
			operation.Logic = logic
		} else {
			operation.Logic = "OR" // 默认为OR
		}

		// 解析条件
		if conditions, exists := opMap["conditions"]; exists {
			parsedConditions, err := parseConditionOperations(conditions)
			if err != nil {
				return nil, false
			}
			operation.Conditions = append(operation.Conditions, parsedConditions...)
		}

		operations = append(operations, operation)
	}
	return operations, true
}

func checkConditions(data []byte, contextJSON string, conditions []ConditionOperation, logic string) (bool, error) {
	if len(conditions) == 0 {
		return true, nil // 没有条件，直接通过
	}
	results := make([]bool, len(conditions))
	for i, condition := range conditions {
		result, err := checkSingleCondition(data, contextJSON, condition)
		if err != nil {
			return false, err
		}
		results[i] = result
	}

	if strings.ToUpper(logic) == "AND" {
		return lo.EveryBy(results, func(item bool) bool { return item }), nil
	}
	return lo.SomeBy(results, func(item bool) bool { return item }), nil
}

func checkSingleCondition(data []byte, contextJSON string, condition ConditionOperation) (bool, error) {
	// 处理负数索引
	path := processNegativeIndex(data, condition.Path)
	value := gjson.GetBytes(data, path)
	if !value.Exists() && contextJSON != "" {
		value = gjson.Get(contextJSON, condition.Path)
	}
	if !value.Exists() {
		if condition.PassMissingKey {
			return true, nil
		}
		return false, nil
	}

	// 利用gjson的类型解析
	targetBytes, err := common.Marshal(condition.Value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal condition value: %v", err)
	}
	targetValue := gjson.ParseBytes(targetBytes)

	result, err := compareGjsonValues(value, targetValue, strings.ToLower(condition.Mode))
	if err != nil {
		return false, fmt.Errorf("comparison failed for path %s: %v", condition.Path, err)
	}

	if condition.Invert {
		result = !result
	}
	return result, nil
}

func processNegativeIndex(data []byte, path string) string {
	matches := negativeIndexRegexp.FindAllStringSubmatch(path, -1)

	if len(matches) == 0 {
		return path
	}

	result := path
	for _, match := range matches {
		negIndex := match[1]
		index, _ := strconv.Atoi(negIndex)

		arrayPath := strings.Split(path, negIndex)[0]
		if strings.HasSuffix(arrayPath, ".") {
			arrayPath = arrayPath[:len(arrayPath)-1]
		}

		array := gjson.GetBytes(data, arrayPath)
		if array.IsArray() {
			length := len(array.Array())
			actualIndex := length + index
			if actualIndex >= 0 && actualIndex < length {
				result = strings.Replace(result, match[0], "."+strconv.Itoa(actualIndex), 1)
			}
		}
	}

	return result
}

func compareGjsonValues(jsonValue, targetValue gjson.Result, mode string) (bool, error) {
	switch mode {
	case "full":
		return compareEqual(jsonValue, targetValue)
	case "prefix":
		return strings.HasPrefix(jsonValue.String(), targetValue.String()), nil
	case "suffix":
		return strings.HasSuffix(jsonValue.String(), targetValue.String()), nil
	case "contains":
		return strings.Contains(jsonValue.String(), targetValue.String()), nil
	case "gt":
		return compareNumeric(jsonValue, targetValue, "gt")
	case "gte":
		return compareNumeric(jsonValue, targetValue, "gte")
	case "lt":
		return compareNumeric(jsonValue, targetValue, "lt")
	case "lte":
		return compareNumeric(jsonValue, targetValue, "lte")
	default:
		return false, fmt.Errorf("unsupported comparison mode: %s", mode)
	}
}

func compareEqual(jsonValue, targetValue gjson.Result) (bool, error) {
	// 对null值特殊处理：两个都是null返回true，一个是null另一个不是返回false
	if jsonValue.Type == gjson.Null || targetValue.Type == gjson.Null {
		return jsonValue.Type == gjson.Null && targetValue.Type == gjson.Null, nil
	}

	// 对布尔值特殊处理
	if (jsonValue.Type == gjson.True || jsonValue.Type == gjson.False) &&
		(targetValue.Type == gjson.True || targetValue.Type == gjson.False) {
		return jsonValue.Bool() == targetValue.Bool(), nil
	}

	// 如果类型不同，报错
	if jsonValue.Type != targetValue.Type {
		return false, fmt.Errorf("compare for different types, got %v and %v", jsonValue.Type, targetValue.Type)
	}

	switch jsonValue.Type {
	case gjson.True, gjson.False:
		return jsonValue.Bool() == targetValue.Bool(), nil
	case gjson.Number:
		return jsonValue.Num == targetValue.Num, nil
	case gjson.String:
		return jsonValue.String() == targetValue.String(), nil
	default:
		return jsonValue.String() == targetValue.String(), nil
	}
}

func compareNumeric(jsonValue, targetValue gjson.Result, operator string) (bool, error) {
	// 只有数字类型才支持数值比较
	if jsonValue.Type != gjson.Number || targetValue.Type != gjson.Number {
		return false, fmt.Errorf("numeric comparison requires both values to be numbers, got %v and %v", jsonValue.Type, targetValue.Type)
	}

	jsonNum := jsonValue.Num
	targetNum := targetValue.Num

	switch operator {
	case "gt":
		return jsonNum > targetNum, nil
	case "gte":
		return jsonNum >= targetNum, nil
	case "lt":
		return jsonNum < targetNum, nil
	case "lte":
		return jsonNum <= targetNum, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}
