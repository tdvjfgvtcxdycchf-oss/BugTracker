import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_URL } from './config';
import BugsModal from './BugsModal';
import BugDetailEditor from './BugDetailEditor';

const P = '#7C5CBF';

function formatDate(dateStr?: string) {
  if (!dateStr) return '—';
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return '—';
  const today = new Date();
  if (d.toDateString() === today.toDateString()) return 'Сегодня';
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' });
}

function ChevronRight() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="9 18 15 12 9 6"/>
    </svg>
  );
}

function SearchIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
    </svg>
  );
}

function OrgIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
      <circle cx="9" cy="7" r="4"/>
      <path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
    </svg>
  );
}

export default function MainPage() {
  const navigate = useNavigate();
  const [isEditorOpen, setIsEditorOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState<any>(null);
  const [selectedBugId, setSelectedBugId] = useState<string | undefined>();
  const [tasks, setTasks] = useState<any[]>([]);
  const [orgs, setOrgs] = useState<any[]>([]);
  const [projects, setProjects] = useState<any[]>([]);
  const [selectedOrgId, setSelectedOrgId] = useState<number>(() => Number(localStorage.getItem('selectedOrgId') || '0'));
  const [selectedProjectId, setSelectedProjectId] = useState<number>(() => Number(localStorage.getItem('selectedProjectId') || '0'));
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState('');
  const currentUserId = Number(localStorage.getItem('userId') || '0');
  const jwtToken = localStorage.getItem('jwtToken') || '';
  const authHeaders = jwtToken ? { Authorization: `Bearer ${jwtToken}` } : {};

  const fetchOrgs = async () => {
    try {
      const res = await fetch(`${API_URL}/orgs`, { headers: authHeaders });
      const data = await res.json();
      const list = Array.isArray(data) ? data : [];
      setOrgs(list);
      const stored = Number(localStorage.getItem('selectedOrgId') || '0');
      const firstId = list[0]?.id || 0;
      const nextOrg = stored && list.some((o: any) => o.id === stored) ? stored : firstId;
      if (nextOrg && nextOrg !== selectedOrgId) {
        setSelectedOrgId(nextOrg);
        localStorage.setItem('selectedOrgId', String(nextOrg));
      }
      const role = list.find((o: any) => o.id === (nextOrg || selectedOrgId))?.role;
      if (role) localStorage.setItem('selectedOrgRole', String(role));
    } catch (e) { console.error(e); }
  };

  const fetchProjects = async (orgId: number) => {
    if (!orgId) return;
    try {
      const res = await fetch(`${API_URL}/projects?org_id=${orgId}`, { headers: authHeaders });
      const data = await res.json();
      const list = Array.isArray(data) ? data : [];
      setProjects(list);
      const stored = Number(localStorage.getItem('selectedProjectId') || '0');
      const firstId = list[0]?.id || 0;
      const nextProject = stored && list.some((p: any) => p.id === stored) ? stored : firstId;
      if (nextProject && nextProject !== selectedProjectId) {
        setSelectedProjectId(nextProject);
        localStorage.setItem('selectedProjectId', String(nextProject));
      }
    } catch (e) { console.error(e); }
  };

  const fetchTasks = async () => {
    setIsLoading(true);
    try {
      if (!selectedProjectId) { setTasks([]); return; }
      const res = await fetch(`${API_URL}/tasks?project_id=${selectedProjectId}`, { headers: authHeaders });
      const data = await res.json();
      setTasks(data || []);
    } catch (err) { console.error(err); }
    finally { setIsLoading(false); }
  };

  useEffect(() => { fetchOrgs(); }, []);
  useEffect(() => { if (selectedOrgId) fetchProjects(selectedOrgId); }, [selectedOrgId]);
  useEffect(() => { fetchTasks(); }, [selectedProjectId]);

  const handleDeleteTask = async (taskId: number) => {
    if (!taskId || !currentUserId) return;
    if (!confirm('Удалить задачу?')) return;
    try {
      const res = await fetch(`${API_URL}/tasks/${taskId}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ owner_id: currentUserId }),
      });
      if (!res.ok) throw new Error(`Delete failed: ${res.status}`);
      if (selectedTask?.id === taskId) { setSelectedTask(null); setIsEditorOpen(false); setSelectedBugId(undefined); }
      await fetchTasks();
    } catch { alert('Не удалось удалить задачу'); }
  };

  const handleBugSavedInState = (updatedBugs: any[]) => {
    setTasks(prev => prev.map(t => t.id === selectedTask?.id ? { ...t, bugs: updatedBugs } : t));
  };

  if (!isLoading && orgs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full py-32 px-4 text-center">
        <div className="text-6xl mb-5">🏢</div>
        <h2 className="text-xl font-bold text-gray-900 mb-2">Добро пожаловать!</h2>
        <p className="text-gray-400 mb-6 max-w-xs text-sm">Создайте первую организацию и проект, чтобы начать работу.</p>
        <button
          onClick={() => navigate('/admin')}
          className="text-white px-6 py-3 rounded-xl font-semibold text-sm"
          style={{ background: P }}
        >
          Создать организацию
        </button>
      </div>
    );
  }

  const visible = tasks.filter(t =>
    !search || t.title?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <h1 className="text-xl font-bold text-gray-900 text-center mb-5">Мои задачи</h1>

      {/* Search */}
      <div className="relative mb-4">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
          <SearchIcon />
        </span>
        <input
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Поиск..."
          className="w-full bg-white border border-gray-200 rounded-xl pl-9 pr-4 py-2.5 text-sm text-gray-700 outline-none focus:border-[#7C5CBF]"
        />
      </div>

      {/* Filters */}
      <div className="flex gap-3 mb-5">
        <div className="flex items-center gap-1.5 bg-white border border-gray-200 rounded-xl px-3 py-2 flex-1">
          <span className="text-gray-400 shrink-0"><OrgIcon /></span>
          <select
            value={selectedOrgId || ''}
            onChange={e => {
              const v = Number(e.target.value);
              setSelectedOrgId(v);
              setSelectedProjectId(0);
              localStorage.setItem('selectedOrgId', String(v));
              localStorage.removeItem('selectedProjectId');
            }}
            className="flex-1 text-sm text-gray-700 bg-transparent outline-none"
          >
            <option value="">Выберите организацию</option>
            {orgs.map((o: any) => <option key={o.id} value={o.id}>{o.name}</option>)}
          </select>
        </div>
        <div className="flex items-center gap-1.5 bg-white border border-gray-200 rounded-xl px-3 py-2 flex-1">
          <select
            value={selectedProjectId || ''}
            onChange={e => {
              const v = Number(e.target.value);
              setSelectedProjectId(v);
              localStorage.setItem('selectedProjectId', String(v));
            }}
            className="flex-1 text-sm text-gray-700 bg-transparent outline-none"
            disabled={!selectedOrgId}
          >
            <option value="">Выберите проект</option>
            {projects.map((p: any) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </select>
        </div>
      </div>

      {/* Task list */}
      {isLoading ? (
        <div className="flex justify-center py-20">
          <div className="w-7 h-7 rounded-full border-2 border-t-transparent animate-spin" style={{ borderColor: `${P} transparent ${P} ${P}` }} />
        </div>
      ) : visible.length === 0 ? (
        <div className="text-center py-20">
          <p className="text-gray-400 text-sm">{tasks.length > 0 ? 'Ничего не найдено' : 'Задач пока нет'}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {visible.map(task => (
            <div
              key={task.id}
              onClick={() => setSelectedTask(task)}
              className="group flex items-center gap-3 bg-white rounded-xl border border-gray-100 px-4 py-3 cursor-pointer hover:border-[#C4B0E8] hover:shadow-sm transition-all"
            >
              <span className="text-gray-200 select-none text-xs leading-none">⠿⠿</span>
              <span className="flex-1 text-sm font-medium text-gray-800 truncate">{task.title}</span>
              <span className="text-xs text-gray-400 shrink-0">Тип: задача</span>
              <span className="text-xs text-gray-400 shrink-0">Изменено: {formatDate(task.updated_at || task.created_at)}</span>
              {task.owner_id === currentUserId && (
                <button
                  type="button"
                  onClick={e => { e.stopPropagation(); handleDeleteTask(task.id); }}
                  className="opacity-0 group-hover:opacity-100 text-xs text-red-400 hover:text-red-600 transition-all shrink-0"
                >
                  ✕
                </button>
              )}
              <span className="text-gray-300 shrink-0"><ChevronRight /></span>
            </div>
          ))}
        </div>
      )}

      {selectedTask && !isEditorOpen && (
        <BugsModal
          task={tasks.find(t => t.id === selectedTask.id) || selectedTask}
          onClose={() => setSelectedTask(null)}
          setIsEditorOpen={setIsEditorOpen}
          setSelectedBugId={setSelectedBugId}
          onBugsLoaded={handleBugSavedInState}
        />
      )}

      {isEditorOpen && selectedTask && (
        <BugDetailEditor
          isOpen={isEditorOpen}
          onClose={() => { setIsEditorOpen(false); setSelectedBugId(undefined); }}
          task={tasks.find(t => t.id === selectedTask.id) || selectedTask}
          currentBug={tasks.find(t => t.id === selectedTask.id)?.bugs?.find((b: any) => b.id === Number(selectedBugId))}
          bugId={selectedBugId}
          onBugSaved={handleBugSavedInState}
        />
      )}
    </div>
  );
}
