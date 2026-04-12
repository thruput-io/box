package handlers

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

const (
	formKeyState             = "state"
	emptyType                = ""
	errInvalidCredentialsMsg = "invalid username or password"
)

func authorizeHandler(request *http.Request, application *app.App) Response {
	data, domErrOpt := buildAuthorizeLoginTemplateData(request, application)
	if domErr, ok := domErrOpt.Get(); ok {
		return badRequest(domErr)
	}

	var buf bytes.Buffer

	err := application.LoginTemplate.Execute(&buf, data)
	if err != nil {
		return internalError("failed to render login page")
	}

	return okHTML(buf.Bytes())
}

type validatedLogin struct {
	tenant      *domain.Tenant
	clientID    domain.ClientID
	redirectURI domain.RedirectURL
}

func emptyLoginTemplateData() loginTemplateData {
	return loginTemplateData{
		ClientID:     domain.ClientID{},
		RedirectURI:  domain.RedirectURL{},
		State:        mo.None[domain.OAuthState](),
		Scope:        nil,
		ResponseType: mo.None[domain.ResponseType](),
		Tenant:       mo.None[domain.TenantID](),
		Nonce:        mo.None[domain.Nonce](),
		ResponseMode: mo.None[domain.ResponseMode](),
		Users:        nil,
	}
}

func buildAuthorizeLoginTemplateData(request *http.Request, application *app.App) (loginTemplateData, mo.Option[domain.Error]) {
	err := validateParamLengths(request.URL.Query())
	if err != nil {
		return emptyLoginTemplateData(), mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Error()))
	}

	clientIDStr := request.URL.Query().Get("client_id")
	tenantIDStr := extractTenantID(request)

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return emptyLoginTemplateData(), mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Error()))
	}

	clientIDEither := domain.NewClientID(clientIDStr)

	clientID, ok := clientIDEither.Right()

	if !ok {
		err, _ := clientIDEither.Left()

		return emptyLoginTemplateData(), mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Message))
	}

	redirectURIStr := request.URL.Query().Get("redirect_uri")

	redirectURIEither := domain.NewRedirectURL(redirectURIStr)

	redirectURI, ok := redirectURIEither.Right()

	if !ok {
		err, _ := redirectURIEither.Left()

		return emptyLoginTemplateData(), mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Message))
	}

	allowedURLs, err := app.FindRedirectURLs(*tenant, clientID)
	if err != nil {
		return emptyLoginTemplateData(), mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Error()))
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return emptyLoginTemplateData(), mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Error()))
	}

	tenantOpt := parseTenantIDOption(tenantIDStr)

	data := loginTemplateData{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		State:        parseOptionalOAuthState(request.URL.Query().Get("state")),
		Scope:        parseScopeValues(request.URL.Query().Get("scope")),
		ResponseType: parseOptionalResponseType(request.URL.Query().Get("response_type")),
		Tenant:       tenantOpt,
		Nonce:        parseOptionalNonce(request.URL.Query().Get("nonce")),
		ResponseMode: parseOptionalResponseMode(request.URL.Query().Get("response_mode")),
		Users:        buildUsersDisplay(*tenant, clientID),
	}

	return data, mo.None[domain.Error]()
}

func validateLoginRequest(request *http.Request, application *app.App) (validatedLogin, mo.Option[domain.Error]) {
	tenantIDStr := request.Form.Get("tenant")

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return validatedLogin{tenant: nil, clientID: domain.ClientID{}, redirectURI: domain.RedirectURL{}}, mo.Some(domain.ErrTenantNotFound)
	}

	clientIDEither := domain.NewClientID(request.Form.Get("client_id"))

	clientID, ok := clientIDEither.Right()

	if !ok {
		err, _ := clientIDEither.Left()

		return validatedLogin{tenant: nil, clientID: domain.ClientID{}, redirectURI: domain.RedirectURL{}}, mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Message))
	}

	redirectURIStr := request.Form.Get("redirect_uri")

	redirectURIEither := domain.NewRedirectURL(redirectURIStr)

	redirectURI, ok := redirectURIEither.Right()

	if !ok {
		err, _ := redirectURIEither.Left()

		return validatedLogin{tenant: nil, clientID: domain.ClientID{}, redirectURI: domain.RedirectURL{}}, mo.Some(domain.NewError(domain.ErrCodeInvalidRequest, err.Message))
	}

	allowedURLs, err := app.FindRedirectURLs(*tenant, clientID)
	if err != nil {
		return validatedLogin{tenant: nil, clientID: domain.ClientID{}, redirectURI: domain.RedirectURL{}}, mo.Some(domain.ErrClientNotFound)
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return validatedLogin{tenant: nil, clientID: domain.ClientID{}, redirectURI: domain.RedirectURL{}}, mo.Some(domain.ErrInvalidRedirectURI)
	}

	return validatedLogin{tenant: tenant, clientID: clientID, redirectURI: redirectURI}, mo.None[domain.Error]()
}

