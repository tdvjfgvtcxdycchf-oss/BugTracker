import { useEffect, useMemo, useState } from 'react';
import { API_URL } from './config';

export default function ProfilePage({ onBack }: { onBack: () => void }) {
  const jwtToken = localStorage.getItem('jwtToken') || '';
  const authHeaders = useMemo(() => (jwtToken ? { Authorization: `Bearer ${jwtToken}` } : {}), [jwtToken]);

  const [email, setEmail] = useState('');
  const [role, setRole] = useState('');
  const [loading, setLoading] = useState(false);

  const [newEmail, setNewEmail] = useState('');
  const [emailCurrentPassword, setEmailCurrentPassword] = useState('');

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');

  const fetchMe = async () => {
    const res = await fetch(`${API_URL}/me`, { headers: authHeaders });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) return;
    setEmail(data.email || '');
    setRole(data.role || '');
    setNewEmail(data.email || '');
  };

  useEffect(() => { fetchMe(); }, []);

  const changeEmail = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/me/email`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ new_email: newEmail.trim(), current_password: emailCurrentPassword }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Не удалось сменить email');
      localStorage.setItem('userEmail', newEmail.trim());
      setEmailCurrentPassword('');
      await fetchMe();
      alert('Email обновлён');
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  };

  const changePassword = async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/me/password`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', ...authHeaders },
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data?.error || 'Не удалось сменить пароль');
      setCurrentPassword('');
      setNewPassword('');
      localStorage.clear();
      alert('Пароль обновлён. Войди заново.');
      window.location.href = '/login';
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  };

  const logoutAll = async () => {
    if (!confirm('Выйти из всех сессий?')) return;
    setLoading(true);
    try {
      const res = await fetch(`${API_URL}/me/logout-all`, {
        method: 'POST',
        headers: authHeaders,
      });
      if (!res.ok) throw new Error('Не удалось завершить все сессии');
      localStorage.clear();
      window.location.href = '/login';
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 p-6">
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="text-slate-500 hover:text-slate-900 text-2xl leading-none">←</button>
          <div>
            <h1 className="text-3xl font-black text-slate-900">Профиль</h1>
            <p className="text-sm text-slate-500">Аккаунт и безопасность</p>
          </div>
        </div>

        <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5">
          <p className="text-sm text-slate-500">Текущий email</p>
          <p className="text-lg font-bold text-slate-900">{email || '—'}</p>
          <p className="text-xs text-slate-400 mt-1">Роль: {role || '—'}</p>
        </div>

        <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5 space-y-3">
          <h2 className="text-lg font-black text-slate-900">Сменить email</h2>
          <input value={newEmail} onChange={(e) => setNewEmail(e.target.value)} className="w-full px-3 py-2.5 rounded-xl border border-slate-200 text-sm" placeholder="Новый email" />
          <input type="password" value={emailCurrentPassword} onChange={(e) => setEmailCurrentPassword(e.target.value)} className="w-full px-3 py-2.5 rounded-xl border border-slate-200 text-sm" placeholder="Текущий пароль" />
          <button disabled={loading} onClick={changeEmail} className="px-4 py-2.5 rounded-xl bg-indigo-600 text-white font-bold text-sm disabled:opacity-50">Сохранить email</button>
        </div>

        <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-5 space-y-3">
          <h2 className="text-lg font-black text-slate-900">Сменить пароль</h2>
          <input type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} className="w-full px-3 py-2.5 rounded-xl border border-slate-200 text-sm" placeholder="Текущий пароль" />
          <input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} className="w-full px-3 py-2.5 rounded-xl border border-slate-200 text-sm" placeholder="Новый пароль (мин. 8, A-z, 0-9)" />
          <button disabled={loading} onClick={changePassword} className="px-4 py-2.5 rounded-xl bg-slate-900 text-white font-bold text-sm disabled:opacity-50">Сменить пароль</button>
        </div>

        <div className="bg-white rounded-2xl border border-red-100 shadow-sm p-5 space-y-3">
          <h2 className="text-lg font-black text-red-700">Сессии</h2>
          <p className="text-sm text-slate-500">Завершит все активные сессии на всех устройствах.</p>
          <button disabled={loading} onClick={logoutAll} className="px-4 py-2.5 rounded-xl bg-red-600 text-white font-bold text-sm disabled:opacity-50">Выйти из всех сессий</button>
        </div>
      </div>
    </div>
  );
}

