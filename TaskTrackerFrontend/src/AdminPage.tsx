import { useEffect, useState } from 'react';
import { API_URL } from './config';
import { apiFetch } from './api';

const P = '#7C5CBF';

type Group = { id: number; name: string; role?: string };
type Project = { id: number; org_id: number; name: string; role?: string };
type Member = { user_id: number; email: string; role: string };

const inputCls = 'flex-1 px-3 py-2.5 rounded-xl border border-gray-200 text-sm outline-none focus:border-[#7C5CBF] bg-white';
const selectCls = 'px-3 py-2.5 rounded-xl border border-gray-200 text-sm outline-none bg-white';

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-5">
      <h2 className="text-sm font-bold text-gray-900 mb-4">{title}</h2>
      {children}
    </div>
  );
}

export default function AdminPage() {
  const role = localStorage.getItem('userRole') || 'student';
  const isTeacher = role === 'teacher';

  const [groups, setGroups] = useState<Group[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [selectedGroupId, setSelectedGroupId] = useState<number>(0);
  const [selectedProjectId, setSelectedProjectId] = useState<number>(
    () => Number(localStorage.getItem('selectedProjectId') || '0')
  );
  const [loading, setLoading] = useState(false);

  const [newGroupName, setNewGroupName] = useState('');
  const [newProjectName, setNewProjectName] = useState('');
  const [memberEmail, setMemberEmail] = useState('');
  const [memberHint, setMemberHint] = useState<string | null>(null);

  const fetchGroups = async () => {
    setLoading(true);
    try {
      const res = await apiFetch(`${API_URL}/orgs`);
      const data = await res.json().catch(() => []);
      const list: Group[] = Array.isArray(data) ? data : [];
      setGroups(list);
      const stored = Number(localStorage.getItem('selectedOrgId') || '0');
      const first = list[0]?.id || 0;
      const next = stored && list.some(g => g.id === stored) ? stored : first;
      if (next) { setSelectedGroupId(next); localStorage.setItem('selectedOrgId', String(next)); }
    } finally { setLoading(false); }
  };

  const fetchProjects = async (groupId: number) => {
    if (!groupId) return;
    setLoading(true);
    try {
      const res = await apiFetch(`${API_URL}/projects?org_id=${groupId}`);
      const data = await res.json().catch(() => []);
      const list: Project[] = Array.isArray(data) ? data : [];
      setProjects(list);
      const stored = Number(localStorage.getItem('selectedProjectId') || '0');
      const first = list[0]?.id || 0;
      const next = stored && list.some(p => p.id === stored) ? stored : first;
      if (next) { setSelectedProjectId(next); localStorage.setItem('selectedProjectId', String(next)); }
      else { setSelectedProjectId(0); }
    } finally { setLoading(false); }
  };

  const fetchMembers = async (projectId: number) => {
    if (!projectId) { setMembers([]); return; }
    const res = await apiFetch(`${API_URL}/projects/${projectId}/members`);
    const data = await res.json().catch(() => []);
    setMembers(Array.isArray(data) ? data : []);
  };

  useEffect(() => { fetchGroups(); }, []);
  useEffect(() => { if (selectedGroupId) fetchProjects(selectedGroupId); }, [selectedGroupId]);
  useEffect(() => { fetchMembers(selectedProjectId); }, [selectedProjectId]);

  const createGroup = async () => {
    const name = newGroupName.trim();
    if (!name) return;
    setLoading(true);
    try {
      const res = await apiFetch(`${API_URL}/orgs`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setNewGroupName('');
      await fetchGroups();
    } catch (e: any) { alert(e.message); }
    finally { setLoading(false); }
  };

  const createProject = async () => {
    const name = newProjectName.trim();
    if (!name || !selectedGroupId) return;
    setLoading(true);
    try {
      const res = await apiFetch(`${API_URL}/projects`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ org_id: selectedGroupId, name }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setNewProjectName('');
      await fetchProjects(selectedGroupId);
    } catch (e: any) { alert(e.message); }
    finally { setLoading(false); }
  };

  const addMember = async () => {
    const email = memberEmail.trim();
    if (!email || !selectedProjectId) return;
    setMemberHint(null);
    setLoading(true);
    try {
      const res = await apiFetch(`${API_URL}/projects/${selectedProjectId}/members`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, role: 'qa' }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Ошибка');
      setMemberEmail('');
      await fetchMembers(selectedProjectId);
    } catch (e: any) {
      if (e.message === 'user_not_found') setMemberHint('Пользователь не зарегистрирован');
      else setMemberHint(e.message);
    } finally { setLoading(false); }
  };

  const removeMember = async (userId: number) => {
    if (!selectedProjectId || !confirm('Удалить участника?')) return;
    await apiFetch(`${API_URL}/projects/${selectedProjectId}/members/${userId}`, { method: 'DELETE' });
    await fetchMembers(selectedProjectId);
  };

  // Student view: read-only project info
  if (!isTeacher) {
    return (
      <div className="p-6 max-w-2xl mx-auto">
        <h1 className="text-xl font-bold text-gray-900 mb-6">Мои проекты</h1>
        <Section title="Проект">
          <div className="space-y-2">
            <select
              value={selectedProjectId || ''}
              onChange={e => {
                const v = Number(e.target.value);
                setSelectedProjectId(v);
                localStorage.setItem('selectedProjectId', String(v));
              }}
              className={selectCls + ' w-full'}
            >
              <option value="">Выберите проект</option>
              {projects.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
          </div>
        </Section>
        {members.length > 0 && (
          <div className="mt-4">
            <Section title="Участники проекта">
              <div className="space-y-2">
                {members.map(m => (
                  <div key={m.user_id} className="flex items-center gap-2 py-1.5 border-b border-gray-50 last:border-0">
                    <span className="text-sm text-gray-700 flex-1">{m.email}</span>
                    <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">{m.role}</span>
                  </div>
                ))}
              </div>
            </Section>
          </div>
        )}
      </div>
    );
  }

  // Teacher view
  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-gray-900">Управление</h1>
        {loading && <span className="text-xs text-gray-400">Загрузка…</span>}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">

        {/* Create group */}
        <Section title="Создать группу">
          <div className="flex gap-2">
            <input
              value={newGroupName}
              onChange={e => setNewGroupName(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && createGroup()}
              placeholder="Название группы"
              className={inputCls}
            />
            <button onClick={createGroup} className="text-white px-4 py-2.5 rounded-xl font-semibold text-sm shrink-0" style={{ background: P }}>
              Создать
            </button>
          </div>
          {groups.length > 0 && (
            <div className="mt-3 space-y-1">
              {groups.map(g => (
                <button
                  key={g.id}
                  onClick={() => { setSelectedGroupId(g.id); localStorage.setItem('selectedOrgId', String(g.id)); }}
                  className="w-full text-left text-sm px-3 py-2 rounded-xl transition-colors"
                  style={selectedGroupId === g.id ? { background: '#EDE9F7', color: P, fontWeight: 600 } : { color: '#374151' }}
                >
                  {g.name}
                </button>
              ))}
            </div>
          )}
        </Section>

        {/* Create project */}
        <Section title="Создать проект">
          {!selectedGroupId ? (
            <p className="text-sm text-gray-400">Сначала выберите группу.</p>
          ) : (
            <>
              <div className="flex gap-2 mb-3">
                <input
                  value={newProjectName}
                  onChange={e => setNewProjectName(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && createProject()}
                  placeholder="Название проекта"
                  className={inputCls}
                />
                <button onClick={createProject} className="bg-gray-900 text-white px-4 py-2.5 rounded-xl font-semibold text-sm shrink-0">
                  Создать
                </button>
              </div>
              {projects.length > 0 && (
                <div className="space-y-1">
                  {projects.map(p => (
                    <button
                      key={p.id}
                      onClick={() => { setSelectedProjectId(p.id); localStorage.setItem('selectedProjectId', String(p.id)); }}
                      className="w-full text-left text-sm px-3 py-2 rounded-xl transition-colors"
                      style={selectedProjectId === p.id ? { background: '#EDE9F7', color: P, fontWeight: 600 } : { color: '#374151' }}
                    >
                      {p.name}
                    </button>
                  ))}
                </div>
              )}
            </>
          )}
        </Section>

        {/* Add member */}
        <Section title="Добавить студента в проект">
          {!selectedProjectId ? (
            <p className="text-sm text-gray-400">Сначала выберите проект.</p>
          ) : (
            <div className="space-y-2">
              <div className="flex gap-2">
                <input
                  value={memberEmail}
                  onChange={e => { setMemberEmail(e.target.value); setMemberHint(null); }}
                  onKeyDown={e => e.key === 'Enter' && addMember()}
                  placeholder="Email студента"
                  className={inputCls}
                />
                <button onClick={addMember} className="text-white px-4 py-2.5 rounded-xl font-semibold text-sm shrink-0" style={{ background: '#10B981' }}>
                  Добавить
                </button>
              </div>
              <div className="flex items-center justify-between">
                <p className="text-xs text-gray-400">Студент должен быть зарегистрирован</p>
                <button
                  onClick={async () => { await navigator.clipboard.writeText(`${window.location.origin}/login`); alert('Ссылка скопирована'); }}
                  className="text-xs font-bold"
                  style={{ color: P }}
                >
                  Скопировать ссылку
                </button>
              </div>
              {memberHint && <p className="text-xs text-orange-600 bg-orange-50 border border-orange-100 rounded-xl p-2">{memberHint}</p>}
            </div>
          )}
        </Section>

        {/* Members list */}
        <Section title="Участники проекта">
          {!selectedProjectId ? (
            <p className="text-sm text-gray-400">Выберите проект.</p>
          ) : members.length === 0 ? (
            <p className="text-sm text-gray-400">Нет участников</p>
          ) : (
            <div className="space-y-2">
              {members.map(m => (
                <div key={m.user_id} className="flex items-center gap-2 p-2 rounded-xl border border-gray-100">
                  <span className="text-sm text-gray-700 flex-1 truncate">{m.email}</span>
                  <span className="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full shrink-0">{m.role}</span>
                  <button onClick={() => removeMember(m.user_id)} className="text-xs font-bold text-red-400 hover:text-red-600 shrink-0">
                    Удалить
                  </button>
                </div>
              ))}
            </div>
          )}
        </Section>
      </div>
    </div>
  );
}
