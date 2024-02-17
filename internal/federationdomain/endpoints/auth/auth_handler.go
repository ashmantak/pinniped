// Copyright 2020-2024 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package auth provides a handler for the OIDC authorization endpoint.
package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"k8s.io/utils/strings/slices"

	oidcapi "go.pinniped.dev/generated/latest/apis/supervisor/oidc"
	"go.pinniped.dev/internal/federationdomain/csrftoken"
	"go.pinniped.dev/internal/federationdomain/downstreamsession"
	"go.pinniped.dev/internal/federationdomain/federationdomainproviders"
	"go.pinniped.dev/internal/federationdomain/formposthtml"
	"go.pinniped.dev/internal/federationdomain/oidc"
	"go.pinniped.dev/internal/federationdomain/resolvedprovider"
	"go.pinniped.dev/internal/httputil/responseutil"
	"go.pinniped.dev/internal/httputil/securityheader"
	"go.pinniped.dev/internal/plog"
	"go.pinniped.dev/internal/psession"
	"go.pinniped.dev/pkg/oidcclient/nonce"
	"go.pinniped.dev/pkg/oidcclient/pkce"
)

const (
	promptParamName = "prompt"
	promptParamNone = "none"
)

type authorizeHandler struct {
	downstreamIssuerURL       string
	idpFinder                 federationdomainproviders.FederationDomainIdentityProvidersFinderI
	oauthHelperWithoutStorage fosite.OAuth2Provider
	oauthHelperWithStorage    fosite.OAuth2Provider
	generateCSRF              func() (csrftoken.CSRFToken, error)
	generatePKCE              func() (pkce.Code, error)
	generateNonce             func() (nonce.Nonce, error)
	upstreamStateEncoder      oidc.Encoder
	cookieCodec               oidc.Codec
}

func NewHandler(
	downstreamIssuerURL string,
	idpFinder federationdomainproviders.FederationDomainIdentityProvidersFinderI,
	oauthHelperWithoutStorage fosite.OAuth2Provider,
	oauthHelperWithStorage fosite.OAuth2Provider,
	generateCSRF func() (csrftoken.CSRFToken, error),
	generatePKCE func() (pkce.Code, error),
	generateNonce func() (nonce.Nonce, error),
	upstreamStateEncoder oidc.Encoder,
	cookieCodec oidc.Codec,
) http.Handler {
	h := &authorizeHandler{
		downstreamIssuerURL:       downstreamIssuerURL,
		idpFinder:                 idpFinder,
		oauthHelperWithoutStorage: oauthHelperWithoutStorage,
		oauthHelperWithStorage:    oauthHelperWithStorage,
		generateCSRF:              generateCSRF,
		generatePKCE:              generatePKCE,
		generateNonce:             generateNonce,
		upstreamStateEncoder:      upstreamStateEncoder,
		cookieCodec:               cookieCodec,
	}
	// During a response_mode=form_post auth request using the browser flow, the custom form_post html page may
	// be used to post certain errors back to the CLI from this handler's response, so allow the form_post
	// page's CSS and JS to run.
	return securityheader.WrapWithCustomCSP(h, formposthtml.ContentSecurityPolicy())
}

