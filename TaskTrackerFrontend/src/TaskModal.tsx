import React, { useState } from 'react';

interface TaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (task: { name: string; desc: string; photo?: string }) => void;
}

export default function TaskModal({ isOpen, onClose, onCreate }: TaskModalProps) {
  const [name, setName] = useState('');
  const [desc, setDesc] = useState('');
  const [photo, setPhoto] = useState<string | undefined>();
  const [photoName, setPhotoName] = useState('');

  if (!isOpen) return null;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setPhotoName(file.name);
    const reader = new FileReader();
    reader.onload = () => setPhoto(reader.result as string);
    reader.readAsDataURL(file);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCreate({ name, desc, photo });
    setName('');
    setDesc('');
    setPhoto(undefined);
    setPhotoName('');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-100 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl overflow-hidden">
        <div className="flex justify-between items-center px-6 py-4 border-b">
          <h2 className="text-xl font-bold text-gray-800">Создать задачу</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-2xl">×</button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Название задачи</label>
            <input
              type="text"
              required
              className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none"
              placeholder="Введите название задачи..."
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div>
            <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Описание</label>
            <textarea
              required
              rows={3}
              className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none"
              placeholder="Опишите требования..."
              value={desc}
              onChange={(e) => setDesc(e.target.value)}
            />
          </div>

          <div>
            <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Фото (необязательно)</label>
            <label className="flex items-center gap-3 cursor-pointer w-full bg-gray-50 border border-dashed border-gray-300 rounded-lg px-4 py-2.5 hover:border-blue-400 transition-colors">
              <span className="text-gray-400 text-lg">📎</span>
              <span className="text-sm text-gray-500 truncate">{photoName || 'Выбрать изображение...'}</span>
              <input type="file" accept="image/*" className="hidden" onChange={handleFileChange} />
            </label>
            {photo && (
              <img src={photo} alt="preview" className="mt-2 w-full max-h-32 object-cover rounded-lg border border-gray-200" />
            )}
          </div>

          <div className="flex gap-3 pt-4">
            <button type="button" onClick={onClose} className="flex-1 py-3 text-gray-500 font-semibold hover:bg-gray-100 rounded-xl">Отмена</button>
            <button type="submit" className="flex-1 py-3 bg-blue-600 text-white font-semibold rounded-xl shadow-lg hover:bg-blue-700">Создать задачу</button>
          </div>
        </form>
      </div>
    </div>
  );
}