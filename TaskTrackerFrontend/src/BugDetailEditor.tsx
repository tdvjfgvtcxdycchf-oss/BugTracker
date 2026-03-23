import React, { useState, useEffect } from 'react';

interface BugDetailEditorProps {
  isOpen: boolean;
  onClose: () => void;
  task: any;
  bugId?: string | number;
  onBugSaved: (updatedBugs: any[]) => void;
}

export default function BugDetailEditor({ isOpen, onClose, task, bugId, onBugSaved }: BugDetailEditorProps) {
  const [description, setDescription] = useState('');
  const [steps, setSteps] = useState('');
  const [expected, setExpected] = useState('');
  const [actual, setActual] = useState('');
  const [severity, setSeverity] = useState('Major');
  const [priority, setPriority] = useState('High');
  const [status, setStatus] = useState('New');
  const [version, setVersion] = useState('v1.0.0');
  const [selectedOS, setSelectedOS] = useState<string[]>(['Web']);
  const [currentBug, setCurrentBug] = useState<any>(null);

  const currentUserEmail = localStorage.getItem('userEmail') || 'Unknown User';
  const osOptions = ['Android', 'iOS', 'Web', 'macOS', 'Windows', 'Linux'];

  useEffect(() => {
    if (isOpen && task) {
      const bug = task.bugs?.find((b: any) => String(b.id) === String(bugId));
      
      if (bug) {
        setCurrentBug(bug);
        setDescription(bug.description || '');
        setSteps(bug.playback_description || '');
        setExpected(bug.expected_result || '');
        setActual(bug.actual_result || '');
        setSeverity(bug.severity || 'Major');
        setPriority(bug.priority || 'High');
        setStatus(bug.status || 'New');
        setVersion(bug.version_product || 'v1.0.0');
        setSelectedOS(bug.os ? bug.os.split(', ') : ['Web']);
      } else {
        setCurrentBug(null);
        setDescription(''); setSteps(''); setExpected(''); setActual('');
        setSeverity('Major'); setPriority('High'); setStatus('New');
        setVersion('v1.0.0'); setSelectedOS(['Web']);
      }
    }
  }, [bugId, isOpen, task]);

  const handleSave = async () => {
  // Проверка обязательного поля
  if (!description) return alert("Description is required");

  const payload = {
    task_id: task.id,
    severity,
    priority,
    status,
    os: selectedOS.join(', '),
    version_product: version,
    description,
    playback_description: steps,
    expected_result: expected,
    actual_result: actual,
    // Не забываем ID пользователя для бэкенда
    created_by: Number(localStorage.getItem('userId')) 
  };

  try {
    const baseUrl = (import.meta as any).env.VITE_API_URL;
    
    // Если bugId есть — PATCH на /bugs/{id}, если нет — POST на создание
    const url = bugId 
      ? `${baseUrl}/bugs/${bugId}` 
      : `${baseUrl}/bugs/${task.id}`;

    const method = bugId ? 'PATCH' : 'POST';

    const response = await fetch(url, {
      method: method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Ошибка при сохранении');
    }

    // Бэкенд теперь возвращает весь список багов задачи
    const updatedBugs = await response.json();
    
    if (typeof onBugSaved === 'function') {
      onBugSaved(updatedBugs);
      onClose();
    }
    } catch (err: any) {
      console.error("Save error:", err);
      alert(err.message);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[300] bg-slate-900/40 backdrop-blur-sm flex justify-center items-center p-4 text-slate-900">
      <div className="bg-white w-full max-w-5xl rounded-[2.5rem] shadow-2xl flex flex-col max-h-[95vh] border border-slate-200 overflow-hidden">
        
        {/* Header */}
        <div className="p-8 flex justify-between items-center border-b border-slate-100 bg-white">
          <div>
            <h1 className="text-3xl font-black tracking-tight">Technical Issue Report</h1>
            <p className="text-sm text-slate-400 mt-1 font-bold uppercase tracking-widest">Task: {task?.title || '—'}</p>
          </div>
          <div className="flex gap-4">
            <button onClick={onClose} className="px-6 py-2 text-slate-400 font-bold hover:text-slate-600 transition-colors">Cancel</button>
            <button onClick={handleSave} className="px-10 py-3 bg-blue-600 text-white rounded-2xl font-black shadow-lg hover:bg-blue-700 transition-all active:scale-95">
              {bugId ? 'Update Bug' : 'Create Bug'}
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-8 bg-[#F8FAFC]">
          <div className="grid grid-cols-3 gap-8">
            
            {/* Left Column (Inputs) */}
            <div className="col-span-2 space-y-6">
              <div>
                <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Detailed Description</label>
                <textarea className="w-full mt-2 p-4 border border-slate-200 rounded-2xl min-h-[120px] outline-none bg-white focus:ring-2 ring-blue-500/10 transition-all" 
                  value={description} onChange={e => setDescription(e.target.value)} placeholder="Summarize the core problem..." />
              </div>

              <div className="grid grid-cols-2 gap-6">
                <div>
                  <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Expected Result</label>
                  <textarea className="w-full mt-2 p-4 border border-green-100 bg-green-50/20 rounded-2xl min-h-[100px] outline-none" 
                    value={expected} onChange={e => setExpected(e.target.value)} />
                </div>
                <div>
                  <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Actual Result</label>
                  <textarea className="w-full mt-2 p-4 border border-red-100 bg-red-50/20 rounded-2xl min-h-[100px] outline-none" 
                    value={actual} onChange={e => setActual(e.target.value)} />
                </div>
              </div>

              <div>
                <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Steps to Reproduce</label>
                <textarea className="w-full mt-2 p-4 border border-slate-200 rounded-2xl min-h-[150px] font-mono text-sm outline-none bg-white focus:ring-2 ring-blue-500/10 transition-all" 
                  value={steps} onChange={e => setSteps(e.target.value)} />
              </div>
            </div>

            {/* Right Column (Controls) */}
            <div className="bg-white p-6 rounded-[2rem] border border-slate-100 shadow-sm space-y-6 h-fit">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Severity</label>
                  <select className="w-full p-3 bg-slate-50 rounded-xl font-bold border-none outline-none appearance-none cursor-pointer" 
                    value={severity} onChange={e => setSeverity(e.target.value)}>
                    <option>Critical</option><option>Major</option><option>Minor</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Priority</label>
                  <select className="w-full p-3 bg-slate-50 rounded-xl font-bold border-none outline-none appearance-none cursor-pointer" 
                    value={priority} onChange={e => setPriority(e.target.value)}>
                    <option>High</option><option>Medium</option><option>Low</option>
                  </select>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Target OS</label>
                <div className="flex flex-wrap gap-2 pt-1">
                  {osOptions.map(os => (
                    <button key={os} type="button" onClick={() => setSelectedOS(prev => prev.includes(os) ? prev.filter(x => x !== os) : [...prev, os])}
                      className={`px-3 py-1.5 rounded-lg text-[10px] font-bold transition-all ${selectedOS.includes(os) ? 'bg-blue-600 text-white shadow-md shadow-blue-200' : 'bg-slate-100 text-slate-400 hover:bg-slate-200'}`}>
                      {os}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Status</label>
                <select className="w-full p-3 bg-slate-50 rounded-xl font-bold border-none outline-none appearance-none cursor-pointer" 
                  value={status} onChange={e => setStatus(e.target.value)}>
                  <option>New</option><option>In Progress</option><option>Resolved</option><option>Closed</option>
                </select>
              </div>

              <div className="space-y-2">
                <label className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Version</label>
                <input type="text" className="w-full p-3 bg-slate-50 rounded-xl font-bold outline-none border-none focus:ring-2 ring-blue-500/10" 
                  value={version} onChange={e => setVersion(e.target.value)} />
              </div>
            </div>
          </div>

          {/* Lifecycle Audit Trail */}
          <div className="mt-12">
            <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">Lifecycle Audit Trail</label>
            <div className="mt-4 grid grid-cols-4 gap-4">
              <div className="bg-white border border-slate-100 p-5 rounded-3xl shadow-sm">
                <span className="text-xs font-black text-slate-900">Создан:</span>
                <div className="mt-2 text-sm text-slate-400 space-y-1">
                  <p>
                    {currentBug?.created_at 
                      ? new Date(currentBug.created_at).toLocaleDateString() 
                      : new Date().toLocaleDateString()}
                  </p>
                  <p>by <span className="text-slate-600 font-medium">{currentBug?.creator_email || currentUserEmail}</span></p>
                </div>
              </div>
              
              {['Закреплен за', 'Сдал', 'Принял'].map((label, i) => (
                <div key={i} className="bg-white border border-slate-100 p-5 rounded-3xl shadow-sm">
                  <span className="text-xs font-black text-slate-900">{label}:</span>
                  <div className="mt-2 text-sm text-slate-400 space-y-1">
                    <p>—</p>
                    <p>by <span className="text-slate-600 font-medium">—</span></p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}