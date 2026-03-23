import React, { useState, useEffect } from 'react';

interface BugsModalProps {
  task: any;
  onClose: () => void;
  setIsEditorOpen: (open: boolean) => void;
  setSelectedBugId: (id: string | undefined) => void;
  onBugsLoaded: (bugs: any[]) => void; // Добавили колбэк
}

export default function BugsModal({ task, onClose, setIsEditorOpen, setSelectedBugId, onBugsLoaded }: BugsModalProps) {
  const [bugs, setBugs] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const currentUserId = Number(localStorage.getItem('userId') || '0');
  const [pendingDeleteBugId, setPendingDeleteBugId] = useState<number | null>(null);

  const fetchBugs = async () => {
    if (!task?.id) return;
    setLoading(true);
    try {
      const baseUrl = (import.meta as any).env.VITE_API_URL;
      const response = await fetch(`${baseUrl}/bugs/${task.id}`);
      const data = await response.json();
      const loadedBugs = data || [];
      setBugs(loadedBugs);
      onBugsLoaded(loadedBugs); // Синхронизируем с Dashboard
    } catch (err) { console.error(err); }
    finally { setLoading(false); }
  };

  useEffect(() => {
    fetchBugs();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [task?.id]);

  const handleDeleteBug = async (bugId: number) => {
    if (!bugId || !currentUserId) return;
    if (!confirm('Удалить баг?')) return;

    setPendingDeleteBugId(bugId);
    try {
      const baseUrl = (import.meta as any).env.VITE_API_URL;
      const res = await fetch(`${baseUrl}/bugs/${bugId}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ created_by: currentUserId }),
      });

      if (!res.ok) throw new Error(`Delete failed: ${res.status}`);
      await fetchBugs();
    } catch (err) {
      console.error(err);
      alert('Не удалось удалить баг');
    } finally {
      setPendingDeleteBugId(null);
    }
  };

  if (!task) return null;

  return (
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/50 backdrop-blur-md p-4">
      <div className="bg-white w-full max-w-lg rounded-2xl shadow-2xl flex flex-col max-h-[80vh]">
        <div className="p-6 border-b flex justify-between items-center bg-gray-50 rounded-t-2xl">
          <div>
            <h2 className="text-xl font-black text-gray-900 line-clamp-1">{task.title}</h2>
            <p className="text-[10px] text-blue-500 font-bold uppercase tracking-widest mt-1">Bugs Repository</p>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-900 text-3xl">&times;</button>
        </div>

        <div className="flex-1 overflow-y-auto p-6 space-y-3">
          {loading ? (
            <div className="text-center py-10 text-gray-400">Загрузка багов...</div>
          ) : bugs.length > 0 ? (
            bugs.map((bug: any) => (
              <div key={bug.id} onClick={() => { setSelectedBugId(bug.id); setIsEditorOpen(true); }}
                className="p-4 bg-red-50 border border-red-100 rounded-xl flex items-center justify-between cursor-pointer hover:border-red-400 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="w-2 h-2 bg-red-500 rounded-full"></div>
                  <span className="text-sm font-bold text-gray-700">Bug ID: {bug.id}</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-[10px] font-bold text-red-400 uppercase">{bug.status}</span>
                  {bug.created_by === currentUserId && (
                    <button
                      type="button"
                      onClick={(e) => { e.stopPropagation(); handleDeleteBug(bug.id); }}
                      disabled={pendingDeleteBugId === bug.id}
                      className="text-[10px] font-bold text-red-600 bg-red-50 border border-red-100 px-2 py-1 rounded-lg hover:bg-red-100 disabled:opacity-60"
                    >
                      {pendingDeleteBugId === bug.id ? 'Удаляю…' : 'Удалить'}
                    </button>
                  )}
                </div>
              </div>
            ))
          ) : (
            <div className="text-center py-10 text-gray-400 italic">No bugs reported yet.</div>
          )}
        </div>

        <div className="p-6 border-t bg-white rounded-b-2xl">
          <button onClick={() => { setSelectedBugId(undefined); setIsEditorOpen(true); }}
            className="w-full py-3 border-2 border-dashed border-blue-200 text-blue-600 font-bold rounded-xl hover:bg-blue-50 transition-all">
            + Create New Bug
          </button>
        </div>
      </div>
    </div>
  );
}