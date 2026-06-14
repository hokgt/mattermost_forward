package main

type ForwardTargetRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ForwardRequest struct {
	PostID       string                 `json:"post_id"`
	TargetType   string                 `json:"target_type"`
	TargetID     string                 `json:"target_id"`
	Targets      []ForwardTargetRequest `json:"targets"`
	TargetTeamID string                 `json:"team_id"`
	IncludeText  bool                   `json:"include_text"`
	IncludeFiles bool                   `json:"include_files"`
	Note         string                 `json:"note"`
}
type ForwardResponse struct {
	Success          bool     `json:"success"`
	NewPostID        string   `json:"new_post_id,omitempty"`
	TargetChannelID  string   `json:"target_channel_id,omitempty"`
	NewPostIDs       []string `json:"new_post_ids,omitempty"`
	TargetChannelIDs []string `json:"target_channel_ids,omitempty"`
	ForwardedCount   int      `json:"forwarded_count"`
}
type AuditRecord struct {
	ID                string `json:"id"`
	OriginalPostID    string `json:"original_post_id"`
	SourceChannelID   string `json:"source_channel_id"`
	SourceChannelType string `json:"source_channel_type"`
	OriginalUserID    string `json:"original_user_id"`
	ForwardedByUserID string `json:"forwarded_by_user_id"`
	TargetChannelID   string `json:"target_channel_id"`
	TargetType        string `json:"target_type"`
	NewPostID         string `json:"new_post_id"`
	IncludedText      bool   `json:"included_text"`
	IncludedFiles     bool   `json:"included_files"`
	FileCount         int    `json:"file_count"`
	CreatedAt         int64  `json:"created_at"`
}
type TargetOption struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}
