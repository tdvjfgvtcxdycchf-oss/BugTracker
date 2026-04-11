import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_URL } from './config';
import { apiFetch } from './api';

const P = '#7C5CBF';

export default function NewTaskPage() {
  const navigate = useNavigate();
  const [taskType, setTaskType] = useState('Задача');
  const [title, setTitle] = useState('');
  const [desc, setDesc] = useState('');
  const [pending, setPending] = useState(false);

  const selectedProjectId = Number(localStorage.getItem('selectedProjectId') || '0');

  const handleCreate = async () => {
    const taskTitle = title.trim() || `${taskType}: ${desc.trim().slice(0, 40)}`;
    if (!taskTitle || !selectedProjectId) {
      alert(!selectedProjectId ? 'Выберите проект в разделе «Мои задачи»' : 'Введите описание');
      return;
    }
    setPending(true);
    try {
      const res = await apiFetch(`${API_URL}/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: taskTitle, description: desc.trim(), project_id: selectedProjectId }),
      });
      if (!res.ok) throw new Error('Не удалось создать задачу');
      navigate('/');
    } catch (err: any) {
      alert(err.message);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-full p-6">
      <div className="bg-white rounded-2xl p-8 w-full max-w-sm shadow-sm border border-gray-100">
        <h1 className="text-xl font-bold text-center text-gray-900 mb-6">Создать задачу</h1>

        <div className="mb-4">
          <p className="text-sm text-gray-500 mb-2">Выберите тип задачи</p>
          <div className="relative">
            <select
              value={taskType}
              onChange={e => setTaskType(e.target.value)}
              className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm text-gray-700 bg-white appearance-none outline-none focus:border-[#7C5CBF]"
            >
              <option>Баг</option>
              <option>Задача</option>
              <option>Улучшение</option>
            </select>
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none">▾</span>
          </div>
        </div>

        <div className="mb-4">
          <input
            placeholder="Название задачи (необязательно)"
            value={title}
            onChange={e => setTitle(e.target.value)}
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm text-gray-700 outline-none focus:border-[#7C5CBF]"
          />
        </div>

        <div className="mb-6">
          <textarea
            placeholder="Описание..."
            value={desc}
            onChange={e => setDesc(e.target.value)}
            rows={5}
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm text-gray-700 resize-none outline-none focus:border-[#7C5CBF]"
          />
        </div>

        <button
          onClick={handleCreate}
          disabled={pending}
          className="w-full py-3 rounded-xl text-white font-semibold text-sm transition-opacity disabled:opacity-60"
          style={{ background: P }}
        >
          {pending ? 'Создание...' : 'Создать задачу'}
        </button>
      </div>
    </div>
  );
}
