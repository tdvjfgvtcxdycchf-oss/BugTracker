import React from 'react';

interface BugsModalProps {
  task: any;
  onClose: () => void;
  onAddBug: (bugData: any) => void; // Добавили этот тип
  setIsEditorOpen: (open: boolean) => void;
  setSelectedBugId: (id: string | undefined) => void;
}

export default function BugsModal({ task, onClose, setIsEditorOpen, setSelectedBugId }: BugsModalProps) {
  if (!task) return null;

  return (
    <div className="fixed inset-0 z-110 flex items-center justify-center bg-black/50 backdrop-blur-md p-4">
      <div className="bg-white w-full max-w-lg rounded-2xl shadow-2xl flex flex-col max-h-[80vh]">
        {/* Шапка модалки */}
        <div className="p-6 border-b flex justify-between items-center bg-gray-50 rounded-t-2xl">
        <div>
            <h2 className="text-xl font-black text-gray-900 line-clamp-1">{task.title}</h2>
            {/* ДОБАВЛЕНО ОПИСАНИЕ: */}
            <p className="text-sm text-gray-500 mt-1">{task.description}</p>
            <p className="text-[10px] text-blue-500 font-bold uppercase tracking-widest mt-2">Bugs Repository</p>
        </div>
        <button onClick={onClose} className="text-gray-400 hover:text-gray-900 text-3xl">&times;</button>
        </div>

        <div className="flex-1 overflow-y-auto p-6 space-y-3">
          {task.bugs && task.bugs.length > 0 ? (
            task.bugs.map((bug: any) => (
              <div 
                key={bug.id} 
                onClick={() => {
                  setSelectedBugId(bug.id); 
                  setIsEditorOpen(true); 
                }}
                className="p-4 bg-red-50 border border-red-100 rounded-xl flex items-center gap-3 cursor-pointer hover:border-red-400 transition-colors"
              >
                <div className="w-2 h-2 bg-red-500 rounded-full shadow-[0_0_5px_rgba(239,68,68,0.5)]"></div>
                <span className="text-sm font-bold text-gray-700">{bug.title}</span>
              </div>
            ))
          ) : (
            <div className="text-center py-10 text-gray-400 italic">No bugs reported yet.</div>
          )}
        </div>

        <div className="p-6 border-t bg-white rounded-b-2xl">
          <button 
            type="button" 
            onClick={() => {
              setSelectedBugId(undefined); 
              setIsEditorOpen(true);       
            }}
            className="w-full py-3 border-2 border-dashed border-blue-200 text-blue-600 font-bold rounded-xl hover:bg-blue-50 transition-all"
          >
            + Create New Bug
          </button>
        </div>
      </div>
    </div>
  );
}