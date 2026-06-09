package main
import ("encoding/json"; "net/http"; "strings"; "github.com/mattermost/mattermost/server/public/plugin")
func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); path:=strings.TrimPrefix(r.URL.Path,"/"); switch {case r.Method==http.MethodPost && path=="api/v1/forward": p.handleForward(w,r); case r.Method==http.MethodGet && path=="api/v1/targets": p.handleTargets(w,r); default: http.NotFound(w,r)} }
type apiError struct{ Success bool `json:"success"`; Error string `json:"error"`; Message string `json:"message"` }
func writeJSON(w http.ResponseWriter, code int, v interface{}){ w.WriteHeader(code); _=json.NewEncoder(w).Encode(v) }
func writeError(w http.ResponseWriter, code int, errCode,msg string){ writeJSON(w,code,apiError{false,errCode,msg}) }
func userIDFromRequest(r *http.Request) string { return r.Header.Get("Mattermost-User-Id") }
