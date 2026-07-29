package common

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)
func setHeaderOverrideInContext(context map[string]interface{}, headerName string, value interface{}, keepOrigin bool) error {
	headerName = normalizeHeaderContextKey(headerName)
	if headerName == "" {
		return fmt.Errorf("header name is required")
	}

	rawHeaders := ensureMapKeyInContext(context, paramOverrideContextHeaderOverride)
	if keepOrigin {
		if existing, ok := rawHeaders[headerName]; ok {
			existingValue := strings.TrimSpace(fmt.Sprintf("%v", existing))
			if existingValue != "" {
				return nil
			}
		}
	}

	headerValue, hasValue, err := resolveHeaderOverrideValue(context, headerName, value)
	if err != nil {
		return err
	}
	if !hasValue {
		delete(rawHeaders, headerName)
		return nil
	}

	rawHeaders[headerName] = headerValue
	return nil
}

func resolveHeaderOverrideValue(context map[string]interface{}, headerName string, value interface{}) (string, bool, error) {
	if value == nil {
		return "", false, fmt.Errorf("header value is required")
	}

	if mapping, ok := value.(map[string]interface{}); ok {
		return resolveHeaderOverrideValueByMapping(context, headerName, mapping)
	}
	if mapping, ok := value.(map[string]string); ok {
		converted := make(map[string]interface{}, len(mapping))
		for key, item := range mapping {
			converted[key] = item
		}
		return resolveHeaderOverrideValueByMapping(context, headerName, converted)
	}

	headerValue := strings.TrimSpace(fmt.Sprintf("%v", value))
	if headerValue == "" {
		return "", false, nil
	}
	return headerValue, true, nil
}

func resolveHeaderOverrideValueByMapping(context map[string]interface{}, headerName string, mapping map[string]interface{}) (string, bool, error) {
	if len(mapping) == 0 {
		return "", false, fmt.Errorf("header value mapping cannot be empty")
	}

	appendTokens, err := parseHeaderAppendTokens(mapping)
	if err != nil {
		return "", false, err
	}
	keepOnlyDeclared := parseHeaderKeepOnlyDeclared(mapping)

	sourceValue, exists := getHeaderValueFromContext(context, headerName)
	sourceTokens := make([]string, 0)
	if exists {
		sourceTokens = splitHeaderListValue(sourceValue)
	}

	wildcardValue, hasWildcard := mapping["*"]
	resultTokens := make([]string, 0, len(sourceTokens)+len(appendTokens))
	for _, token := range sourceTokens {
		replacementRaw, hasReplacement := mapping[token]
		if !hasReplacement && hasWildcard && !keepOnlyDeclared {
			replacementRaw = wildcardValue
			hasReplacement = true
		}
		if !hasReplacement {
			if keepOnlyDeclared {
				continue
			}
			resultTokens = append(resultTokens, token)
			continue
		}
		replacementTokens, err := parseHeaderReplacementTokens(replacementRaw)
		if err != nil {
			return "", false, err
		}
		resultTokens = append(resultTokens, replacementTokens...)
	}

	resultTokens = append(resultTokens, appendTokens...)
	resultTokens = lo.Uniq(resultTokens)
	if len(resultTokens) == 0 {
		return "", false, nil
	}
	return strings.Join(resultTokens, ","), true, nil
}

func parseHeaderAppendTokens(mapping map[string]interface{}) ([]string, error) {
	appendRaw, ok := mapping["$append"]
	if !ok {
		return nil, nil
	}
	return parseHeaderReplacementTokens(appendRaw)
}

func parseHeaderKeepOnlyDeclared(mapping map[string]interface{}) bool {
	keepOnlyDeclaredRaw, ok := mapping["$keep_only_declared"]
	if !ok {
		return false
	}
	keepOnlyDeclared, ok := keepOnlyDeclaredRaw.(bool)
	if !ok {
		return false
	}
	return keepOnlyDeclared
}

func parseHeaderReplacementTokens(value interface{}) ([]string, error) {
	switch raw := value.(type) {
	case nil:
		return nil, nil
	case string:
		return splitHeaderListValue(raw), nil
	case []string:
		tokens := make([]string, 0, len(raw))
		for _, item := range raw {
			tokens = append(tokens, splitHeaderListValue(item)...)
		}
		return lo.Uniq(tokens), nil
	case []interface{}:
		tokens := make([]string, 0, len(raw))
		for _, item := range raw {
			itemTokens, err := parseHeaderReplacementTokens(item)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, itemTokens...)
		}
		return lo.Uniq(tokens), nil
	case map[string]interface{}, map[string]string:
		return nil, fmt.Errorf("header replacement value must be string, array or null")
	default:
		token := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if token == "" {
			return nil, nil
		}
		return []string{token}, nil
	}
}

func splitHeaderListValue(raw string) []string {
	items := strings.Split(raw, ",")
	return lo.FilterMap(items, func(item string, _ int) (string, bool) {
		token := strings.TrimSpace(item)
		if token == "" {
			return "", false
		}
		return token, true
	})
}

func copyHeaderInContext(context map[string]interface{}, fromHeader, toHeader string, keepOrigin bool) error {
	fromHeader = normalizeHeaderContextKey(fromHeader)
	toHeader = normalizeHeaderContextKey(toHeader)
	if fromHeader == "" || toHeader == "" {
		return fmt.Errorf("copy_header from/to is required")
	}
	value, exists := getHeaderValueFromContext(context, fromHeader)
	if !exists {
		return fmt.Errorf("%w: %s", errSourceHeaderNotFound, fromHeader)
	}
	return setHeaderOverrideInContext(context, toHeader, value, keepOrigin)
}

func moveHeaderInContext(context map[string]interface{}, fromHeader, toHeader string, keepOrigin bool) error {
	fromHeader = normalizeHeaderContextKey(fromHeader)
	toHeader = normalizeHeaderContextKey(toHeader)
	if fromHeader == "" || toHeader == "" {
		return fmt.Errorf("move_header from/to is required")
	}
	if err := copyHeaderInContext(context, fromHeader, toHeader, keepOrigin); err != nil {
		return err
	}
	if strings.EqualFold(fromHeader, toHeader) {
		return nil
	}
	return deleteHeaderOverrideInContext(context, fromHeader)
}

func deleteHeaderOverrideInContext(context map[string]interface{}, headerName string) error {
	headerName = normalizeHeaderContextKey(headerName)
	if headerName == "" {
		return fmt.Errorf("header name is required")
	}
	rawHeaders := ensureMapKeyInContext(context, paramOverrideContextHeaderOverride)
	delete(rawHeaders, headerName)
	return nil
}
