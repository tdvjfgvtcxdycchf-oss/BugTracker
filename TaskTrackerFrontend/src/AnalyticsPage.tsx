import { useState, useEffect } from 'react';
import { API_URL } from './config';

interface StatItem { status: string; count: number; }
interface Task { id: number; title: string; }

const STATUS_COLORS: Record<string, string> = {
  'New': 'bg-slate-400',
  'Open': 'bg-blue-500',
  'In Progress': 'bg-indigo-500',
  'Fixed': 'bg-emerald-500',
  'Ready for Retest': 'bg-cyan-500',
  'Verified': 'bg-green-600',
  'Reopened': 'bg-red-500',
  'Rejected': 'bg-orange-500',
  "Can't Reproduce": 'bg-yellow-500',
};

export default function AnalyticsPage({ onBack }: { onBack: () => void }) {
  const jwtToken = localStorage.getItem('jwtToken') || '';
  const authHeaders = jwtToken ? { Authorization: `Bearer ${jwtToken}` } : {};

  const [stats, setStats] = useState<StatItem[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [selectedTaskId, setSelectedTaskId] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetch(`${API_URL}/tasks`, { headers: authHeaders })
      .then(r => r.json())
      .then(d => setTasks(Array.isArray(d) ? d : []))
      .catch(console.error);
  }, []);

  useEffect(() => {
    setLoading(true);
    const url = selectedTaskId
      ? `${API_URL}/stats/tasks/${selectedTaskId}`
      : `${API_URL}/stats`;
    fetch(url, { headers: authHeaders })
      .then(r => r.json())
      .then(d => setStats(Array.isArray(d) ? d : []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [selectedTaskId]);

  const total = stats.reduce((s, x) => s + x.count, 0);
  const sorted = [...stats].sort((a, b) => b.count - a.count);

  return (
    <div className="min-h-screen bg-slate-50 p-6">
      <div className="max-w-3xl mx-auto space-y-8">

        <div className="flex items-center gap-4">
          <button onClick={onBack} className="text-slate-500 hover:text-slate-900 text-2xl leading-none">←</button>
          <div>
            <h1 className="text-3xl font-black text-slate-900">Аналитика</h1>
            <p className="text-sm text-slate-500">Статистика по багам</p>
          </div>
        </div>

        <div className="flex gap-3 items-center">
          <select
            value={selectedTaskId}
            onChange={e => setSelectedTaskId(e.target.value)}
            className="px-4 py-2 rounded-xl border border-slate-200 bg-white text-sm outline-none"
          >
            <option value="">Все задачи</option>
            {tasks.map(t => <option key={t.id} value={t.id}>{t.title}</option>)}
          </select>
          {loading && <span className="text-sm text-slate-400">Загрузка...</span>}
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div className="bg-white rounded-2xl p-5 shadow-sm border border-slate-100 col-span-2 sm:col-span-4">
            <p className="text-xs font-bold text-slate-400 uppercase mb-1">Всего багов</p>
            <p className="text-5xl font-black text-slate-900">{total}</p>
          </div>
          {sorted.slice(0, 4).map(s => (
            <div key={s.status} className="bg-white rounded-2xl p-4 shadow-sm border border-slate-100">
              <p className="text-[10px] font-bold text-slate-400 uppercase mb-1 truncate">{s.status}</p>
              <p className="text-3xl font-black text-slate-900">{s.count}</p>
            </div>
          ))}
        </div>

        <div className="bg-white rounded-2xl p-6 shadow-sm border border-slate-100 space-y-4">
          <h2 className="text-lg font-black text-slate-900">По статусам</h2>
          {total === 0 && !loading && <p className="text-sm text-slate-400">Нет данных</p>}
          {sorted.map(s => {
            const pct = total > 0 ? Math.round((s.count / total) * 100) : 0;
            const color = STATUS_COLORS[s.status] ?? 'bg-slate-300';
            return (
              <div key={s.status}>
                <div className="flex justify-between text-sm mb-1">
                  <span className="font-semibold text-slate-700">{s.status}</span>
                  <span className="text-slate-500">{s.count} ({pct}%)</span>
                </div>
                <div className="w-full bg-slate-100 rounded-full h-2.5">
                  <div className={`${color} h-2.5 rounded-full transition-all`} style={{ width: `${pct}%` }} />
                </div>
              </div>
            );
          })}
        </div>

      </div>
    </div>
  );
}
