package main

import "github.com/mattermost/mattermost/server/public/model"

func sourceTypeAllowed(c *Configuration, t model.ChannelType) bool {
	switch t {
	case model.ChannelTypeOpen:
		return c.AllowForwardFromPublicChannels
	case model.ChannelTypePrivate:
		return c.AllowForwardFromPrivateChannels
	case model.ChannelTypeDirect, model.ChannelTypeGroup:
		return c.AllowForwardFromDirectMessages
	default:
		return false
	}
}
func (p *Plugin) canReadSource(userID string, ch *model.Channel) bool {
	if ch == nil || userID == "" {
		return false
	}
	if !p.API.HasPermissionToChannel(userID, ch.Id, model.PermissionReadChannel) && !p.API.HasPermissionToChannel(userID, ch.Id, model.PermissionReadChannelContent) {
		return false
	}
	if ch.Type == model.ChannelTypePrivate || ch.Type == model.ChannelTypeDirect || ch.Type == model.ChannelTypeGroup {
		if _, err := p.API.GetChannelMember(ch.Id, userID); err != nil {
			return false
		}
	}
	return true
}
func (p *Plugin) canPostTarget(userID string, ch *model.Channel) bool {
	if ch == nil || userID == "" {
		return false
	}
	if _, err := p.API.GetChannelMember(ch.Id, userID); err != nil {
		return false
	}
	return p.API.HasPermissionToChannel(userID, ch.Id, model.PermissionCreatePost)
}
