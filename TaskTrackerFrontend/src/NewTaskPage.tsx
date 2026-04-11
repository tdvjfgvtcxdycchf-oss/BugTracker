import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_URL } from './config';
import { apiFetch } from './api';

const P = '#7C5CBF';

export default function NewTaskPage() {
  const navigate = useNavigate();
  const [taskType, setTaskType] = useState('Задача');
  const [title, setTitle] = useState('');
  const [desc, setDesc] = useState('');
  const [photo, setPhoto] = useState<File | null>(null);
  const [photoPreview, setPhotoPreview] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  const selectedProjectId = Number(localStorage.getItem('selectedProjectId') || '0');

  const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] || null;
    setPhoto(file);
    if (file) {
      const reader = new FileReader();
      reader.onload = ev => setPhotoPreview(ev.target?.result as string);
      reader.readAsDataURL(file);
    } else {
      setPhotoPreview(null);
    }
  };

  const removePhoto = () => {
    setPhoto(null);
    setPhotoPreview(null);
    if (fileRef.current) fileRef.current.value = '';
  };

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
      const data = await res.json();
      const taskId = data.id;

      if (photo && taskId) {
        const form = new FormData();
        form.append('photo', photo);
        await apiFetch(`${API_URL}/tasks/${taskId}/photo`, { method: 'POST', body: form });
      }

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

        <div className="mb-4">
          <textarea
            placeholder="Описание..."
            value={desc}
            onChange={e => setDesc(e.target.value)}
            rows={4}
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm text-gray-700 resize-none outline-none focus:border-[#7C5CBF]"
          />
        </div>

        {/* Photo */}
        <div className="mb-6">
          {photoPreview ? (
            <div className="relative rounded-xl overflow-hidden border border-gray-200">
              <img src={photoPreview} alt="preview" className="w-full max-h-48 object-cover" />
              <button
                onClick={removePhoto}
                className="absolute top-2 right-2 bg-black/50 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs hover:bg-black/70 transition-colors"
              >
                ✕
              </button>
            </div>
          ) : (
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              className="w-full border-2 border-dashed border-gray-200 rounded-xl py-4 flex flex-col items-center gap-1 text-gray-400 hover:border-[#7C5CBF] hover:text-[#7C5CBF] transition-colors"
            >
              <span className="text-2xl">📎</span>
              <span className="text-xs font-medium">Прикрепить фото</span>
              <span className="text-[10px]">до 15 МБ</span>
            </button>
          )}
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={handlePhotoChange}
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
