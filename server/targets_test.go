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
	api.On("GetChannelMember", "channel1", "user1").Return(&model.ChannelMember{ChannelId: "channel1", UserId: "user1"}, (*model.AppError)(nil)).Once()
	api.On("HasPermissionToChannel", "user1", "channel1", model.PermissionCreatePost).Return(true).Once()
	api.On("SearchUsers", mock.MatchedBy(func(search *model.UserSearch) bool {
		return search != nil && search.Term == "operations" && search.Limit == 50 && !search.AllowInactive
	})).Return([]*model.User{}, (*model.AppError)(nil)).Once()
	api.On("GetUsers", mock.Anything).Return([]*model.User{}, (*model.AppError)(nil)).Once()

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

func TestDisambiguateDuplicateChannelLabels(t *testing.T) {
	targets := []TargetOption{
		{Type: "channel", ID: "channel1", Name: "frappe-auto-packing-list", DisplayName: "#Autofetch Packing List"},
		{Type: "channel", ID: "channel2", Name: "frappe-autofetch", DisplayName: "#Autofetch Packing List"},
		{Type: "user", ID: "user2", Name: "auto", DisplayName: "@auto"},
	}
	disambiguateDuplicateChannelLabels(targets)
	if targets[0].DisplayName != "#Autofetch Packing List (frappe-auto-packing-list)" {
		t.Fatalf("unexpected first label: %q", targets[0].DisplayName)
	}
	if targets[1].DisplayName != "#Autofetch Packing List (frappe-autofetch)" {
		t.Fatalf("unexpected second label: %q", targets[1].DisplayName)
	}
	if targets[2].DisplayName != "@auto" {
		t.Fatalf("user target should not change: %q", targets[2].DisplayName)
	}
}

func TestHandleTargetsSkipsUnjoinedSearchChannels(t *testing.T) {
	api := &plugintest.API{}
	teamID := "team1"
	api.On("SearchChannels", teamID, "auto").Return([]*model.Channel{
		{Id: "joined", Name: "frappe-auto-packing-list", DisplayName: "Autofetch Packing List", Type: model.ChannelTypeOpen},
		{Id: "unjoined", Name: "frappe-autofetch", DisplayName: "Autofetch Packing List", Type: model.ChannelTypeOpen},
	}, (*model.AppError)(nil)).Once()
	api.On("GetChannelMember", "joined", "user1").Return(&model.ChannelMember{ChannelId: "joined", UserId: "user1"}, (*model.AppError)(nil)).Once()
	api.On("HasPermissionToChannel", "user1", "joined", model.PermissionCreatePost).Return(true).Once()
	api.On("GetChannelMember", "unjoined", "user1").Return((*model.ChannelMember)(nil), model.NewAppError("test", "not_found", nil, "", 404)).Once()
	api.On("GetChannelsForTeamForUser", "", "user1", false).Return([]*model.Channel{
		{Id: "joined", Name: "frappe-auto-packing-list", DisplayName: "Autofetch Packing List", Type: model.ChannelTypeOpen},
	}, (*model.AppError)(nil)).Once()
	api.On("SearchUsers", mock.Anything).Return([]*model.User{}, (*model.AppError)(nil)).Once()
	api.On("GetUsers", mock.Anything).Return([]*model.User{}, (*model.AppError)(nil)).Once()

	p := &Plugin{}
	p.API = api

	req := httptest.NewRequest("GET", "/api/v1/targets?q=auto&team_id="+teamID, nil)
	rec := httptest.NewRecorder()
	p.handleTargets(rec, req, "user1")

	var body struct {
		Success bool           `json:"success"`
		Targets []TargetOption `json:"targets"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Targets) != 1 || body.Targets[0].ID != "joined" {
		t.Fatalf("expected only joined channel target, got %+v", body.Targets)
	}
	api.AssertExpectations(t)
}

func TestUserMatchesForwardSearchFullName(t *testing.T) {
	u := &model.User{Username: "hok", FirstName: "Suhendri", LastName: "Wijaya"}
	for _, term := range []string{"@hok", "hok", "suhendri", "wijaya", "Suhendri Wijaya"} {
		if !userMatchesForwardSearch(u, term) {
			t.Fatalf("expected %q to match user", term)
		}
	}
	if userForwardLabel(u) != "@hok — Suhendri Wijaya" {
		t.Fatalf("unexpected label: %q", userForwardLabel(u))
	}
}

func TestAppendUserTargetsIncludesFullNameMatchesAcrossTeams(t *testing.T) {
	api := &plugintest.API{}
	api.On("SearchUsers", mock.MatchedBy(func(search *model.UserSearch) bool {
		return search != nil && search.Term == "Suhendri" && search.Limit == 50 && !search.AllowInactive
	})).Return([]*model.User{}, (*model.AppError)(nil)).Once()
	api.On("GetUsers", mock.MatchedBy(func(options *model.UserGetOptions) bool {
		return options != nil && options.Active && options.Page == 0 && options.PerPage == 200
	})).Return([]*model.User{
		{Id: "target", Username: "hok", FirstName: "Suhendri", LastName: "Wijaya"},
	}, (*model.AppError)(nil)).Once()
	p := &Plugin{}
	p.API = api
	targets := p.appendUserTargets(nil, "user1", "team1", "Suhendri")
	if len(targets) != 1 || targets[0].ID != "target" || targets[0].DisplayName != "@hok — Suhendri Wijaya" {
		t.Fatalf("unexpected targets: %+v", targets)
	}
	api.AssertExpectations(t)
}
