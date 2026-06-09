package main

import (
    "fmt"
    "sync"
    "github.com/mattermost/mattermost/server/public/model"
    "github.com/mattermost/mattermost/server/public/plugin"
)
const pluginID = "com.wijayacorp.message-forward"
type Plugin struct { plugin.MattermostPlugin; configurationLock sync.RWMutex; configuration *Configuration }
func (p *Plugin) OnActivate() error { if err:=p.OnConfigurationChange(); err!=nil { return fmt.Errorf("failed to load configuration: %w", err)}; if _,err:=p.API.EnsureBotUser(&model.Bot{Username:"messageforward",DisplayName:"Message Forward",Description:"Bot account for Wijaya Message Forward plugin"}); err!=nil { p.API.LogWarn("Failed to ensure plugin bot user", "error", err.Error())}; p.API.LogInfo("Wijaya Message Forward activated"); return nil }
func (p *Plugin) OnDeactivate() error { p.API.LogInfo("Wijaya Message Forward deactivated"); return nil }
func main(){ plugin.ClientMain(&Plugin{}) }
