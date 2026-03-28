package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"identity/app"
)

func tokenHandler(request *http.Request, server *Server) HTTPResponse {
	if err := request.ParseForm(); err != nil {
		return badRequest(err)
	}

	if err := validateParamLengths(request.Form); err != nil {
		return badRequest(err)
	}

	raw := app.RawTokenRequest{
		GrantType:     request.Form.Get("grant_type"),
		ClientID:      request.Form.Get("client_id"),
		ClientSecret:  request.Form.Get("client_secret"),
		Username:      request.Form.Get("username"),
		Password:      request.Form.Get("password"),
		Scope:         request.Form.Get("scope"),
		Code:          request.Form.Get("code"),
		RedirectURI:   request.Form.Get("redirect_uri"),
		RefreshToken:  request.Form.Get("refresh_token"),
		Nonce:         request.Form.Get("nonce"),
		TenantID:      extractTenantID(request),
		IsV2:          strings.Contains(request.URL.Path, "/v2.0"),
		CorrelationID: correlationID(request),
		BaseURL:       extractBaseURL(request),
	}

	response, err := app.IssueToken(server.Config, server.Key, raw)
	if err != nil {
		return tokenError(err)
	}

	return encodeTokenResponse(response)
}

func encodeTokenResponse(response app.TokenResponse) HTTPResponse {
	body := map[string]any{
		"access_token": response.AccessToken,
		"token_type":   response.TokenType,
		"expires_in":   response.ExpiresIn,
		"scope":        response.Scope,
	}

	if response.IDToken != nil {
		body["id_token"] = *response.IDToken
	}

	if response.RefreshToken != nil {
		body["refresh_token"] = *response.RefreshToken
	}

	if response.ClientInfo != nil {
		body["client_info"] = *response.ClientInfo
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return internalError("failed to encode token response")
	}

	return HTTPResponse{
		Status:      http.StatusOK,
		Body:        encoded,
		ContentType: "application/json",
		Headers: map[string]string{
			"Client-Request-Id": response.CorrelationID,
			"X-Ms-Request-Id":   response.CorrelationID,
		},
	}
}

func correlationID(request *http.Request) string {
	if id := request.Header.Get("Client-Request-Id"); id != "" {
		return id
	}

	return uuid.New().String()
}
