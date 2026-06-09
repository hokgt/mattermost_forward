package main
import ("encoding/json"; "fmt")
func (p *Plugin) saveAuditRecord(r *AuditRecord) error { b,err:=json.Marshal(r); if err!=nil{return err}; if appErr:=p.API.KVSet(fmt.Sprintf("audit:%s",r.ID),b); appErr!=nil{return appErr}; return nil }
