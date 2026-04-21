package handler

import "encoding/json"

func extractLocalesByID(v interface{}, m map[int64]map[string]interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		if id, ok := val["id"]; ok {
			if locales, ok := val["locales"]; ok {
				if idFloat, ok := id.(float64); ok {
					if localesMap, ok := locales.(map[string]interface{}); ok {
						m[int64(idFloat)] = localesMap
					}
				}
			}
		}
		for _, v := range val {
			extractLocalesByID(v, m)
		}
	case []interface{}:
		for _, v := range val {
			extractLocalesByID(v, m)
		}
	}
}

func resolveLocales(v interface{}, locale string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		if locales, ok := val["locales"]; ok {
			if localesMap, ok := locales.(map[string]interface{}); ok {
				resolved := make(map[string]interface{}, len(localesMap))
				for k, v := range localesMap {
					if m, ok := v.(map[string]interface{}); ok {
						if s, ok := m[locale]; ok {
							if str, ok := s.(string); ok {
								resolved[k] = str
							} else {
								resolved[k] = ""
							}
						} else {
							resolved[k] = ""
						}
					} else {
						resolved[k] = v
					}
				}
				val["locales"] = resolved
			}
		}
		for k, v := range val {
			val[k] = resolveLocales(v, locale)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = resolveLocales(v, locale)
		}
		return val
	default:
		return v
	}
}

func wrapLocales(v interface{}, locale string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		if locales, ok := val["locales"]; ok {
			if localesMap, ok := locales.(map[string]interface{}); ok {
				wrapped := make(map[string]interface{}, len(localesMap))
				for k, v := range localesMap {
					if s, ok := v.(string); ok {
						wrapped[k] = map[string]interface{}{locale: s}
					} else {
						wrapped[k] = v
					}
				}
				val["locales"] = wrapped
			}
		}
		for k, v := range val {
			val[k] = wrapLocales(v, locale)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = wrapLocales(v, locale)
		}
		return val
	default:
		return v
	}
}

func mergeLocales(incoming interface{}, existing map[int64]map[string]interface{}, locale string) interface{} {
	switch val := incoming.(type) {
	case map[string]interface{}:
		if locales, ok := val["locales"]; ok {
			if localesMap, ok := locales.(map[string]interface{}); ok {
				merged := make(map[string]interface{}, len(localesMap))
				for k, v := range localesMap {
					if s, ok := v.(string); ok {
						if id := val["id"]; id != nil {
							if idFloat, ok := id.(float64); ok {
								if existLocales, ok := existing[int64(idFloat)]; ok {
									if existMap, ok := existLocales[k].(map[string]interface{}); ok {
										existMap[locale] = s
										merged[k] = existMap
										continue
									}
								}
							}
						}
						merged[k] = map[string]interface{}{locale: s}
					} else {
						merged[k] = v
					}
				}
				val["locales"] = merged
			}
		}
		for k, v := range val {
			if k == "locales" {
				continue
			}
			val[k] = mergeLocales(v, existing, locale)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = mergeLocales(v, existing, locale)
		}
		return val
	default:
		return incoming
	}
}

func resolveLayoutLocales(layout json.RawMessage, locale string) (json.RawMessage, error) {
	var v interface{}
	if err := json.Unmarshal(layout, &v); err != nil {
		return nil, err
	}
	resolved := resolveLocales(v, locale)
	return json.Marshal(resolved)
}

func wrapLayoutLocales(layout json.RawMessage, locale string) (json.RawMessage, error) {
	var v interface{}
	if err := json.Unmarshal(layout, &v); err != nil {
		return nil, err
	}
	wrapped := wrapLocales(v, locale)
	return json.Marshal(wrapped)
}

func mergeLayoutLocales(incoming json.RawMessage, existing json.RawMessage, locale string) (json.RawMessage, error) {
	var inc interface{}
	if err := json.Unmarshal(incoming, &inc); err != nil {
		return nil, err
	}

	var exist interface{}
	if err := json.Unmarshal(existing, &exist); err != nil {
		return nil, err
	}

	existingLocales := make(map[int64]map[string]interface{})
	extractLocalesByID(exist, existingLocales)

	merged := mergeLocales(inc, existingLocales, locale)
	return json.Marshal(merged)
}
