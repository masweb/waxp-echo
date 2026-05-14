package i18n

import (
	"encoding/json"
	"fmt"
)

var localeMapKeys = map[string]bool{
	"menu": true,
}

func Resolve(data json.RawMessage, locale string) (json.RawMessage, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("i18n resolve: %w", err)
	}
	out, err := json.Marshal(resolveValue(v, locale))
	if err != nil {
		return nil, fmt.Errorf("i18n resolve: %w", err)
	}
	return out, nil
}

func Wrap(data json.RawMessage, locale string) (json.RawMessage, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("i18n wrap: %w", err)
	}
	out, err := json.Marshal(wrapValue(v, locale))
	if err != nil {
		return nil, fmt.Errorf("i18n wrap: %w", err)
	}
	return out, nil
}

func Merge(incoming, existing json.RawMessage, locale string) (json.RawMessage, error) {
	var inc interface{}
	if err := json.Unmarshal(incoming, &inc); err != nil {
		return nil, fmt.Errorf("i18n merge: %w", err)
	}
	var exist interface{}
	if err := json.Unmarshal(existing, &exist); err != nil {
		return nil, fmt.Errorf("i18n merge: %w", err)
	}
	out, err := json.Marshal(mergeSync(inc, exist, locale))
	if err != nil {
		return nil, fmt.Errorf("i18n merge: %w", err)
	}
	return out, nil
}

// --- Resolve ---

func resolveValue(v interface{}, locale string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		if loc, ok := val["locales"].(map[string]interface{}); ok {
			resolved := make(map[string]interface{}, len(loc))
			for k, entry := range loc {
				resolved[k] = pickString(entry, locale)
			}
			val["locales"] = resolved
		}
		for key := range localeMapKeys {
			if m, ok := val[key].(map[string]interface{}); ok {
				if data, ok := m[locale]; ok {
					val[key] = data
				} else {
					val[key] = []interface{}{}
				}
			}
		}
		for k, child := range val {
			val[k] = resolveValue(child, locale)
		}
		return val
	case []interface{}:
		for i, elem := range val {
			val[i] = resolveValue(elem, locale)
		}
		return val
	default:
		return v
	}
}

func pickString(entry interface{}, locale string) interface{} {
	m, ok := entry.(map[string]interface{})
	if !ok {
		return entry
	}
	if s, ok := m[locale].(string); ok {
		return s
	}
	return ""
}

// --- Wrap ---

func wrapValue(v interface{}, locale string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		if loc, ok := val["locales"].(map[string]interface{}); ok {
			val["locales"] = wrapLocalesMap(loc, locale)
		}
		for key := range localeMapKeys {
			if data, exists := val[key]; exists && data != nil {
				if _, isMap := data.(map[string]interface{}); !isMap {
					val[key] = map[string]interface{}{locale: data}
				}
			}
		}
		for k, child := range val {
			val[k] = wrapValue(child, locale)
		}
		return val
	case []interface{}:
		for i, elem := range val {
			val[i] = wrapValue(elem, locale)
		}
		return val
	default:
		return v
	}
}

// --- Merge ---

func mergeSync(incoming, existing interface{}, locale string) interface{} {
	switch inc := incoming.(type) {
	case map[string]interface{}:
		if exist, ok := existing.(map[string]interface{}); ok {
			return mergeSyncMap(inc, exist, locale)
		}
		return wrapValue(incoming, locale)
	case []interface{}:
		if existArr, ok := existing.([]interface{}); ok {
			return mergeSyncArray(inc, existArr, locale)
		}
		return wrapValue(incoming, locale)
	default:
		return incoming
	}
}

func mergeSyncMap(inc, exist map[string]interface{}, locale string) map[string]interface{} {
	if loc, ok := inc["locales"].(map[string]interface{}); ok {
		if existLoc, ok := exist["locales"].(map[string]interface{}); ok {
			inc["locales"] = mergeLocalesMap(loc, existLoc, locale)
		} else {
			inc["locales"] = wrapLocalesMap(loc, locale)
		}
	}
	for key := range localeMapKeys {
		incData, hasInc := inc[key]
		if !hasInc {
			continue
		}
		if existMap, ok := exist[key].(map[string]interface{}); ok {
			existMap[locale] = incData
			inc[key] = existMap
		} else {
			inc[key] = map[string]interface{}{locale: incData}
		}
	}
	for k, v := range inc {
		if k == "locales" || localeMapKeys[k] {
			continue
		}
		if existChild, ok := exist[k]; ok {
			inc[k] = mergeSync(v, existChild, locale)
		} else {
			inc[k] = wrapValue(v, locale)
		}
	}
	for k, v := range exist {
		if _, ok := inc[k]; !ok && k != "locales" && !localeMapKeys[k] {
			inc[k] = v
		}
	}
	return inc
}

func mergeSyncArray(inc []interface{}, exist []interface{}, locale string) []interface{} {
	existByID := make(map[interface{}]interface{}, len(exist))
	for _, e := range exist {
		if m, ok := e.(map[string]interface{}); ok {
			if id, ok := m["id"]; ok {
				existByID[id] = e
			}
		}
	}
	for i, elem := range inc {
		var existElem interface{}
		if m, ok := elem.(map[string]interface{}); ok {
			if id, ok := m["id"]; ok {
				if e, found := existByID[id]; found {
					existElem = e
				}
			}
		}
		if existElem == nil && i < len(exist) {
			existElem = exist[i]
		}
		if existElem != nil {
			inc[i] = mergeSync(elem, existElem, locale)
		} else {
			inc[i] = wrapValue(elem, locale)
		}
	}
	return inc
}

// --- Shared helpers ---

func mergeLocalesMap(incoming, existing map[string]interface{}, locale string) map[string]interface{} {
	merged := make(map[string]interface{}, len(incoming))
	for k, v := range incoming {
		if s, ok := v.(string); ok {
			if existMap, ok := existing[k].(map[string]interface{}); ok {
				existMap[locale] = s
				merged[k] = existMap
			} else {
				merged[k] = map[string]interface{}{locale: s}
			}
		} else {
			merged[k] = v
		}
	}
	return merged
}

func wrapLocalesMap(loc map[string]interface{}, locale string) map[string]interface{} {
	wrapped := make(map[string]interface{}, len(loc))
	for k, v := range loc {
		if s, ok := v.(string); ok {
			wrapped[k] = map[string]interface{}{locale: s}
		} else {
			wrapped[k] = v
		}
	}
	return wrapped
}
