package http

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"identity/app"
	"identity/domain"
)

func authorizeHandler(request *http.Request, server *Server) HTTPResponse {
	if err := validateParamLengths(request.URL.Query()); err != nil {
		return badRequest(err)
	}

	clientIDStr := request.URL.Query().Get("client_id")
	tenantIDStr := extractTenantID(request)

	tenant, err := app.FindTenant(server.Config, tenantIDStr)
	if err != nil {
		return badRequest(err)
	}

	clientID, err := domain.NewClientID(clientIDStr)
	if err != nil {
		return badRequest(err)
	}

	redirectURI := request.URL.Query().Get("redirect_uri")

	allowedURLs, err := app.FindRedirectURLs(tenant, clientID)
	if err != nil {
		return badRequest(err)
	}

	if err = app.ValidateRedirectURI(redirectURI, allowedURLs); err != nil {
		return badRequest(err)
	}

	data := loginTemplateData{
		ClientID:     clientIDStr,
		RedirectURI:  redirectURI,
		State:        request.URL.Query().Get("state"),
		Scope:        request.URL.Query().Get("scope"),
		ResponseType: request.URL.Query().Get("response_type"),
		Tenant:       tenantIDStr,
		Nonce:        request.URL.Query().Get("nonce"),
		ResponseMode: request.URL.Query().Get("response_mode"),
		Users:        buildUsersDisplay(tenant, clientID),
	}

	var buf bytes.Buffer

	if err = server.LoginTemplate.Execute(&buf, data); err != nil {
		return internalError("failed to render login page")
	}

	return okHTML(buf.Bytes())
}

func loginHandler(request *http.Request, server *Server) HTTPResponse {
	if request.Method != http.MethodPost {
		return methodNotAllowed()
	}

	request.Body = http.MaxBytesReader(nil, request.Body, app.MaxBodyBytes)

	if err := request.ParseForm(); err != nil {
		return badRequest(err)
	}

	if err := validateParamLengths(request.Form); err != nil {
		return badRequest(err)
	}

	tenantIDStr := request.Form.Get("tenant")

	tenant, err := app.FindTenant(server.Config, tenantIDStr)
	if err != nil {
		return badRequest(err)
	}

	clientID, err := domain.NewClientID(request.Form.Get("client_id"))
	if err != nil {
		return badRequest(err)
	}

	redirectURI := request.Form.Get("redirect_uri")

	allowedURLs, err := app.FindRedirectURLs(tenant, clientID)
	if err != nil {
		return badRequest(err)
	}

	if err = app.ValidateRedirectURI(redirectURI, allowedURLs); err != nil {
		return badRequest(err)
	}

	user, err := app.AuthenticateUser(tenant, request.Form.Get("username"), request.Form.Get("password"))
	if err != nil {
		return unauthorized(err)
	}

	authCode := app.IssueAuthCode(server.Key, user, clientID, redirectURI, request.Form.Get("scope"), tenantIDStr, request.Form.Get("nonce"))

	target, err := url.Parse(redirectURI)
	if err != nil {
		return badRequest(err)
	}

	return buildLoginRedirect(target, authCode, request.Form.Get("state"), request.Form.Get("response_mode"))
}

func buildLoginRedirect(target *url.URL, authCode, state, responseMode string) HTTPResponse {
	values := url.Values{}
	values.Set("code", authCode)

	if state != "" {
		values.Set("state", state)
	}

	var location string

	if responseMode == "fragment" {
		target.RawQuery = ""
		location = target.String() + "#" + values.Encode()
	} else {
		query := target.Query()
		for key, vals := range values {
			query.Set(key, vals[0])
		}

		target.RawQuery = query.Encode()
		location = target.String()
	}

	return HTTPResponse{
		Status:      http.StatusFound,
		ContentType: "",
		Body:        nil,
		Headers:     map[string]string{"Location": location},
	}
}

// loginTemplateData is the data passed to the login template.
type loginTemplateData struct {
	ClientID     string
	RedirectURI  string
	State        string
	Scope        string
	ResponseType string
	Tenant       string
	Nonce        string
	ResponseMode string
	Users        []userDisplay
}

// userDisplay is a view model for a user on the login page.
type userDisplay struct {
	Username    string
	Password    string
	DisplayName string
	Roles       []string
}

func buildUsersDisplay(tenant domain.Tenant, clientID domain.ClientID) []userDisplay {
	activeClient, _ := app.FindClient(tenant, clientID)
	hasClient := activeClient.ClientID().UUID() != [16]byte{}

	result := make([]userDisplay, 0, len(tenant.Users()))

	for _, user := range tenant.Users() {
		result = append(result, userDisplay{
			Username:    user.Username().String(),
			Password:    user.Password().String(),
			DisplayName: user.DisplayName().String(),
			Roles:       resolveDisplayRoles(user, activeClient, tenant, hasClient),
		})
	}

	return result
}

func resolveDisplayRoles(user domain.User, client domain.Client, tenant domain.Tenant, hasClient bool) []string {
	if !hasClient {
		return nil
	}

	appRoles := make(map[string][]string)

	for _, assignment := range client.GroupRoleAssignments() {
		for _, groupName := range user.Groups() {
			if assignment.GroupName().String() != groupName.String() {
				continue
			}

			appIDStr := assignment.ApplicationID().String()
			for _, roleValue := range assignment.Roles() {
				appRoles[appIDStr] = append(appRoles[appIDStr], roleValue.String())
			}
		}
	}

	result := make([]string, 0, len(appRoles))

	for appIDStr, roles := range appRoles {
		appName := appIDStr

		for _, registration := range tenant.AppRegistrations() {
			if registration.ClientID().String() == appIDStr {
				appName = registration.Name().String()

				break
			}
		}

		result = append(result, appName+": "+strings.Join(roles, ", "))
	}

	return result
}
