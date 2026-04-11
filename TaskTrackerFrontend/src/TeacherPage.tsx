import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_URL } from './config';
import { apiFetch } from './api';

const P = '#7C5CBF';

type User = { id: number; email: string; role: string; created_at?: string };

export default function TeacherPage() {
  const navigate = useNavigate();
  const role = localStorage.getItem('userRole') || 'student';

  useEffect(() => {
    if (role !== 'teacher') navigate('/', { replace: true });
  }, []);

  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch(`${API_URL}/admin/users`);
      if (res.status === 404 || res.status === 501) {
        setError('Эндпоинт /admin/users ещё не реализован на бэкенде.');
        return;
      }
      if (!res.ok) {
        setError('Не удалось загрузить пользователей.');
        return;
      }
      const data = await res.json().catch(() => []);
      setUsers(Array.isArray(data) ? data : []);
    } catch (e: any) {
      setError('Ошибка соединения с сервером.');
    } finally { setLoading(false); }
  };

  useEffect(() => { fetchUsers(); }, []);

  const changeRole = async (userId: number, newRole: string) => {
    try {
      const res = await apiFetch(`${API_URL}/admin/users/${userId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: newRole }),
      });
      if (!res.ok) throw new Error();
      setUsers(prev => prev.map(u => u.id === userId ? { ...u, role: newRole } : u));
    } catch { alert('Не удалось изменить роль'); }
  };

  const deleteUser = async (userId: number, email: string) => {
    if (!confirm(`Удалить пользователя ${email}?`)) return;
    try {
      const res = await apiFetch(`${API_URL}/admin/users/${userId}`, { method: 'DELETE' });
      if (!res.ok) throw new Error();
      setUsers(prev => prev.filter(u => u.id !== userId));
    } catch { alert('Не удалось удалить пользователя'); }
  };

  const myId = Number(localStorage.getItem('userId') || '0');

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-gray-900">Пользователи</h1>
        <button
          onClick={fetchUsers}
          className="text-sm font-medium px-3 py-2 rounded-xl border border-gray-200 text-gray-600 hover:border-[#7C5CBF] hover:text-[#7C5CBF] transition-colors"
        >
          Обновить
        </button>
      </div>

      {loading && (
        <div className="flex justify-center py-20">
          <div className="w-7 h-7 rounded-full border-2 border-t-transparent animate-spin" style={{ borderColor: `${P} transparent ${P} ${P}` }} />
        </div>
      )}

      {error && (
        <div className="bg-orange-50 border border-orange-100 rounded-2xl p-5 text-center">
          <p className="text-orange-600 text-sm font-medium mb-1">{error}</p>
          <p className="text-xs text-gray-400">Попросите Вадима реализовать GET /admin/users, PATCH /admin/users/&#123;id&#125;, DELETE /admin/users/&#123;id&#125;</p>
        </div>
      )}

      {!loading && !error && (
        <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
          {users.length === 0 ? (
            <p className="text-center text-gray-400 text-sm py-10">Нет пользователей</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100">
                  <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Email</th>
                  <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Роль</th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {users.map(u => (
                  <tr key={u.id} className="border-b border-gray-50 last:border-0 hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 text-gray-800 font-medium">
                      {u.email}
                      {u.id === myId && <span className="ml-2 text-xs text-gray-400">(вы)</span>}
                    </td>
                    <td className="px-4 py-3">
                      {u.id === myId ? (
                        <span className="text-xs bg-purple-100 text-purple-700 px-2 py-1 rounded-full font-semibold">{u.role}</span>
                      ) : (
                        <select
                          value={u.role}
                          onChange={e => changeRole(u.id, e.target.value)}
                          className="text-xs border border-gray-200 rounded-lg px-2 py-1 outline-none bg-white focus:border-[#7C5CBF]"
                        >
                          <option value="student">student</option>
                          <option value="teacher">teacher</option>
                        </select>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {u.id !== myId && (
                        <button
                          onClick={() => deleteUser(u.id, u.email)}
                          className="text-xs text-red-400 hover:text-red-600 font-medium transition-colors"
                        >
                          Удалить
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}
