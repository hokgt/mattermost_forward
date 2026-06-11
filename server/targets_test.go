package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/mock"
)

func TestChannelMatchesForwardTargetDisplayName(t *testing.T) {
	ch := &model.Channel{Id: "channel1", Name: "ops-alerts", DisplayName: "Operations Alerts", Type: model.ChannelTypeOpen}
	if !channelMatchesForwardTarget(ch, "operations") {
		t.Fatal("expected channel display name to match search term")
	}
	if !channelMatchesForwardTarget(ch, "ops") {
		t.Fatal("expected channel name to match search term")
	}
	if channelMatchesForwardTarget(&model.Channel{Id: "dm1", Name: "dm", Type: model.ChannelTypeDirect}, "dm") {
		t.Fatal("direct channels should not be returned as channel targets")
	}
}

func TestHandleTargetsFindsJoinedChannelsWhenTeamIDMissing(t *testing.T) {
	api := &plugintest.API{}
	api.On("GetChannelsForTeamForUser", "", "user1", false).Return([]*model.Channel{
		{Id: "channel1", Name: "ops-alerts", DisplayName: "Operations Alerts", Type: model.ChannelTypeOpen},
	}, (*model.AppError)(nil)).Once()
	api.On("HasPermissionToChannel", "user1", "channel1", model.PermissionCreatePost).Return(true).Once()
	api.On("SearchUsers", mock.MatchedBy(func(search *model.UserSearch) bool {
		return search != nil && search.Term == "operations" && search.Limit == 20 && !search.AllowInactive
	})).Return([]*model.User{}, (*model.AppError)(nil)).Once()

	p := &Plugin{}
	p.API = api

	req := httptest.NewRequest("GET", "/api/v1/targets?q=operations", nil)
	rec := httptest.NewRecorder()
	p.handleTargets(rec, req, "user1")

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Success bool           `json:"success"`
		Targets []TargetOption `json:"targets"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || len(body.Targets) != 1 {
		t.Fatalf("expected one channel target, got %+v", body)
	}
	if got := body.Targets[0]; got.Type != "channel" || got.ID != "channel1" || !strings.Contains(got.DisplayName, "Operations Alerts") {
		t.Fatalf("unexpected target: %+v", got)
	}
	api.AssertExpectations(t)
}
