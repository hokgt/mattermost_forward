package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

func displayUser(u *model.User) string {
	if u == nil {
		return "unknown"
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	return u.Id
}

func displayChannel(c *model.Channel) string {
	if c == nil {
		return "unknown"
	}
	if c.DisplayName != "" {
		return c.DisplayName
	}
	if c.Name != "" {
		return c.Name
	}
	return c.Id
}

func quoteMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	lines := strings.Split(msg, "\n")
	for i, l := range lines {
		lines[i] = "> " + l
	}
	return strings.Join(lines, "\n")
}

func buildForwardedMessage(original *model.Post, source *model.Channel, originalUser, forwardUser *model.User, note string, includeText bool) string {
	originalTime := time.UnixMilli(original.CreateAt).Format("2006-01-02 15:04")
	parts := []string{
		fmt.Sprintf("**Original sender:** %s", displayUser(originalUser)),
		fmt.Sprintf("**Original time:** %s", originalTime),
	}
	note = strings.TrimSpace(note)
	if note != "" {
		parts = append(parts, "", "**Note:**", note)
	}
	if includeText {
		if q := quoteMessage(original.Message); q != "" {
			parts = append(parts, "", "---", q)
		}
	}
	return strings.Join(parts, "\n")
}
