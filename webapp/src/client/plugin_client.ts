import {ForwardRequest,TargetOption} from '../types';
const pluginId='com.wijayacorp.message-forward';
function csrf(){const t=(window as any).MMCSRF||'';return t?{'X-CSRF-Token':t}:{}}
async function request(path:string, options:RequestInit={}){const res=await fetch(`/plugins/${pluginId}${path}`,{credentials:'same-origin',...options,headers:{'Content-Type':'application/json',...csrf(),...(options.headers||{})}});const data=await res.json().catch(()=>({}));if(!res.ok||data.success===false)throw new Error(data.message||'Request failed');return data}
export async function searchTargets(q:string, teamId:string):Promise<TargetOption[]>{if(!q.trim())return[];const p=new URLSearchParams({q:q.trim()});if(teamId)p.set('team_id',teamId);const d=await request(`/api/v1/targets?${p.toString()}`,{method:'GET'});return d.targets||[]}
export async function forwardMessage(req:ForwardRequest){return request('/api/v1/forward',{method:'POST',body:JSON.stringify(req)})}