func (h *authorizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		// https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest
		// Authorization Servers MUST support the use of the HTTP GET and POST methods defined in
		// RFC 2616 [RFC2616] at the Authorization Endpoint.
		responseutil.HTTPErrorf(w, http.StatusMethodNotAllowed, "%s (try GET or POST)", r.Method)
		return
	}

	// The client set a username or password header, so they are trying to log in without using a browser.
	requestedBrowserlessFlow := len(r.Header.Values(oidcapi.AuthorizeUsernameHeaderName)) > 0 ||
		len(r.Header.Values(oidcapi.AuthorizePasswordHeaderName)) > 0

	// Need to parse the request params, so we can get the IDP name. The style and text of the error is inspired by
	// fosite's implementation of NewAuthorizeRequest(). Fosite only calls ParseMultipartForm() there. However,
	// although ParseMultipartForm() calls ParseForm(), it swallows errors from ParseForm() sometimes. To avoid
	// having any errors swallowed, we call both. When fosite calls ParseMultipartForm() later, it will be a noop.
	if err := r.ParseForm(); err != nil {
		oidc.WriteAuthorizeError(r, w,
			h.oauthHelperWithoutStorage,
			fosite.NewAuthorizeRequest(),
			fosite.ErrInvalidRequest.
				WithHint("Unable to parse form params, make sure to send a properly formatted query params or form request body.").
				WithWrap(err).WithDebug(err.Error()),
			requestedBrowserlessFlow)
		return
	}
	if err := r.ParseMultipartForm(1 << 20); err != nil && err != http.ErrNotMultipart {
		oidc.WriteAuthorizeError(r, w,
			h.oauthHelperWithoutStorage,
			fosite.NewAuthorizeRequest(),
			fosite.ErrInvalidRequest.
				WithHint("Unable to parse multipart HTTP body, make sure to send a properly formatted form request body.").
				WithWrap(err).WithDebug(err.Error()),
			requestedBrowserlessFlow)
		return
	}

	// Note that the client might have used oidcapi.AuthorizeUpstreamIDPNameParamName and
	// oidcapi.AuthorizeUpstreamIDPTypeParamName query (or form) params to request a certain upstream IDP.
	// The Pinniped CLI has been sending these params since v0.9.0.
	idpNameQueryParamValue := r.Form.Get(oidcapi.AuthorizeUpstreamIDPNameParamName)

	// Check if we are in a special case where we should inject an interstitial page to ask the user
	// which IDP they would like to use.
	if shouldShowIDPChooser(h.idpFinder, idpNameQueryParamValue, requestedBrowserlessFlow) {
		// Redirect to the IDP chooser page with all the same query/form params. When the user chooses an IDP,
		// it will redirect back to here with all the same params again, with the pinniped_idp_name param added.
		http.Redirect(w, r,
			fmt.Sprintf("%s%s?%s", h.downstreamIssuerURL, oidc.ChooseIDPEndpointPath, r.Form.Encode()),
			http.StatusSeeOther,
		)
		return
	}

	idp, err := chooseUpstreamIDP(idpNameQueryParamValue, h.idpFinder)
	if err != nil {
		oidc.WriteAuthorizeError(r, w,
			h.oauthHelperWithoutStorage,
			fosite.NewAuthorizeRequest(),
			fosite.ErrInvalidRequest.
				WithHintf("%q param error: %s", oidcapi.AuthorizeUpstreamIDPNameParamName, err.Error()).
				WithWrap(err).WithDebug(err.Error()),
			requestedBrowserlessFlow)
		return
	}

	h.authorize(w, r, requestedBrowserlessFlow, idpNameQueryParamValue, idp)
}

func (h *authorizeHandler) authorize(
	w http.ResponseWriter,
	r *http.Request,
	requestedBrowserlessFlow bool,
	idpNameQueryParamValue string,
	idp resolvedprovider.FederationDomainResolvedIdentityProvider,
) {
	// Browser flows do not need session storage at this step. For browser flows, the request parameters
	// should be forwarded to the next step as upstream state parameters to avoid storing session state
	// until the user successfully authenticates.
	oauthHelper := h.oauthHelperWithoutStorage
	if requestedBrowserlessFlow {
		oauthHelper = h.oauthHelperWithStorage
	}

	authorizeRequester, err := oauthHelper.NewAuthorizeRequest(r.Context(), r)
	if err != nil {
		oidc.WriteAuthorizeError(r, w, oauthHelper, authorizeRequester, err, requestedBrowserlessFlow)
		return
	}

	maybeLogDeprecationWarningForMissingIDPParam(idpNameQueryParamValue, authorizeRequester)

	// Automatically grant certain scopes, but only if they were requested.
	// Grant the openid scope (for now) if they asked for it so that `NewAuthorizeResponse` will perform its OIDC validations.
	// There don't seem to be any validations inside `NewAuthorizeResponse` related to the offline_access scope
	// at this time, however we will temporarily grant the scope just in case that changes in a future release of fosite.
	// This is instead of asking the user to approve these scopes. Note that `NewAuthorizeRequest` would have returned
	// an error if the client requested a scope that they are not allowed to request, so we don't need to worry about that here.
	downstreamsession.AutoApproveScopes(authorizeRequester)

	if requestedBrowserlessFlow {
		err = h.authorizeWithoutBrowser(r, w, oauthHelper, authorizeRequester, idp)
	} else {
		err = h.authorizeWithBrowser(r, w, oauthHelper, authorizeRequester, idp)
	}
	if err != nil {
		oidc.WriteAuthorizeError(r, w, oauthHelper, authorizeRequester, err, requestedBrowserlessFlow)
	}
}

