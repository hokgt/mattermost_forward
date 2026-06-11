package main

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

func (p *Plugin) handleTargets(w http.ResponseWriter, r *http.Request, userID string) {
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "not_authenticated", "You must be logged in.")
		return
	}
	term := strings.TrimSpace(r.URL.Query().Get("q"))
	teamID := strings.TrimSpace(r.URL.Query().Get("team_id"))
	targets := []TargetOption{}
	seen := map[string]bool{}
	if len(term) >= 1 {
		if teamID != "" {
			if channels, err := p.API.SearchChannels(teamID, term); err == nil {
				targets = p.appendChannelTargets(targets, userID, term, channels, seen)
			}
		}

		// Mattermost's webapp can provide an empty currentTeamId from DM/GM contexts,
		// and SearchChannels can miss private/joined channels depending on server search
		// behavior. Always supplement from the user's joined channels; an empty teamID
		// asks Mattermost for channels across all of the user's teams.
		if channels, err := p.API.GetChannelsForTeamForUser(teamID, userID, false); err == nil {
			targets = p.appendChannelTargets(targets, userID, term, channels, seen)
		}

		if users, err := p.API.SearchUsers(&model.UserSearch{Term: term, Limit: 20, AllowInactive: false}); err == nil {
			for _, u := range users {
				if u == nil || u.Id == userID || u.DeleteAt != 0 {
					continue
				}
				targets = append(targets, TargetOption{"user", u.Id, u.Username, "@" + u.Username})
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "targets": targets})
}

func (p *Plugin) appendChannelTargets(targets []TargetOption, userID, term string, channels []*model.Channel, seen map[string]bool) []TargetOption {
	for _, ch := range channels {
		if !channelMatchesForwardTarget(ch, term) || seen[ch.Id] {
			continue
		}
		if p.canPostTarget(userID, ch) {
			seen[ch.Id] = true
			targets = append(targets, TargetOption{"channel", ch.Id, ch.Name, channelForwardLabel(ch)})
		}
	}
	return targets
}

func channelMatchesForwardTarget(ch *model.Channel, term string) bool {
	if ch == nil || ch.DeleteAt != 0 || (ch.Type != model.ChannelTypeOpen && ch.Type != model.ChannelTypePrivate) {
		return false
	}
	needle := strings.ToLower(strings.TrimSpace(term))
	if needle == "" {
		return false
	}
	return strings.Contains(strings.ToLower(ch.Name), needle) || strings.Contains(strings.ToLower(ch.DisplayName), needle)
}

func channelForwardLabel(ch *model.Channel) string {
	label := "#" + ch.Name
	if strings.TrimSpace(ch.DisplayName) != "" {
		label = "#" + ch.DisplayName
	}
	return label
}
