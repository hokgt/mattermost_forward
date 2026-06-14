package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost/server/public/model"
	"net/http"
	"strings"
)

const maxForwardTargets = 10

type resolvedForwardTarget struct {
	Channel *model.Channel
	Kind    string
}

func (p *Plugin) handleForward(w http.ResponseWriter, r *http.Request, userID string) {
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "not_authenticated", "You must be logged in.")
		return
	}
	c := p.getConfiguration()
	if !c.EnablePlugin {
		writeError(w, http.StatusForbidden, "plugin_disabled", "Message forwarding is disabled.")
		return
	}
	var req ForwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body.")
		return
	}
	req.TargetTeamID = strings.TrimSpace(req.TargetTeamID)
	req.Note = strings.TrimSpace(req.Note)
	targetReqs := normalizeForwardTargets(req)
	if req.PostID == "" || len(targetReqs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "Post and at least one target are required.")
		return
	}
	if len(targetReqs) > maxForwardTargets {
		writeError(w, http.StatusBadRequest, "too_many_targets", fmt.Sprintf("Select up to %d users or channels.", maxForwardTargets))
		return
	}
	if !req.IncludeText && !req.IncludeFiles {
		writeError(w, http.StatusBadRequest, "nothing_to_forward", "Select message text, files, or both.")
		return
	}
	post, appErr := p.API.GetPost(req.PostID)
	if appErr != nil || post == nil || post.DeleteAt != 0 {
		writeError(w, http.StatusNotFound, "source_post_not_found", "Source post was not found or cannot be forwarded.")
		return
	}
	if strings.HasPrefix(post.Type, "system_") {
		writeError(w, http.StatusBadRequest, "source_post_not_forwardable", "System messages cannot be forwarded.")
		return
	}
	source, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil || source == nil || source.DeleteAt != 0 {
		writeError(w, http.StatusNotFound, "source_channel_not_found", "Source channel was not found.")
		return
	}
	if !sourceTypeAllowed(c, source.Type) {
		writeError(w, http.StatusForbidden, "source_type_disabled", "Forwarding from this conversation type is disabled.")
		return
	}
	if !p.canReadSource(userID, source) {
		writeError(w, http.StatusForbidden, "permission_denied", "You do not have permission to read the source message.")
		return
	}
	resolved := make([]resolvedForwardTarget, 0, len(targetReqs))
	seenChannels := map[string]bool{}
	for _, targetReq := range targetReqs {
		target, targetKind, msg := p.resolveTarget(targetReq.Type, targetReq.ID, userID, req.TargetTeamID)
		if msg != "" {
			writeError(w, http.StatusBadRequest, "target_not_found", msg)
			return
		}
		if target == nil || seenChannels[target.Id] {
			continue
		}
		if !p.canPostTarget(userID, target) {
			writeError(w, http.StatusForbidden, "permission_denied", "You do not have permission to post in one of the selected targets.")
			return
		}
		seenChannels[target.Id] = true
		resolved = append(resolved, resolvedForwardTarget{Channel: target, Kind: targetKind})
	}
	if len(resolved) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "Select at least one unique target.")
		return
	}
	if req.IncludeFiles {
		if !c.AllowFileForwarding {
			writeError(w, http.StatusForbidden, "file_forwarding_disabled", "File forwarding is disabled.")
			return
		}
		if len(post.FileIds) == 0 {
			req.IncludeFiles = false
		} else {
			if len(post.FileIds) > c.MaxFilesPerForward {
				writeError(w, http.StatusBadRequest, "too_many_files", fmt.Sprintf("This message has more than %d files.", c.MaxFilesPerForward))
				return
			}
			if c.MaxFileSizeMB > 0 {
				max := int64(c.MaxFileSizeMB) * 1024 * 1024
				for _, fid := range post.FileIds {
					info, appErr := p.API.GetFileInfo(fid)
					if appErr != nil || info == nil {
						writeError(w, http.StatusBadRequest, "file_not_found", "One of the attached files could not be read.")
						return
					}
					if info.Size > max {
						writeError(w, http.StatusBadRequest, "file_too_large", fmt.Sprintf("File %s is larger than the configured limit.", info.Name))
						return
					}
				}
			}
		}
	}
	originalUser, _ := p.API.GetUser(post.UserId)
	forwardUser, _ := p.API.GetUser(userID)
	newPostIDs := make([]string, 0, len(resolved))
	targetChannelIDs := make([]string, 0, len(resolved))
	for _, item := range resolved {
		copied := []string{}
		if req.IncludeFiles && len(post.FileIds) > 0 {
			ids, appErr := p.API.CopyFileInfos(userID, post.FileIds)
			if appErr != nil {
				writeError(w, http.StatusInternalServerError, "file_copy_failed", "Could not copy attached files.")
				return
			}
			copied = ids
		}
		newPost, appErr := p.API.CreatePost(&model.Post{UserId: userID, ChannelId: item.Channel.Id, Message: buildForwardedMessage(post, source, originalUser, forwardUser, req.Note, req.IncludeText), FileIds: copied})
		if appErr != nil || newPost == nil {
			writeError(w, http.StatusInternalServerError, "create_post_failed", "Could not create the forwarded post.")
			return
		}
		newPostIDs = append(newPostIDs, newPost.Id)
		targetChannelIDs = append(targetChannelIDs, item.Channel.Id)
		rec := &AuditRecord{ID: model.NewId(), OriginalPostID: post.Id, SourceChannelID: source.Id, SourceChannelType: string(source.Type), OriginalUserID: post.UserId, ForwardedByUserID: userID, TargetChannelID: item.Channel.Id, TargetType: item.Kind, NewPostID: newPost.Id, IncludedText: req.IncludeText, IncludedFiles: len(copied) > 0, FileCount: len(copied), CreatedAt: model.GetMillis()}
		if err := p.saveAuditRecord(rec); err != nil {
			p.API.LogWarn("Failed to save forward audit record", "error", err.Error())
		}
	}
	resp := ForwardResponse{Success: true, NewPostIDs: newPostIDs, TargetChannelIDs: targetChannelIDs, ForwardedCount: len(newPostIDs)}
	if len(newPostIDs) == 1 {
		resp.NewPostID = newPostIDs[0]
		resp.TargetChannelID = targetChannelIDs[0]
	}
	writeJSON(w, http.StatusOK, resp)
}