func (h *authorizeHandler) authorizeWithoutBrowser(
	r *http.Request,
	w http.ResponseWriter,
	oauthHelper fosite.OAuth2Provider,
	authorizeRequester fosite.AuthorizeRequester,
	idp resolvedprovider.FederationDomainResolvedIdentityProvider,
) error {
	if err := requireStaticClientForUsernameAndPasswordHeaders(authorizeRequester); err != nil {
		return err
	}

	submittedUsername, submittedPassword, err := requireNonEmptyUsernameAndPasswordHeaders(r)
	if err != nil {
		return err
	}

	groupsWillBeIgnored := !slices.Contains(authorizeRequester.GetGrantedScopes(), oidcapi.ScopeGroups)

	identity, loginExtras, err := idp.Login(r.Context(), submittedUsername, submittedPassword, groupsWillBeIgnored)
	if err != nil {
		return err
	}

	username, groups, err := downstreamsession.ApplyIdentityTransformations(r.Context(),
		idp.GetTransforms(), identity.UpstreamUsername, identity.UpstreamGroups)
	if err != nil {
		return fosite.ErrAccessDenied.WithHintf("Reason: %s.", err.Error())
	}

	session := downstreamsession.NewPinnipedSession(idp, &downstreamsession.SessionConfig{
		UpstreamIdentity:    identity,
		UpstreamLoginExtras: loginExtras,
		Username:            username,
		Groups:              groups,
		ClientID:            authorizeRequester.GetClient().GetID(),
		GrantedScopes:       authorizeRequester.GetGrantedScopes(),
	})

	oidc.PerformAuthcodeRedirect(r, w, oauthHelper, authorizeRequester, session, true)

	return nil
}

func (h *authorizeHandler) authorizeWithBrowser(
	r *http.Request,
	w http.ResponseWriter,
	oauthHelper fosite.OAuth2Provider,
	authorizeRequester fosite.AuthorizeRequester,
	idp resolvedprovider.FederationDomainResolvedIdentityProvider,
) error {
	authRequestState, err := generateUpstreamAuthorizeRequestState(r, w,
		authorizeRequester,
		oauthHelper,
		h.generateCSRF,
		h.generateNonce,
		h.generatePKCE,
		idp.GetDisplayName(),
		idp.GetSessionProviderType(),
		h.cookieCodec,
		h.upstreamStateEncoder,
	)
	if err != nil {
		return err
	}

	redirectURL, err := idp.UpstreamAuthorizeRedirectURL(authRequestState, h.downstreamIssuerURL)
	if err != nil {
		return err
	}

	http.Redirect(w, r, redirectURL,
		http.StatusSeeOther, // match fosite and https://tools.ietf.org/id/draft-ietf-oauth-security-topics-18.html#section-4.11
	)

	return nil
}

func shouldShowIDPChooser(
	idpFinder federationdomainproviders.FederationDomainIdentityProvidersFinderI,
	idpNameQueryParamValue string,
	requestedBrowserlessFlow bool,
) bool {
	clientDidNotRequestSpecificIDP := len(idpNameQueryParamValue) == 0
	clientRequestedBrowserBasedFlow := !requestedBrowserlessFlow
	inBackwardsCompatMode := idpFinder.HasDefaultIDP()
	federationDomainSpecHasSomeValidIDPs := idpFinder.IDPCount() > 0

	return clientDidNotRequestSpecificIDP && clientRequestedBrowserBasedFlow &&
		!inBackwardsCompatMode && federationDomainSpecHasSomeValidIDPs
}

func requireStaticClientForUsernameAndPasswordHeaders(authorizeRequester fosite.AuthorizeRequester) error {
	if !(authorizeRequester.GetClient().GetID() == oidcapi.ClientIDPinnipedCLI) {
		return fosite.ErrAccessDenied.WithHint("This client is not allowed to submit username or password headers to this endpoint.")
	}
	return nil
}

func requireNonEmptyUsernameAndPasswordHeaders(r *http.Request) (string, string, error) {
	username := r.Header.Get(oidcapi.AuthorizeUsernameHeaderName)
	password := r.Header.Get(oidcapi.AuthorizePasswordHeaderName)
	if username == "" || password == "" {
		return "", "", fosite.ErrAccessDenied.WithHint("Missing or blank username or password.")
	}
	return username, password, nil
}

