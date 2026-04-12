package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"identity/domain"
)

const (
	uuidByteLen      = 16
	uuidVersionByte  = 6
	uuidVariantByte  = 8
	uuidVersionMask  = 0x40
	uuidVariantMask  = 0x80
	uuidVersionClear = 0x0f
	uuidVariantClear = 0x3f
	uuidHexStart     = 0
	uuidHexP1        = 8
	uuidHexP2        = 12
	uuidHexP3        = 16
	uuidHexP4        = 20
	uuidHexEnd       = 32
	uuidSep          = "-"
)

func newUUID() string {
	uuidBytes := make([]byte, uuidByteLen)

	_, err := rand.Read(uuidBytes)
	if err != nil {
		panic(err)
	}

	uuidBytes[uuidVersionByte] = (uuidBytes[uuidVersionByte] & uuidVersionClear) | uuidVersionMask
	uuidBytes[uuidVariantByte] = (uuidBytes[uuidVariantByte] & uuidVariantClear) | uuidVariantMask
	hexStr := hex.EncodeToString(uuidBytes)

	return hexStr[uuidHexStart:uuidHexP1] + uuidSep +
		hexStr[uuidHexP1:uuidHexP2] + uuidSep +
		hexStr[uuidHexP2:uuidHexP3] + uuidSep +
		hexStr[uuidHexP3:uuidHexP4] + uuidSep +
		hexStr[uuidHexP4:uuidHexEnd]
}

const contentTypePlain = "text/plain; charset=utf-8"

// Response is the result of a handler — status + body bytes + content type.
type Response struct {
	Status      int
	Body        []byte
	ContentType string
	Headers     map[string]string
}

func ok(body []byte, contentType string) Response {
	return Response{Status: http.StatusOK, Body: body, ContentType: contentType, Headers: nil}
}

func okJSON(body []byte) Response {
	return ok(body, "application/json")
}

func okText(body []byte) Response {
	return ok(body, "text/plain; charset=utf-8")
}

func okHTML(body []byte) Response {
	return ok(body, "text/html; charset=utf-8")
}

func badRequest(err error) Response {
	return oauthError("invalid_request", err.Error(), http.StatusBadRequest)
}

func notFound(msg string) Response {
	return Response{
		Status: http.StatusNotFound, Body: []byte(msg),
		ContentType: contentTypePlain, Headers: nil,
	}
}

func methodNotAllowed() Response {
	return Response{
		Status: http.StatusMethodNotAllowed, Body: []byte("method not allowed"),
		ContentType: contentTypePlain, Headers: nil,
	}
}

func internalError(msg string) Response {
	return Response{
		Status: http.StatusInternalServerError, Body: []byte(msg),
		ContentType: contentTypePlain, Headers: nil,
	}
}

func fromDomainError(domErr domain.Error) Response {
	httpStatus := domainErrorToHTTPStatus(domErr.Code)
	oauthCode := domainErrorToOAuthCode(domErr.Code)

	return oauthError(oauthCode, domErr.Message, httpStatus)
}

func domainErrorToOAuthCode(code domain.ErrorCode) string {
	switch code {
	case domain.ErrCodeInvalidCredentials,
		domain.ErrCodeInvalidGrant:
		return "invalid_grant"
	case domain.ErrCodeUnsupportedGrantType:
		return "unsupported_grant_type"
	case domain.ErrCodeInvalidRequest:
		return "invalid_request"
	case domain.ErrCodeTenantNotFound,
		domain.ErrCodeAppNotFound,
		domain.ErrCodeUserNotFound,
		domain.ErrCodeClientNotFound:
		return "invalid_client"
	default:
		return string(code)
	}
}

func domainErrorToHTTPStatus(code domain.ErrorCode) int {
	switch code {
	case domain.ErrCodeInvalidCredentials,
		domain.ErrCodeClientNotFound,
		domain.ErrCodeInvalidGrant:
		return http.StatusUnauthorized
	case domain.ErrCodeTenantNotFound,
		domain.ErrCodeAppNotFound,
		domain.ErrCodeUserNotFound:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func oauthError(code, description string, status int) Response {
	body := []byte(`{"error":"` + code + `",` +
		`"error_description":"` + description + `",` +
		`"correlation_id":"` + newUUID() + `",` +
		`"trace_id":"` + newUUID() + `"}`)

	return Response{Status: status, Body: body, ContentType: "application/json", Headers: nil}
}

// Write writes a Response to an http.ResponseWriter.
func Write(writer http.ResponseWriter, response Response) {
	if response.ContentType != "" {
		writer.Header().Set("Content-Type", response.ContentType)
	}

	for key, value := range response.Headers {
		writer.Header().Set(key, value)
	}

	writer.WriteHeader(response.Status)
	_, _ = writer.Write(response.Body)
}
