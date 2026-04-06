import { useEffect, useMemo, useState } from 'react';
import { API_URL } from './config';

const P = '#7C5CBF';

type Org = { id: number; name: string; role?: string };
type Project = { id: number; org_id: number; name: string; role?: string };
type OrgMember = { user_id: number; email: string; role: 'owner' | 'admin' | 'member' };
type ProjectMember = { user_id: number; email: string; role: 'pm' | 'dev' | 'qa' | 'viewer' };

function Panel({ title, children }: { title: string; children: any }) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-5">
      <h2 className="text-base font-bold text-gray-900 mb-4">{title}</h2>
      {children}
    </div>
  );
}

export default function AdminPage() {
  const jwtToken = localStorage.getItem('jwtToken') || '';
  const authHeaders = useMemo(() => (jwtToken ? { Authorization: `Bearer ${jwtToken}` } : {}), [jwtToken]);

  const [orgs, setOrgs] = useState<Org[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [orgMembers, setOrgMembers] = useState<OrgMember[]>([]);
  const [projectMembers, setProjectMembers] = useState<ProjectMember[]>([]);
  const [loading, setLoading] = useState(false);

  const [selectedOrgId, setSelectedOrgId] = useState<number>(() => Number(localStorage.getItem('selectedOrgId') || '0'));
  const [selectedProjectId, setSelectedProjectId] = useState<number>(() => Number(localStorage.getItem('selectedProjectId') || '0'));

  const selectedOrg = orgs.find(o => o.id === selectedOrgId);
  const isOrgAdmin = selectedOrg?.role === 'owner' || selectedOrg?.role === 'admin';

  const [newOrgName, setNewOrgName] = useState('');
  const [newProjectName, setNewProjectName] = useState('');
  const [orgMemberEmail, setOrgMemberEmail] = useState('');
  const [orgMemberRole, setOrgMemberRole] = useState<'member' | 'admin'>('member');
  const [orgMemberHint, setOrgMemberHint] = useState<string | null>(null);
  const [projectMemberEmail, setProjectMemberEmail] = useState('');
  const [projectMemberRole, setProjectMemberRole] = useState<'pm' | 'dev' | 'qa' | 'viewer'>('viewer');
  const [projectMemberHint, setProjectMemberHint] = useState<string | null>(null);

  const fetchOrgs = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/orgs`, { headers: authHeaders });
      const data = await res.json().catch(() => []);
      const orgsData: Org[] = Array.isArray(data) ? data : [];
      setOrgs(orgsData);
      const stored = Number(localStorage.getItem('selectedOrgId') || '0');
      const firstId = orgsData[0]?.id || 0;
      const nextOrg = stored && orgsData.some(o => o.id === stored) ? stored : firstId;
      if (nextOrg && nextOrg !== selectedOrgId) {
        setSelectedOrgId(nextOrg);
        localStorage.setItem('selectedOrgId', String(nextOrg));
      }
    } finally { setLoading(false); }
  };

  const fetchProjects = async (orgId: number) => {
    if (!orgId) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/projects?org_id=${orgId}`, { headers: authHeaders });
      const data = await res.json().catch(() => []);
      const list: Project[] = Array.isArray(data) ? data : [];
      setProjects(list);
      const stored = Number(localStorage.getItem('selectedProjectId') || '0');
      const firstId = list[0]?.id || 0;
      const nextProject = stored && list.some(p => p.id === stored) ? stored : firstId;
      if (nextProject && nextProject !== selectedProjectId) {
        setSelectedProjectId(nextProject);
        localStorage.setItem('selectedProjectId', String(nextProject));
      }
    } finally { setLoading(false); }
  };

  const fetchOrgMembers = async (orgId: number) => {
    if (!orgId || !isOrgAdmin) { setOrgMembers([]); return; }
    const res = await fetch(`${API_URL}/orgs/${orgId}/members`, { headers: authHeaders });
    const data = await res.json().catch(() => []);
    setOrgMembers(Array.isArray(data) ? data : []);
  };

  const fetchProjectMembers = async (projectId: number) => {
    if (!projectId || !isOrgAdmin) { setProjectMembers([]); return; }
    const res = await fetch(`${API_URL}/projects/${projectId}/members`, { headers: authHeaders });
    const data = await res.json().catch(() => []);
    setProjectMembers(Array.isArray(data) ? data : []);
  };

  useEffect(() => { fetchOrgs(); }, []);
  useEffect(() => { if (selectedOrgId) fetchProjects(selectedOrgId); }, [selectedOrgId]);
  useEffect(() => { if (selectedOrgId) fetchOrgMembers(selectedOrgId); }, [selectedOrgId, isOrgAdmin]);
  useEffect(() => { if (selectedProjectId) fetchProjectMembers(selectedProjectId); }, [selectedProjectId, isOrgAdmin]);

  const createOrg = async () => {
    const name = newOrgName.trim();
    if (!name) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/orgs`, { method: 'POST', headers: { 'Content-Type': 'application/json', ...authHeaders }, body: JSON.stringify({ name }) });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setNewOrgName('');
      await fetchOrgs();
    } catch (e: any) { alert(e.message); }
    finally { setLoading(false); }
  };

  const createProject = async () => {
    const name = newProjectName.trim();
    if (!name || !selectedOrgId) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/projects`, { method: 'POST', headers: { 'Content-Type': 'application/json', ...authHeaders }, body: JSON.stringify({ org_id: selectedOrgId, name }) });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setNewProjectName('');
      await fetchProjects(selectedOrgId);
    } catch (e: any) { alert(e.message); }
    finally { setLoading(false); }
  };

  const addOrgMember = async () => {
    if (!selectedOrgId) return;
    const email = orgMemberEmail.trim();
    if (!email) return;
    setOrgMemberHint(null);
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/orgs/${selectedOrgId}/members`, { method: 'POST', headers: { 'Content-Type': 'application/json', ...authHeaders }, body: JSON.stringify({ email, role: orgMemberRole }) });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setOrgMemberEmail('');
      await fetchOrgMembers(selectedOrgId);
    } catch (e: any) {
      if (e.message === 'user_not_found') setOrgMemberHint('Пользователь не зарегистрирован.');
      else if (e.message === 'not_allowed') setOrgMemberHint('Недостаточно прав.');
      else alert(e.message);
    } finally { setLoading(false); }
  };

  const addProjectMember = async () => {
    if (!selectedProjectId) return;
    const email = projectMemberEmail.trim();
    if (!email) return;
    setProjectMemberHint(null);
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/projects/${selectedProjectId}/members`, { method: 'POST', headers: { 'Content-Type': 'application/json', ...authHeaders }, body: JSON.stringify({ email, role: projectMemberRole }) });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setProjectMemberEmail('');
      await fetchProjectMembers(selectedProjectId);
    } catch (e: any) {
      if (e.message === 'user_not_found') setProjectMemberHint('Пользователь не зарегистрирован.');
      else if (e.message === 'not_allowed') setProjectMemberHint('Недостаточно прав.');
      else alert(e.message);
    } finally { setLoading(false); }
  };

  const updateOrgMemberRole = async (userId: number, role: OrgMember['role']) => {
    if (!selectedOrgId) return;
    await fetch(`${API_URL}/orgs/${selectedOrgId}/members/${userId}`, { method: 'PATCH', headers: { 'Content-Type': 'application/json', ...authHeaders }, body: JSON.stringify({ role }) });
    await fetchOrgMembers(selectedOrgId);
  };

  const deleteOrgMember = async (userId: number) => {
    if (!selectedOrgId || !confirm('Удалить участника?')) return;
    await fetch(`${API_URL}/orgs/${selectedOrgId}/members/${userId}`, { method: 'DELETE', headers: authHeaders });
    await fetchOrgMembers(selectedOrgId);
  };

  const updateProjectMemberRole = async (userId: number, role: ProjectMember['role']) => {
    if (!selectedProjectId) return;
    await fetch(`${API_URL}/projects/${selectedProjectId}/members/${userId}`, { method: 'PATCH', headers: { 'Content-Type': 'application/json', ...authHeaders }, body: JSON.stringify({ role }) });
    await fetchProjectMembers(selectedProjectId);
  };

  const deleteProjectMember = async (userId: number) => {
    if (!selectedProjectId || !confirm('Удалить участника?')) return;
    await fetch(`${API_URL}/projects/${selectedProjectId}/members/${userId}`, { method: 'DELETE', headers: authHeaders });
    await fetchProjectMembers(selectedProjectId);
  };

  const inputCls = 'flex-1 px-3 py-2.5 rounded-xl border border-gray-200 text-sm outline-none focus:border-[#7C5CBF]';
  const selectCls = 'px-3 py-2.5 rounded-xl border border-gray-200 text-sm outline-none bg-white';

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-gray-900">Управление</h1>
        {loading && <span className="text-sm text-gray-400">Загрузка…</span>}
      </div>

      <div className="flex flex-wrap gap-3 mb-6">
        <select value={selectedOrgId || ''} onChange={e => { const v = Number(e.target.value); setSelectedOrgId(v); setSelectedProjectId(0); localStorage.setItem('selectedOrgId', String(v)); localStorage.removeItem('selectedProjectId'); }} className={selectCls}>
          {orgs.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}
        </select>
        <select value={selectedProjectId || ''} onChange={e => { const v = Number(e.target.value); setSelectedProjectId(v); localStorage.setItem('selectedProjectId', String(v)); }} className={selectCls} disabled={!selectedOrgId}>
          {projects.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
        </select>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Panel title="Создать организацию">
          <div className="flex gap-2">
            <input value={newOrgName} onChange={e => setNewOrgName(e.target.value)} placeholder="Название" className={inputCls} />
            <button onClick={createOrg} className="text-white px-4 py-2.5 rounded-xl font-semibold text-sm" style={{ background: P }}>Создать</button>
          </div>
          <p className="text-xs text-gray-400 mt-2">Вы станете owner и сможете выдавать роли.</p>
        </Panel>

        <Panel title="Создать проект">
          <div className="flex gap-2">
            <input value={newProjectName} onChange={e => setNewProjectName(e.target.value)} placeholder="Название" className={inputCls} disabled={!selectedOrgId} />
            <button onClick={createProject} disabled={!selectedOrgId} className="bg-gray-900 text-white px-4 py-2.5 rounded-xl font-semibold text-sm disabled:opacity-50">Создать</button>
          </div>
          {!selectedOrgId && <p className="text-xs text-gray-400 mt-2">Сначала выберите организацию.</p>}
        </Panel>

        <Panel title="Добавить участника в организацию">
          {!isOrgAdmin ? (
            <p className="text-sm text-gray-400">Нужно быть owner/admin организации.</p>
          ) : (
            <div className="space-y-2">
              <input value={orgMemberEmail} onChange={e => setOrgMemberEmail(e.target.value)} placeholder="email" className={inputCls + ' w-full'} />
              <div className="flex gap-2">
                <select value={orgMemberRole} onChange={e => setOrgMemberRole(e.target.value as any)} className={selectCls + ' flex-1'}>
                  <option value="member">member</option><option value="admin">admin</option>
                </select>
                <button onClick={addOrgMember} className="text-white px-4 py-2.5 rounded-xl font-semibold text-sm" style={{ background: '#10B981' }}>Добавить</button>
              </div>
              <div className="flex items-center justify-between">
                <p className="text-xs text-gray-400">Пользователь должен быть зарегистрирован.</p>
                <button onClick={async () => { await navigator.clipboard.writeText(`${window.location.origin}/login`); alert('Ссылка скопирована'); }} className="text-xs font-bold" style={{ color: P }}>Скопировать ссылку</button>
              </div>
              {orgMemberHint && <p className="text-xs text-orange-600 bg-orange-50 border border-orange-100 rounded-xl p-2">{orgMemberHint}</p>}
            </div>
          )}
        </Panel>

        <Panel title="Участники организации">
          {!isOrgAdmin ? (
            <p className="text-sm text-gray-400">Нужно быть owner/admin организации.</p>
          ) : (
            <div className="space-y-2">
              {orgMembers.map(m => (
                <div key={m.user_id} className="flex items-center gap-2 p-2 rounded-xl border border-gray-100">
                  <p className="text-sm text-gray-700 truncate flex-1">{m.email}</p>
                  <select value={m.role} onChange={e => updateOrgMemberRole(m.user_id, e.target.value as OrgMember['role'])} className="px-2 py-1.5 rounded-lg border border-gray-200 text-xs">
                    <option value="owner">owner</option><option value="admin">admin</option><option value="member">member</option>
                  </select>
                  <button onClick={() => deleteOrgMember(m.user_id)} className="text-xs font-bold text-red-500 hover:text-red-700">Удалить</button>
                </div>
              ))}
              {orgMembers.length === 0 && <p className="text-xs text-gray-400">Нет участников</p>}
            </div>
          )}
        </Panel>

        <Panel title="Добавить участника в проект">
          {!isOrgAdmin ? (
            <p className="text-sm text-gray-400">Нужно быть owner/admin организации.</p>
          ) : (
            <div className="space-y-2">
              <input value={projectMemberEmail} onChange={e => setProjectMemberEmail(e.target.value)} placeholder="email" className={inputCls + ' w-full'} disabled={!selectedProjectId} />
              <div className="flex gap-2">
                <select value={projectMemberRole} onChange={e => setProjectMemberRole(e.target.value as any)} className={selectCls + ' flex-1'} disabled={!selectedProjectId}>
                  <option value="pm">pm</option><option value="dev">dev</option><option value="qa">qa</option><option value="viewer">viewer</option>
                </select>
                <button onClick={addProjectMember} disabled={!selectedProjectId} className="text-white px-4 py-2.5 rounded-xl font-semibold text-sm disabled:opacity-50" style={{ background: '#3B82F6' }}>Добавить</button>
              </div>
              {projectMemberHint && <p className="text-xs text-orange-600 bg-orange-50 border border-orange-100 rounded-xl p-2">{projectMemberHint}</p>}
            </div>
          )}
        </Panel>

        <Panel title="Участники проекта">
          {!isOrgAdmin ? (
            <p className="text-sm text-gray-400">Нужно быть owner/admin организации.</p>
          ) : (
            <div className="space-y-2">
              {projectMembers.map(m => (
                <div key={m.user_id} className="flex items-center gap-2 p-2 rounded-xl border border-gray-100">
                  <p className="text-sm text-gray-700 truncate flex-1">{m.email}</p>
                  <select value={m.role} onChange={e => updateProjectMemberRole(m.user_id, e.target.value as ProjectMember['role'])} className="px-2 py-1.5 rounded-lg border border-gray-200 text-xs">
                    <option value="pm">pm</option><option value="dev">dev</option><option value="qa">qa</option><option value="viewer">viewer</option>
                  </select>
                  <button onClick={() => deleteProjectMember(m.user_id)} className="text-xs font-bold text-red-500 hover:text-red-700">Удалить</button>
                </div>
              ))}
              {projectMembers.length === 0 && <p className="text-xs text-gray-400">Нет участников</p>}
            </div>
          )}
        </Panel>
      </div>
    </div>
  );
}