func normalizeForwardTargets(req ForwardRequest) []ForwardTargetRequest {
	items := req.Targets
	if len(items) == 0 && req.TargetID != "" {
		items = []ForwardTargetRequest{{Type: req.TargetType, ID: req.TargetID}}
	}
	out := make([]ForwardTargetRequest, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		t := strings.ToLower(strings.TrimSpace(item.Type))
		id := strings.TrimSpace(item.ID)
		key := t + ":" + id
		if id == "" || (t != "channel" && t != "user") || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ForwardTargetRequest{Type: t, ID: id})
	}
	return out
}

func (p *Plugin) resolveTarget(targetType, targetID, userID, targetTeamID string) (*model.Channel, string, string) {
	if targetType == "channel" {
		ch, appErr := p.API.GetChannel(targetID)
		if appErr != nil || ch == nil || ch.DeleteAt != 0 {
			return nil, "", "Target channel was not found."
		}
		if ch.Type != model.ChannelTypeOpen && ch.Type != model.ChannelTypePrivate {
			return nil, "", "Target channel type is not supported."
		}
		return ch, "channel", ""
	}
	if targetID == userID {
		return nil, "", "Choose another user as the DM target."
	}
	if u, appErr := p.API.GetUser(targetID); appErr != nil || u == nil || u.DeleteAt != 0 {
		return nil, "", "Target user was not found."
	}
	if !p.usersShareTargetTeam(userID, targetID, targetTeamID) {
		return nil, "", "Target user is not in a team/group you are allowed to forward to."
	}
	ch, appErr := p.API.GetDirectChannel(userID, targetID)
	if appErr != nil || ch == nil {
		return nil, "", "Could not create or access the target DM."
	}
	return ch, "dm", ""
}
