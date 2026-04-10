package handlers

import (
	"errors"
	"net/http"
	"strings"

	"identity/app"
	"identity/domain"
)

var errParamTooLong = errors.New("parameter value exceeds maximum length")

func validateParamLengths(values map[string][]string) error {
	for _, vals := range values {
		for _, value := range vals {
			if len(value) > app.MaxParamLength {
				return errParamTooLong
			}
		}
	}

	return nil
}

func extractTenantID(request *http.Request) string {
	if tenantID := request.PathValue("tenant"); tenantID != "" {
		return tenantID
	}

	parts := strings.Split(strings.Trim(request.URL.Path, pathSeparator), pathSeparator)
	if len(parts) > minValidIndex {
		first := parts[minValidIndex]

		switch first {
		case "authorize", segmentCommon, "oauth2", "v2.0", "token",
			"login", "config", segmentMockUtils, "discovery", ".well-known", "health":
			return emptyValue
		default:
			return first
		}
	}

	return emptyValue
}

func extractBaseURL(request *http.Request) domain.BaseURL {
	scheme := "https"
	if request.TLS == nil && request.Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}

	return domain.MustBaseURL(scheme + "://" + request.Host)
}