func loginHandler(request *http.Request, application *app.App) Response {
	if request.Method != http.MethodPost {
		return methodNotAllowed()
	}

	request.Body = http.MaxBytesReader(nil, request.Body, app.MaxBodyBytes)

	err := request.ParseForm()
	if err != nil {
		return badRequest(err)
	}

	paramErr := validateParamLengths(request.Form)
	if paramErr != nil {
		return badRequest(paramErr)
	}

	validated, domErrOpt := validateLoginRequest(request, application)
	if domErr, ok := domErrOpt.Get(); ok {
		return fromDomainError(domErr)
	}

	user, domErrOpt := authenticateLoginRequest(request, *validated.tenant)
	if domErr, ok := domErrOpt.Get(); ok {
		return fromDomainError(domErr)
	}

	authCode := app.IssueAuthCode(
		application.Key, *user, validated.clientID, validated.redirectURI,
		parseScopeValues(request.Form.Get("scope")),
		validated.tenant.TenantID(),
		parseOptionalNonce(request.Form.Get("nonce")),
	)

	target := validated.redirectURI.AsURL()

	return buildLoginRedirect(
		&target, authCode,
		parseOptionalOAuthState(request.Form.Get("state")),
		parseOptionalResponseMode(request.Form.Get("response_mode")),
	)
}

func authenticateLoginRequest(request *http.Request, tenant domain.Tenant) (*domain.User, mo.Option[domain.Error]) {
	usernameEither := domain.NewUsername(request.Form.Get("username"))

	username, ok := usernameEither.Right()
	if !ok {
		return nil, mo.Some(domain.NewError(domain.ErrCodeInvalidCredentials, errInvalidCredentialsMsg))
	}

	passwordEither := domain.NewPassword(request.Form.Get("password"))

	password, ok := passwordEither.Right()
	if !ok {
		return nil, mo.Some(domain.NewError(domain.ErrCodeInvalidCredentials, errInvalidCredentialsMsg))
	}

	user, err := app.AuthenticateUser(tenant, username, password)
	if err != nil {
		return nil, mo.Some(domain.NewError(domain.ErrCodeInvalidCredentials, errInvalidCredentialsMsg))
	}

	return user, mo.None[domain.Error]()
}

