package main

func (p *Plugin) usersShareTargetTeam(userID, targetUserID, teamID string) bool {
	if userID == "" || targetUserID == "" || userID == targetUserID {
		return false
	}
	if teamID != "" {
		if !p.isActiveTeamMember(teamID, userID) {
			return false
		}
		return p.isActiveTeamMember(teamID, targetUserID)
	}
	return p.usersShareAnyActiveTeam(userID, targetUserID)
}

func (p *Plugin) isActiveTeamMember(teamID, userID string) bool {
	member, appErr := p.API.GetTeamMember(teamID, userID)
	return appErr == nil && member != nil && member.DeleteAt == 0
}

func (p *Plugin) usersShareAnyActiveTeam(userID, targetUserID string) bool {
	userTeams, appErr := p.API.GetTeamMembersForUser(userID, 0, 200)
	if appErr != nil {
		return false
	}
	targetTeams, appErr := p.API.GetTeamMembersForUser(targetUserID, 0, 200)
	if appErr != nil {
		return false
	}
	active := map[string]bool{}
	for _, member := range userTeams {
		if member != nil && member.DeleteAt == 0 {
			active[member.TeamId] = true
		}
	}
	for _, member := range targetTeams {
		if member != nil && member.DeleteAt == 0 && active[member.TeamId] {
			return true
		}
	}
	return false
}
