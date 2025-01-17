// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

type jiraTestInstance struct {
	JIRAInstance
}

var _ Instance = (*jiraTestInstance)(nil)

func (jti jiraTestInstance) GetURL() string {
	return "http://jiraTestInstanceURL.some"
}
func (jti jiraTestInstance) GetMattermostKey() string {
	return "jiraTestInstanceMattermostKey"
}
func (jti jiraTestInstance) GetDisplayDetails() map[string]string {
	return map[string]string{}
}
func (jti jiraTestInstance) GetUserConnectURL(mattermostUserId string) (string, error) {
	return "http://jiraTestInstanceUserConnectURL.some", nil
}
func (jti jiraTestInstance) GetJIRAClient(jiraUser JIRAUser) (*jira.Client, error) {
	return nil, errors.New("not implemented")
}

type mockCurrentInstanceStore struct {
	plugin *Plugin
}

func (store mockCurrentInstanceStore) StoreCurrentJIRAInstance(ji Instance) error {
	return nil
}
func (store mockCurrentInstanceStore) LoadCurrentJIRAInstance() (Instance, error) {
	return &jiraTestInstance{
		JIRAInstance: *NewJIRAInstance(store.plugin, "test", "jiraTestInstanceKey"),
	}, nil
}

type mockUserStore struct{}

func (store mockUserStore) StoreUserInfo(ji Instance, mattermostUserId string, jiraUser JIRAUser) error {
	return nil
}
func (store mockUserStore) LoadJIRAUser(ji Instance, mattermostUserId string) (JIRAUser, error) {
	return JIRAUser{}, nil
}
func (store mockUserStore) LoadMattermostUserId(ji Instance, jiraUserName string) (string, error) {
	return "testMattermostUserId012345", nil
}
func (store mockUserStore) DeleteUserInfo(ji Instance, mattermostUserId string) error {
	return nil
}
