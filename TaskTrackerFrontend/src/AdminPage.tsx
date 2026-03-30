import { useEffect, useMemo, useState } from 'react';
import { API_URL } from './config';

type Org = { id: number; name: string; role?: string };
type Project = { id: number; org_id: number; name: string; role?: string };
type OrgMember = { user_id: number; email: string; role: 'owner' | 'admin' | 'member' };
type ProjectMember = { user_id: number; email: string; role: 'pm' | 'dev' | 'qa' | 'viewer' };

function Panel({ title, children }: { title: string; children: any }) {
  return (
    <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5">
      <h2 className="text-lg font-black text-slate-900 mb-4">{title}</h2>
      {children}
    </div>
  );
}

export default function AdminPage({ onBack }: { onBack: () => void }) {
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
      const data = await res.json();
      const list: Org[] = Array.isArray(data) ? data : [];
      setOrgs(list);

      const stored = Number(localStorage.getItem('selectedOrgId') || '0');
      const firstId = list[0]?.id || 0;
      const nextOrg = stored && list.some(o => o.id === stored) ? stored : firstId;
      if (nextOrg && nextOrg !== selectedOrgId) {
        setSelectedOrgId(nextOrg);
        localStorage.setItem('selectedOrgId', String(nextOrg));
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchProjects = async (orgId: number) => {
    if (!orgId) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/projects?org_id=${orgId}`, { headers: authHeaders });
      const data = await res.json();
      const list: Project[] = Array.isArray(data) ? data : [];
      setProjects(list);

      const stored = Number(localStorage.getItem('selectedProjectId') || '0');
      const firstId = list[0]?.id || 0;
      const nextProject = stored && list.some(p => p.id === stored) ? stored : firstId;
      if (nextProject && nextProject !== selectedProjectId) {
        setSelectedProjectId(nextProject);
        localStorage.setItem('selectedProjectId', String(nextProject));
      }
    } finally {
      setLoading(false);
    }
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
      const res = await fetch(`${API_URL}/orgs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ name }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Не удалось создать организацию');
      setNewOrgName('');
      await fetchOrgs();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  };

  const createProject = async () => {
    const name = newProjectName.trim();
    if (!name || !selectedOrgId) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/projects`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ org_id: selectedOrgId, name }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Не удалось создать проект');
      setNewProjectName('');
      await fetchProjects(selectedOrgId);
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  };

  const addOrgMember = async () => {
    if (!selectedOrgId) return;
    const email = orgMemberEmail.trim();
    if (!email) return;
    setOrgMemberHint(null);
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/orgs/${selectedOrgId}/members`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ email, role: orgMemberRole }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Не удалось добавить участника');
      setOrgMemberEmail('');
      await fetchOrgMembers(selectedOrgId);
      alert('Участник добавлен');
    } catch (e: any) {
      if (e.message === 'user_not_found') {
        setOrgMemberHint('Пользователь не зарегистрирован. Пусть сначала зарегистрируется, потом добавь по email.');
      } else if (e.message === 'not_allowed') {
        setOrgMemberHint('Недостаточно прав (нужен owner/admin организации).');
      } else {
        alert(e.message);
      }
    } finally {
      setLoading(false);
    }
  };

  const addProjectMember = async () => {
    if (!selectedProjectId) return;
    const email = projectMemberEmail.trim();
    if (!email) return;
    setProjectMemberHint(null);
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/projects/${selectedProjectId}/members`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ email, role: projectMemberRole }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Не удалось добавить участника в проект');
      setProjectMemberEmail('');
      await fetchProjectMembers(selectedProjectId);
      alert('Участник добавлен в проект');
    } catch (e: any) {
      if (e.message === 'user_not_found') {
        setProjectMemberHint('Пользователь не зарегистрирован. Пусть сначала зарегистрируется, потом добавь по email.');
      } else if (e.message === 'not_allowed') {
        setProjectMemberHint('Недостаточно прав (нужен owner/admin организации).');
      } else {
        alert(e.message);
      }
    } finally {
      setLoading(false);
    }
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

  const updateOrgMemberRole = async (userId: number, role: OrgMember['role']) => {
    if (!selectedOrgId) return;
    const res = await fetch(`${API_URL}/orgs/${selectedOrgId}/members/${userId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify({ role }),
    });
    if (!res.ok) return alert('Не удалось обновить роль');
    await fetchOrgMembers(selectedOrgId);
  };

  const deleteOrgMember = async (userId: number) => {
    if (!selectedOrgId) return;
    if (!confirm('Удалить участника из организации?')) return;
    const res = await fetch(`${API_URL}/orgs/${selectedOrgId}/members/${userId}`, {
      method: 'DELETE',
      headers: authHeaders,
    });
    if (!res.ok) return alert('Не удалось удалить участника');
    await fetchOrgMembers(selectedOrgId);
  };

  const updateProjectMemberRole = async (userId: number, role: ProjectMember['role']) => {
    if (!selectedProjectId) return;
    const res = await fetch(`${API_URL}/projects/${selectedProjectId}/members/${userId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify({ role }),
    });
    if (!res.ok) return alert('Не удалось обновить роль');
    await fetchProjectMembers(selectedProjectId);
  };

  const deleteProjectMember = async (userId: number) => {
    if (!selectedProjectId) return;
    if (!confirm('Удалить участника из проекта?')) return;
    const res = await fetch(`${API_URL}/projects/${selectedProjectId}/members/${userId}`, {
      method: 'DELETE',
      headers: authHeaders,
    });
    if (!res.ok) return alert('Не удалось удалить участника');
    await fetchProjectMembers(selectedProjectId);
  };

  return (
    <div className="min-h-screen bg-slate-50 p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="text-slate-500 hover:text-slate-900 text-2xl leading-none">←</button>
          <div>
            <h1 className="text-3xl font-black text-slate-900">Управление</h1>
            <p className="text-sm text-slate-500">Организации, проекты и доступы</p>
          </div>
          {loading && <span className="ml-auto text-sm text-slate-400">Загрузка…</span>}
        </div>

        <div className="flex flex-wrap gap-3">
          <select
            value={selectedOrgId || ''}
            onChange={(e) => {
              const v = Number(e.target.value);
              setSelectedOrgId(v);
              setSelectedProjectId(0);
              localStorage.setItem('selectedOrgId', String(v));
              localStorage.removeItem('selectedProjectId');
            }}
            className="bg-white border border-slate-200 text-slate-700 px-3 py-2.5 rounded-xl text-sm font-semibold"
          >
            {orgs.map(o => <option key={o.id} value={o.id}>{o.name}</option>)}
          </select>

          <select
            value={selectedProjectId || ''}
            onChange={(e) => {
              const v = Number(e.target.value);
              setSelectedProjectId(v);
              localStorage.setItem('selectedProjectId', String(v));
            }}
            className="bg-white border border-slate-200 text-slate-700 px-3 py-2.5 rounded-xl text-sm font-semibold"
            disabled={!selectedOrgId}
          >
            {projects.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          <Panel title="Создать организацию">
            <div className="flex gap-2">
              <input
                value={newOrgName}
                onChange={(e) => setNewOrgName(e.target.value)}
                placeholder="Название организации"
                className="flex-1 px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none"
              />
              <button
                onClick={createOrg}
                className="bg-indigo-600 text-white px-4 py-2.5 rounded-xl font-bold text-sm hover:bg-indigo-700"
              >
                Создать
              </button>
            </div>
            <p className="text-xs text-slate-400 mt-3">После создания ты станешь owner и сможешь выдавать роли.</p>
          </Panel>

          <Panel title="Создать проект (в выбранной орг)">
            <div className="flex gap-2">
              <input
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="Название проекта"
                className="flex-1 px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none"
                disabled={!selectedOrgId}
              />
              <button
                onClick={createProject}
                disabled={!selectedOrgId}
                className="bg-slate-900 text-white px-4 py-2.5 rounded-xl font-bold text-sm hover:bg-slate-800 disabled:opacity-50"
              >
                Создать
              </button>
            </div>
            {!selectedOrgId && <p className="text-xs text-slate-400 mt-3">Сначала выбери организацию.</p>}
          </Panel>

          <Panel title="Добавить участника в организацию">
            {!isOrgAdmin ? (
              <p className="text-sm text-slate-400">Нужно быть owner/admin выбранной организации.</p>
            ) : (
              <div className="space-y-3">
                <input
                  value={orgMemberEmail}
                  onChange={(e) => setOrgMemberEmail(e.target.value)}
                  placeholder="email пользователя"
                  className="w-full px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none"
                />
                <div className="flex gap-2">
                  <select
                    value={orgMemberRole}
                    onChange={(e) => setOrgMemberRole(e.target.value as any)}
                    className="flex-1 px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none bg-white"
                  >
                    <option value="member">member</option>
                    <option value="admin">admin</option>
                  </select>
                  <button
                    onClick={addOrgMember}
                    className="bg-emerald-600 text-white px-4 py-2.5 rounded-xl font-bold text-sm hover:bg-emerald-700"
                  >
                    Добавить
                  </button>
                </div>
                <div className="flex items-center justify-between gap-3">
                  <p className="text-xs text-slate-400">Пользователь должен быть зарегистрирован.</p>
                  <button
                    type="button"
                    onClick={async () => {
                      const link = `${window.location.origin}/login`;
                      await navigator.clipboard.writeText(link);
                      alert('Ссылка на регистрацию скопирована');
                    }}
                    className="text-xs font-bold text-indigo-600 hover:text-indigo-700"
                  >
                    Скопировать ссылку на регистрацию
                  </button>
                </div>
                {orgMemberHint && <p className="text-xs text-orange-600 bg-orange-50 border border-orange-100 rounded-xl p-2">{orgMemberHint}</p>}
              </div>
            )}
          </Panel>

          <Panel title="Участники организации">
            {!isOrgAdmin ? (
              <p className="text-sm text-slate-400">Нужно быть owner/admin выбранной организации.</p>
            ) : (
              <div className="space-y-2">
                {orgMembers.map((m) => (
                  <div key={m.user_id} className="flex items-center gap-2 p-2 rounded-xl border border-slate-100">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-semibold text-slate-700 truncate">{m.email}</p>
                    </div>
                    <select
                      value={m.role}
                      onChange={(e) => updateOrgMemberRole(m.user_id, e.target.value as OrgMember['role'])}
                      className="px-2 py-1.5 rounded-lg border border-slate-200 text-xs"
                    >
                      <option value="owner">owner</option>
                      <option value="admin">admin</option>
                      <option value="member">member</option>
                    </select>
                    <button onClick={() => deleteOrgMember(m.user_id)} className="text-xs font-bold text-red-600 hover:text-red-700">
                      Удалить
                    </button>
                  </div>
                ))}
                {orgMembers.length === 0 && <p className="text-xs text-slate-400">Пусто</p>}
              </div>
            )}
          </Panel>

          <Panel title="Добавить участника в проект">
            {!isOrgAdmin ? (
              <p className="text-sm text-slate-400">Нужно быть owner/admin выбранной организации.</p>
            ) : (
              <div className="space-y-3">
                <input
                  value={projectMemberEmail}
                  onChange={(e) => setProjectMemberEmail(e.target.value)}
                  placeholder="email пользователя"
                  className="w-full px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none"
                  disabled={!selectedProjectId}
                />
                <div className="flex gap-2">
                  <select
                    value={projectMemberRole}
                    onChange={(e) => setProjectMemberRole(e.target.value as any)}
                    className="flex-1 px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none bg-white"
                    disabled={!selectedProjectId}
                  >
                    <option value="pm">pm</option>
                    <option value="dev">dev</option>
                    <option value="qa">qa</option>
                    <option value="viewer">viewer</option>
                  </select>
                  <button
                    onClick={addProjectMember}
                    disabled={!selectedProjectId}
                    className="bg-blue-600 text-white px-4 py-2.5 rounded-xl font-bold text-sm hover:bg-blue-700 disabled:opacity-50"
                  >
                    Добавить
                  </button>
                </div>
                {!selectedProjectId && <p className="text-xs text-slate-400">Выбери проект.</p>}
                {projectMemberHint && <p className="text-xs text-orange-600 bg-orange-50 border border-orange-100 rounded-xl p-2">{projectMemberHint}</p>}
              </div>
            )}
          </Panel>

          <Panel title="Участники проекта">
            {!isOrgAdmin ? (
              <p className="text-sm text-slate-400">Нужно быть owner/admin выбранной организации.</p>
            ) : (
              <div className="space-y-2">
                {projectMembers.map((m) => (
                  <div key={m.user_id} className="flex items-center gap-2 p-2 rounded-xl border border-slate-100">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-semibold text-slate-700 truncate">{m.email}</p>
                    </div>
                    <select
                      value={m.role}
                      onChange={(e) => updateProjectMemberRole(m.user_id, e.target.value as ProjectMember['role'])}
                      className="px-2 py-1.5 rounded-lg border border-slate-200 text-xs"
                    >
                      <option value="pm">pm</option>
                      <option value="dev">dev</option>
                      <option value="qa">qa</option>
                      <option value="viewer">viewer</option>
                    </select>
                    <button onClick={() => deleteProjectMember(m.user_id)} className="text-xs font-bold text-red-600 hover:text-red-700">
                      Удалить
                    </button>
                  </div>
                ))}
                {projectMembers.length === 0 && <p className="text-xs text-slate-400">Пусто</p>}
              </div>
            )}
          </Panel>
        </div>
      </div>
    </div>
  );
}

