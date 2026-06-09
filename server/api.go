package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/")
	userID := p.userIDFromRequest(c, r)
	switch {
	case r.Method == http.MethodPost && path == "api/v1/forward":
		p.handleForward(w, r, userID)
	case r.Method == http.MethodGet && path == "api/v1/targets":
		p.handleTargets(w, r, userID)
	default:
		http.NotFound(w, r)
	}
}

type apiError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, errCode, msg string) {
	writeJSON(w, code, apiError{false, errCode, msg})
}

func (p *Plugin) userIDFromRequest(c *plugin.Context, r *http.Request) string {
	if userID := r.Header.Get("Mattermost-User-Id"); userID != "" {
		return userID
	}
	if userID := r.Header.Get("Mattermost-User-ID"); userID != "" {
		return userID
	}
	if c != nil && c.SessionId != "" {
		if session, appErr := p.API.GetSession(c.SessionId); appErr == nil && session != nil {
			return session.UserId
		}
	}
	return ""
}