func buildLoginRedirect(
	target *url.URL,
	authCode domain.AuthCode,
	state mo.Option[domain.OAuthState],
	responseMode mo.Option[domain.ResponseMode],
) Response {
	values := url.Values{}
	authCode.AddTo(values)

	if s, ok := state.Get(); ok {
		values.Set(formKeyState, s.Value())
	}

	var location string

	isFragment := false
	if rm, ok := responseMode.Get(); ok {
		isFragment = rm.Value() == "fragment"
	}

	if isFragment {
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

	return Response{
		Status:      http.StatusFound,
		ContentType: emptyType,
		Body:        nil,
		Headers:     map[string]string{"Location": location},
	}
}

// loginTemplateData is the data passed to the login template.
type loginTemplateData struct {
	ClientID     domain.ClientID
	RedirectURI  domain.RedirectURL
	State        mo.Option[domain.OAuthState]
	Scope        []domain.ScopeValue
	ResponseType mo.Option[domain.ResponseType]
	Tenant       mo.Option[domain.TenantID]
	Nonce        mo.Option[domain.Nonce]
	ResponseMode mo.Option[domain.ResponseMode]
	Users        []userDisplay
}

// JoinedScope returns the scope values as a space-separated string for template rendering.
func (d loginTemplateData) JoinedScope() string {
	return domain.JoinScopeValues(d.Scope)
}

// StateValue returns the OAuth state string for template rendering, or empty string if absent.
func (d loginTemplateData) StateValue() string {
	if s, ok := d.State.Get(); ok {
		return s.Value()
	}

	return emptyValue
}

// TenantValue returns the tenant ID string for template rendering, or empty string if absent (common).
func (d loginTemplateData) TenantValue() string {
	if t, ok := d.Tenant.Get(); ok {
		return t.Value()
	}

	return emptyValue
}

// NonceValue returns the nonce string for template rendering, or empty string if absent.
func (d loginTemplateData) NonceValue() string {
	if n, ok := d.Nonce.Get(); ok {
		return n.Value()
	}

	return emptyValue
}

// ResponseModeValue returns the response_mode string for template rendering, or empty string if absent.
func (d loginTemplateData) ResponseModeValue() string {
	if rm, ok := d.ResponseMode.Get(); ok {
		return rm.Value()
	}

	return emptyValue
}

// userDisplay is a view model for a user on the login page.
type userDisplay struct {
	Username    domain.Username
	Password    domain.Password
	DisplayName domain.DisplayName
	Roles       []domain.NonEmptyString
}

func buildUsersDisplay(tenant domain.Tenant, clientID domain.ClientID) []userDisplay {
	var activeClient *domain.Client

	client, err := app.FindClient(tenant, clientID)
	if err == nil {
		activeClient = client
	}

	result := make([]userDisplay, emptySliceSize, len(tenant.Users()))

	for _, user := range tenant.Users() {
		display := user.Display()
		result = append(result, userDisplay{
			Username:    display.Username,
			Password:    display.Password,
			DisplayName: display.DisplayName,
			Roles:       resolveDisplayRoles(user, activeClient, tenant),
		})
	}

	return result
}

func resolveDisplayRoles(user domain.User, client *domain.Client, tenant domain.Tenant) []domain.NonEmptyString {
	if client == nil {
		return nil
	}

	appRoles := collectAssignmentRoles(user, *client)
	result := make([]domain.NonEmptyString, emptySliceSize, len(appRoles))

	for appID, roles := range appRoles {
		appName := resolveAppName(tenant, appID)
		roleStrs := make([]string, len(roles))

		for i, r := range roles {
			roleStrs[i] = r.Value()
		}

		display := appName.Value() + ": " + strings.Join(roleStrs, ", ")
		result = append(result, domain.NewNonEmptyString(display).MustRight())
	}

	return result
}

func collectAssignmentRoles(user domain.User, client domain.Client) map[domain.ClientID][]domain.RoleValue {
	userGroups := make(map[domain.GroupName]bool)

	for _, groupName := range user.Groups() {
		userGroups[groupName] = true
	}

	appRoles := make(map[domain.ClientID][]domain.RoleValue)

	for _, assignment := range client.GroupRoleAssignments() {
		if !userGroups[assignment.GroupName()] {
			continue
		}

		addAssignmentRoles(appRoles, assignment)
	}

	return appRoles
}

func addAssignmentRoles(appRoles map[domain.ClientID][]domain.RoleValue, assignment domain.GroupRoleAssignment) {
	appID := assignment.ApplicationID()
	appRoles[appID] = append(appRoles[appID], assignment.Roles()...)
}

func resolveAppName(tenant domain.Tenant, appID domain.ClientID) domain.AppName {
	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID() == appID {
			return registration.Name()
		}
	}

	return domain.MustAppName(appID.Value())
}

// parseTenantIDOption wraps the raw tenant ID string in mo.Option[domain.TenantID].
// Returns None for "common" or empty string (resolved to first tenant by FindTenant).
func parseTenantIDOption(raw string) mo.Option[domain.TenantID] {
	if raw == emptyValue || raw == segmentCommon {
		return mo.None[domain.TenantID]()
	}

	tid, ok := domain.NewTenantID(raw).Right()
	if !ok {
		return mo.None[domain.TenantID]()
	}

	return mo.Some(tid)
}
