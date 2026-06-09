package main
import "fmt"
type Configuration struct { EnablePlugin bool; AllowForwardFromPublicChannels bool; AllowForwardFromPrivateChannels bool; AllowForwardFromDirectMessages bool; AllowFileForwarding bool; MaxFilesPerForward int; MaxFileSizeMB int }
func defaultConfiguration()*Configuration{ return &Configuration{true,true,true,true,true,10,100} }
func (p *Plugin) OnConfigurationChange() error { c:=defaultConfiguration(); if err:=p.API.LoadPluginConfiguration(c); err!=nil {return err}; if c.MaxFilesPerForward<=0 {c.MaxFilesPerForward=10}; if c.MaxFileSizeMB<0 {return fmt.Errorf("MaxFileSizeMB cannot be negative")}; p.configurationLock.Lock(); p.configuration=c; p.configurationLock.Unlock(); return nil }
func (p *Plugin) getConfiguration()*Configuration{ p.configurationLock.RLock(); defer p.configurationLock.RUnlock(); if p.configuration==nil {return defaultConfiguration()}; cp:=*p.configuration; return &cp }
