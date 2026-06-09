import {ForwardRequest, TargetOption} from '../types';

const pluginId = 'com.wijayacorp.message-forward';

function cookieValue(name: string): string {
    const prefix = `${name}=`;
    const cookie = document.cookie.split('; ').find((entry) => entry.startsWith(prefix));
    return cookie ? decodeURIComponent(cookie.substring(prefix.length)) : '';
}

function csrf() {
    const token = cookieValue('MMCSRF') || (window as any).MMCSRF || '';
    return token ? {'X-CSRF-Token': token} : {};
}

async function request(path: string, options: RequestInit = {}) {
    const res = await fetch(`/plugins/${pluginId}${path}`, {
        credentials: 'include',
        ...options,
        headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest',
            ...csrf(),
            ...(options.headers || {}),
        },
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok || data.success === false) {
        throw new Error(data.message || 'Request failed');
    }
    return data;
}

export async function searchTargets(q: string, teamId: string): Promise<TargetOption[]> {
    if (!q.trim()) {
        return [];
    }
    const p = new URLSearchParams({q: q.trim()});
    if (teamId) {
        p.set('team_id', teamId);
    }
    const d = await request(`/api/v1/targets?${p.toString()}`, {method: 'GET'});
    return d.targets || [];
}

export async function forwardMessage(req: ForwardRequest) {
    return request('/api/v1/forward', {method: 'POST', body: JSON.stringify(req)});
}
