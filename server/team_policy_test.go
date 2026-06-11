package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
)

func activeTeamMember(teamID, userID string) *model.TeamMember {
	return &model.TeamMember{TeamId: teamID, UserId: userID, DeleteAt: 0}
}

func TestUsersShareTargetTeamRequiresBothUsersInCurrentTeam(t *testing.T) {
	api := &plugintest.API{}
	api.On("GetTeamMember", "team1", "user1").Return(activeTeamMember("team1", "user1"), (*model.AppError)(nil)).Once()
	api.On("GetTeamMember", "team1", "user2").Return(activeTeamMember("team1", "user2"), (*model.AppError)(nil)).Once()

	p := &Plugin{}
	p.API = api
	if !p.usersShareTargetTeam("user1", "user2", "team1") {
		t.Fatal("expected users in the same current team to be allowed")
	}
	api.AssertExpectations(t)
}

func TestUsersShareTargetTeamRejectsUserOutsideCurrentTeam(t *testing.T) {
	api := &plugintest.API{}
	api.On("GetTeamMember", "team1", "user1").Return(activeTeamMember("team1", "user1"), (*model.AppError)(nil)).Once()
	api.On("GetTeamMember", "team1", "user2").Return((*model.TeamMember)(nil), model.NewAppError("test", "not_found", nil, "", 404)).Once()

	p := &Plugin{}
	p.API = api
	if p.usersShareTargetTeam("user1", "user2", "team1") {
		t.Fatal("expected user outside current team to be rejected")
	}
	api.AssertExpectations(t)
}

func TestUsersShareTargetTeamFallsBackToAnySharedTeam(t *testing.T) {
	api := &plugintest.API{}
	api.On("GetTeamMembersForUser", "user1", 0, 200).Return([]*model.TeamMember{activeTeamMember("team1", "user1")}, (*model.AppError)(nil)).Once()
	api.On("GetTeamMembersForUser", "user2", 0, 200).Return([]*model.TeamMember{activeTeamMember("team2", "user2"), activeTeamMember("team1", "user2")}, (*model.AppError)(nil)).Once()

	p := &Plugin{}
	p.API = api
	if !p.usersShareTargetTeam("user1", "user2", "") {
		t.Fatal("expected users with any shared active team to be allowed")
	}
	api.AssertExpectations(t)
}