func readCSRFCookie(r *http.Request, codec oidc.Decoder) csrftoken.CSRFToken {
	receivedCSRFCookie, err := r.Cookie(oidc.CSRFCookieName)
	if err != nil {
		// Error means that the cookie was not found
		return ""
	}

	var csrfFromCookie csrftoken.CSRFToken
	err = codec.Decode(oidc.CSRFCookieEncodingName, receivedCSRFCookie.Value, &csrfFromCookie)
	if err != nil {
		// We can ignore any errors and just make a new cookie. Hopefully this will
		// make the user experience better if, for example, the server rotated
		// cookie signing keys and then a user submitted a very old cookie.
		return ""
	}

	return csrfFromCookie
}

// chooseUpstreamIDP selects either an OIDC, an LDAP, or an AD IDP, or returns an error.
// Note that AD and LDAP IDPs both return the same interface type, but different ProviderTypes values.
func chooseUpstreamIDP(idpDisplayName string, idpLister federationdomainproviders.FederationDomainIdentityProvidersFinderI) (
	resolvedprovider.FederationDomainResolvedIdentityProvider,
	error,
) {
	// When a request is made to the authorization endpoint which does not specify the IDP name, then it might
	// be an old dynamic client (OIDCClient). We need to make this work, but only in the backwards compatibility case
	// where there is exactly one IDP defined in the namespace and no IDPs listed on the FederationDomain.
	// This backwards compatibility mode is handled by FindDefaultIDP().
	if len(idpDisplayName) == 0 {
		return idpLister.FindDefaultIDP()
	}
	return idpLister.FindUpstreamIDPByDisplayName(idpDisplayName)
}

func maybeLogDeprecationWarningForMissingIDPParam(idpNameQueryParamValue string, authorizeRequester fosite.AuthorizeRequester) {
	if len(idpNameQueryParamValue) != 0 {
		return
	}
	plog.Warning("Client attempted to perform an authorization flow (user login) without specifying the "+
		"query param to choose an identity provider. "+
		"This will not work when identity providers are configured explicitly on a FederationDomain. "+
		"Additionally, this behavior is deprecated and support for any authorization requests missing this query param "+
		"may be removed in a future release. "+
		"Please ask the author of this client to update the authorization request URL to include this query parameter. "+
		"The value of the parameter should be equal to the displayName of the identity provider as declared in the FederationDomain.",
		"missingParameterName", oidcapi.AuthorizeUpstreamIDPNameParamName,
		"clientID", authorizeRequester.GetClient().GetID(),
	)
}

