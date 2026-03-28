package http

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

	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(parts) > 0 {
		first := parts[0]

		switch first {
		case "authorize", "common", "oauth2", "v2.0", "token", "login", "config", "test-tokens", "discovery", ".well-known", "health":
			return ""
		default:
			return first
		}
	}

	return ""
}

func extractBaseURL(request *http.Request) string {
	scheme := "https"
	if request.TLS == nil && request.Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}

	return scheme + "://" + request.Host
}

func tokenError(err error) HTTPResponse {
	switch {
	case errors.Is(err, domain.ErrTenantNotFound):
		return badRequest(err)
	case errors.Is(err, domain.ErrUnsupportedGrantType):
		return badRequest(err)
	case errors.Is(err, domain.ErrInvalidRedirectURI):
		return badRequest(err)
	case errors.Is(err, domain.ErrInvalidCredentials):
		return unauthorized(err)
	case errors.Is(err, domain.ErrClientNotFound):
		return invalidClient(err)
	default:
		return badRequest(err)
	}
}
