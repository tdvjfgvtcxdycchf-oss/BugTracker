import React, { useState } from 'react';

interface TaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (task: { name: string; desc: string }) => void;
}

export default function TaskModal({ isOpen, onClose, onCreate }: TaskModalProps) {
  const [name, setName] = useState('');
  const [desc, setDesc] = useState('');
  const [email, setEmail] = useState('');

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCreate({ name, desc });
    // Сброс полей
    setName('');
    setDesc('');
    setEmail('');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-100 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl overflow-hidden">
        <div className="flex justify-between items-center px-6 py-4 border-b">
          <h2 className="text-xl font-bold text-gray-800">Create New Task</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-2xl">×</button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Поле: Название */}
          <div>
            <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Task Name</label>
            <input
              type="text"
              required
              className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none"
              placeholder="Enter task title..."
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

        

          {/* Поле: Описание */}
          <div>
            <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Description</label>
            <textarea
              required
              rows={3}
              className="w-full bg-gray-50 border border-gray-200 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-blue-500 outline-none"
              placeholder="Describe requirements..."
              value={desc}
              onChange={(e) => setDesc(e.target.value)}
            />
          </div>

          <div className="flex gap-3 pt-4">
            <button type="button" onClick={onClose} className="flex-1 py-3 text-gray-500 font-semibold hover:bg-gray-100 rounded-xl">Cancel</button>
            <button type="submit" className="flex-1 py-3 bg-blue-600 text-white font-semibold rounded-xl shadow-lg hover:bg-blue-700">Create Task</button>
          </div>
        </form>
      </div>
    </div>
  );
}