// generateUpstreamAuthorizeRequestState performs the shared validations and setup between browser based
// auth requests regardless of IDP type.
// It generates the state param, sets the CSRF cookie, and validates the prompt param.
// It returns an error when it encounters an error without handling it, leaving it to
// the caller to decide how to handle it.
// It returns nil with no error when it encounters an error and also has already handled writing
// the error response to the ResponseWriter, in which case the caller should not also try to
// write the error response.
func generateUpstreamAuthorizeRequestState(
	r *http.Request,
	w http.ResponseWriter,
	authorizeRequester fosite.AuthorizeRequester,
	oauthHelper fosite.OAuth2Provider,
	generateCSRF func() (csrftoken.CSRFToken, error),
	generateNonce func() (nonce.Nonce, error),
	generatePKCE func() (pkce.Code, error),
	upstreamDisplayName string,
	idpType psession.ProviderType,
	cookieCodec oidc.Codec,
	upstreamStateEncoder oidc.Encoder,
) (*resolvedprovider.UpstreamAuthorizeRequestState, error) {
	now := time.Now()
	_, err := oauthHelper.NewAuthorizeResponse(r.Context(), authorizeRequester, &psession.PinnipedSession{
		Fosite: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				// Temporary claim values to allow `NewAuthorizeResponse` to perform other OIDC validations.
				Subject:     "none",
				AuthTime:    now,
				RequestedAt: now,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	csrfValue, nonceValue, pkceValue, err := generateValues(generateCSRF, generateNonce, generatePKCE)
	if err != nil {
		plog.Error("authorize generate error", err)
		return nil, fosite.ErrServerError.WithHint("Server could not generate necessary values.").WithWrap(err)
	}
	csrfFromCookie := readCSRFCookie(r, cookieCodec)
	if csrfFromCookie != "" {
		csrfValue = csrfFromCookie
	}

	encodedStateParamValue, err := upstreamStateParam(
		authorizeRequester,
		upstreamDisplayName,
		string(idpType),
		nonceValue,
		csrfValue,
		pkceValue,
		upstreamStateEncoder,
	)
	if err != nil {
		plog.Error("authorize upstream state param error", err)
		return nil, fosite.ErrServerError.WithHint("Error encoding upstream state param.").WithWrap(err)
	}

	promptParam := authorizeRequester.GetRequestForm().Get(promptParamName)
	if promptParam == promptParamNone && oidc.ScopeWasRequested(authorizeRequester, oidcapi.ScopeOpenID) {
		return nil, fosite.ErrLoginRequired
	}

	if csrfFromCookie == "" {
		// We did not receive an incoming CSRF cookie, so write a new one.
		err = addCSRFSetCookieHeader(w, csrfValue, cookieCodec)
		if err != nil {
			plog.Error("error setting CSRF cookie", err)
			return nil, fosite.ErrServerError.WithHint("Error encoding CSRF cookie.").WithWrap(err)
		}
	}

	return &resolvedprovider.UpstreamAuthorizeRequestState{
		EncodedStateParam: encodedStateParamValue,
		PKCE:              pkceValue,
		Nonce:             nonceValue,
	}, nil
}

func generateValues(
	generateCSRF func() (csrftoken.CSRFToken, error),
	generateNonce func() (nonce.Nonce, error),
	generatePKCE func() (pkce.Code, error),
) (csrftoken.CSRFToken, nonce.Nonce, pkce.Code, error) {
	csrfValue, err := generateCSRF()
	if err != nil {
		return "", "", "", fmt.Errorf("error generating CSRF token: %w", err)
	}
	nonceValue, err := generateNonce()
	if err != nil {
		return "", "", "", fmt.Errorf("error generating nonce param: %w", err)
	}
	pkceValue, err := generatePKCE()
	if err != nil {
		return "", "", "", fmt.Errorf("error generating PKCE param: %w", err)
	}
	return csrfValue, nonceValue, pkceValue, nil
}

func upstreamStateParam(
	authorizeRequester fosite.AuthorizeRequester,
	upstreamDisplayName string,
	upstreamType string,
	nonceValue nonce.Nonce,
	csrfValue csrftoken.CSRFToken,
	pkceValue pkce.Code,
	encoder oidc.Encoder,
) (string, error) {
	stateParamData := oidc.UpstreamStateParamData{
		// The auth params might have included oidcapi.AuthorizeUpstreamIDPNameParamName and
		// oidcapi.AuthorizeUpstreamIDPTypeParamName, but those can be ignored by other handlers
		// that are reading from the encoded upstream state param being built here.
		// The UpstreamName and UpstreamType struct fields can be used instead.
		// Remove those params here to avoid potential confusion about which should be used later.
		AuthParams:    removeCustomIDPParams(authorizeRequester.GetRequestForm()).Encode(),
		UpstreamName:  upstreamDisplayName,
		UpstreamType:  upstreamType,
		Nonce:         nonceValue,
		CSRFToken:     csrfValue,
		PKCECode:      pkceValue,
		FormatVersion: oidc.UpstreamStateParamFormatVersion,
	}
	encodedStateParamValue, err := encoder.Encode(oidc.UpstreamStateParamEncodingName, stateParamData)
	if err != nil {
		return "", fmt.Errorf("error encoding upstream state param: %w", err)
	}
	return encodedStateParamValue, nil
}

func removeCustomIDPParams(params url.Values) url.Values {
	p := url.Values{}
	// Copy all params.
	for k, v := range params {
		p[k] = v
	}
	// Remove the unnecessary params.
	delete(p, oidcapi.AuthorizeUpstreamIDPNameParamName)
	delete(p, oidcapi.AuthorizeUpstreamIDPTypeParamName)
	return p
}

func addCSRFSetCookieHeader(w http.ResponseWriter, csrfValue csrftoken.CSRFToken, codec oidc.Encoder) error {
	encodedCSRFValue, err := codec.Encode(oidc.CSRFCookieEncodingName, csrfValue)
	if err != nil {
		return fmt.Errorf("error encoding CSRF cookie: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oidc.CSRFCookieName,
		Value:    encodedCSRFValue,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Path:     "/",
	})

	return nil
}
