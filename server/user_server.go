// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"path"

	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/model"
)

func httpOAuth1Complete(jsi *jiraServerInstance, w http.ResponseWriter, r *http.Request) (int, error) {
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
	if err != nil {
		return http.StatusInternalServerError,
			errors.WithMessage(err, "failed to parse callback request from Jira")
	}

	requestSecret, err := jsi.Plugin.otsStore.LoadOneTimeSecret(requestToken)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = jsi.Plugin.otsStore.DeleteOneTimeSecret(requestToken)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	mattermostUserId := r.Header.Get("Mattermost-User-ID")
	if mattermostUserId == "" {
		return http.StatusUnauthorized, errors.New("not authorized")
	}
	mmuser, appErr := jsi.Plugin.API.GetUser(mattermostUserId)
	if appErr != nil {
		return http.StatusInternalServerError,
			errors.WithMessage(appErr, "failed to load user "+mattermostUserId)
	}

	oauth1Config, err := jsi.GetOAuth1Config()
	if err != nil {
		return http.StatusInternalServerError,
			errors.WithMessage(err, "failed to obtain oauth1 config")
	}

	accessToken, accessSecret, err := oauth1Config.AccessToken(requestToken, requestSecret, verifier)
	if err != nil {
		return http.StatusInternalServerError,
			errors.WithMessage(err, "failed to obtain oauth1 access token")
	}

	jiraUser := JIRAUser{
		Oauth1AccessToken:  accessToken,
		Oauth1AccessSecret: accessSecret,
	}

	jiraClient, err := jsi.GetJIRAClient(jiraUser)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	juser, _, err := jiraClient.User.GetSelf()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	jiraUser.User = *juser

	// Set default settings the first time a user connects
	jiraUser.Settings = &UserSettings{Notifications: true}

	err = jsi.Plugin.StoreUserInfoNotify(jsi, mattermostUserId, jiraUser)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return jsi.Plugin.respondWithTemplate(w, r, "text/html", struct {
		MattermostDisplayName string
		JiraDisplayName       string
		RevokeURL             string
	}{
		JiraDisplayName:       juser.DisplayName + " (" + juser.Name + ")",
		MattermostDisplayName: mmuser.GetDisplayName(model.SHOW_NICKNAME_FULLNAME),
		RevokeURL:             path.Join(jsi.Plugin.GetPluginURLPath(), routeUserDisconnect),
	})
}

func httpOAuth1PublicKey(p *Plugin, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed,
			errors.New("method " + r.Method + " is not allowed, must be GET")
	}

	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		return http.StatusUnauthorized, errors.New("not authorized")
	}

	if !p.API.HasPermissionTo(userID, model.PERMISSION_MANAGE_SYSTEM) {
		return http.StatusForbidden, errors.New("forbidden")
	}

	w.Header().Set("Content-Type", "text/plain")
	pkey, err := publicKeyString(p)
	if err != nil {
		return http.StatusInternalServerError,
			errors.WithMessage(err, "failed to load public key")
	}
	_, err = w.Write(pkey)
	if err != nil {
		return http.StatusInternalServerError,
			errors.WithMessage(err, "failed to write response")
	}
	return http.StatusOK, nil
}

func publicKeyString(p *Plugin) ([]byte, error) {
	rsaKey, err := p.secretsStore.EnsureRSAKey()
	if err != nil {
		return nil, err
	}

	b, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to encode public key")
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}), nil
}
