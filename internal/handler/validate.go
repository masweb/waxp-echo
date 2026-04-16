package handler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

var (
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	domainRegex     = regexp.MustCompile(`^(localhost|([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)+)$`)
	localeCodeRegex = regexp.MustCompile(`^[a-zA-Z]{2,3}(-[a-zA-Z]{2,4})?$`)
	slugRegex       = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?$`)
)

func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateDomain(domain string) error {
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format")
	}
	if len(domain) > 253 {
		return fmt.Errorf("domain must be at most 253 characters")
	}
	return nil
}

func validateLocaleCode(code string) error {
	if len(code) > 10 {
		return fmt.Errorf("locale code must be at most 10 characters")
	}
	if !localeCodeRegex.MatchString(code) {
		return fmt.Errorf("invalid locale code format (expected ISO 639 like 'en', 'es', 'pt-BR')")
	}
	return nil
}

func validateSlug(slug string) error {
	if slug == "" {
		return nil
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug '%s' contains invalid characters (only alphanumeric and hyphens)", slug)
	}
	return nil
}

func validateJSON(data json.RawMessage) error {
	if data == nil {
		return nil
	}
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON")
	}
	return nil
}
