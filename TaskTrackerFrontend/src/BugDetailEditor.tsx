import React, { useState, useEffect } from 'react';

interface BugDetailEditorProps {
  isOpen: boolean;
  onClose: () => void;
  task: any;
  bugId?: string;
  onCreateBug: (bugData: any) => void;
  onUpdateBug: (bugData: any) => void;
}

export default function BugDetailEditor({ 
  isOpen, onClose, task, bugId, onCreateBug, onUpdateBug 
}: BugDetailEditorProps) {
  
  // Состояния полей
  const [displayId, setDisplayId] = useState(''); // Визуальный ID (например, SL-2024-01)
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [steps, setSteps] = useState('');
  const [expected, setExpected] = useState('');
  const [actual, setActual] = useState('');
  const [severity, setSeverity] = useState('Major');
  const [priority, setPriority] = useState('High');
  const [status, setStatus] = useState('New');
  const [version, setVersion] = useState('v2.4.1-beta');
  const [platforms, setPlatforms] = useState<string[]>(['Web']);
  const [showUserList, setShowUserList] = useState(false);

  const users = ['Marcus Thorne', 'Alex Chen', 'Sarah Jenkins', 'Leo Rodriguez', 'Anna Smith'];

  const [audit, setAudit] = useState({
    createdBy: 'Alex Chen',
    createdAt: '14.05.2024',
    assignedTo: 'Not Assigned',
    submittedBy: 'Sarah Jenkins',
    acceptedBy: '—'
  });

  // Функция генерации случайного ID для новых багов
  const generateBugId = () => {
    const randomNum = Math.floor(1000 + Math.random() * 9000);
    const suffix = Math.random().toString(36).substring(7).toUpperCase();
    return `SL-${randomNum}-${suffix}`;
  };

  useEffect(() => {
    if (bugId && task?.bugs) {
      const bug = task.bugs.find((b: any) => b.id === bugId);
      if (bug) {
        setDisplayId(bug.id || bugId); // Используем существующий ID
        setTitle(bug.title || '');
        setDescription(bug.description || '');
        setSteps(bug.steps || '');
        setExpected(bug.expected || '');
        setActual(bug.actual || '');
        setSeverity(bug.severity || 'Major');
        setPriority(bug.priority || 'High');
        setPlatforms(bug.platforms || ['Web']);
        setVersion(bug.version || 'v2.4.1-beta');
        setStatus(bug.status || 'New');
        if (bug.audit) setAudit(bug.audit);
      }
    } else {
      // Для нового бага генерируем временный ID для отображения
      setDisplayId(generateBugId());
      setTitle(''); setDescription(''); setSteps(''); setExpected(''); setActual('');
      setPlatforms(['Web']);
      setAudit({
        createdBy: 'Alex Chen',
        createdAt: new Date().toLocaleDateString(),
        assignedTo: 'Not Assigned',
        submittedBy: '—',
        acceptedBy: '—'
      });
    }
  }, [bugId, task, isOpen]);

  const togglePlatform = (p: string) => {
    setPlatforms(prev => prev.includes(p) ? prev.filter(item => item !== p) : [...prev, p]);
  };

  const handleAssignUser = (userName: string) => {
    setAudit(prev => ({ ...prev, assignedTo: userName }));
    setShowUserList(false);
  };

  const handleAccept = () => {
    setAudit(prev => ({ ...prev, acceptedBy: 'Current User' }));
    setStatus('In Progress');
  };

  const handleSave = () => {
    if (!title) return alert("Title is required");
    
    const bugData = { 
      id: displayId, // Сохраняем сгенерированный или старый ID
      title, description, steps, expected, actual, 
      severity, priority, platforms, version, status,
      audit 
    };
    
    if (bugId) onUpdateBug(bugData);
    else onCreateBug(bugData);
    
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-150 bg-slate-200/60 backdrop-blur-sm flex justify-center items-center p-4">
      <div className="bg-[#F8FAFC] w-full max-w-5xl rounded-3xl shadow-2xl border border-white flex flex-col max-h-[95vh] relative">
        
        {/* HEADER */}
        <div className="p-8 flex justify-between items-start">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <span className="bg-blue-600 text-white text-[10px] font-black px-2 py-0.5 rounded-md uppercase">
                {displayId}
              </span>
              <span className="bg-slate-200 text-slate-600 text-[10px] font-black px-2 py-0.5 rounded-md uppercase">
                {status}
              </span>
              <h1 className="text-2xl font-black text-[#0F172A]">Technical Issue Report</h1>
            </div>
            <p className="text-gray-400 text-sm">Task Context: {task?.name || 'Project Tracking'}</p>
          </div>
          
          <div className="flex gap-3 relative">
            <div className="relative">
              <button 
                onClick={() => setShowUserList(!showUserList)}
                className="px-4 py-2 border-2 border-blue-600 text-blue-600 rounded-xl font-bold text-sm hover:bg-blue-50 transition-all flex items-center gap-2"
              >
                Assign
              </button>
              {showUserList && (
                <div className="absolute top-12 left-0 w-48 bg-white border border-gray-100 rounded-xl shadow-xl z-160 py-2">
                  {users.map(u => (
                    <button key={u} onClick={() => handleAssignUser(u)} className="w-full text-left px-4 py-2 text-xs font-bold text-gray-600 hover:bg-blue-50 transition-colors">
                      {u}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <button onClick={handleAccept} className="px-4 py-2 bg-blue-100 text-blue-700 rounded-xl font-bold text-sm hover:bg-blue-200 transition-all">
              Accept
            </button>

            <button onClick={handleSave} className="px-6 py-2 bg-blue-600 text-white rounded-xl font-bold text-sm hover:bg-blue-700 shadow-lg transition-all">
              {bugId ? 'Update' : 'Create'}
            </button>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-2xl px-2">&times;</button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-8 pb-8">
          <div className="flex gap-6">
            
            {/* LEFT COLUMN */}
            <div className="flex-2 space-y-6">
              <div className="bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
                <label className="block text-[10px] font-black text-gray-400 uppercase tracking-widest mb-2">Issue Title</label>
                <input 
                  className="w-full bg-gray-50 border-none rounded-xl px-4 py-3 text-gray-700 outline-none focus:ring-2 focus:ring-blue-100 font-bold"
                  value={title} onChange={(e) => setTitle(e.target.value)}
                  placeholder="Summarize the technical problem..."
                />
              </div>

              <div className="bg-[#EFF6FF] p-8 rounded-3xl space-y-6 border border-blue-50">
                <div>
                  <h3 className="text-blue-900 font-black text-xs uppercase mb-3 flex items-center gap-2">Description</h3>
                  <textarea className="w-full bg-white rounded-xl p-4 min-h-80px outline-none text-sm text-gray-600 border-none shadow-inner resize-none" value={description} onChange={(e)=>setDescription(e.target.value)} />
                </div>
                <div>
                  <h3 className="text-blue-900 font-black text-xs uppercase mb-3 flex items-center gap-2">Steps to Reproduce</h3>
                  <textarea className="w-full bg-white rounded-xl p-4 min-h-80px outline-none text-sm text-gray-600 border-none shadow-inner resize-none" value={steps} onChange={(e)=>setSteps(e.target.value)} />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <h3 className="text-green-700 font-black text-[10px] uppercase mb-2">Expected Result</h3>
                    <textarea className="w-full bg-white rounded-xl p-3 min-h-60px outline-none text-xs text-gray-500 border-none resize-none" value={expected} onChange={(e)=>setExpected(e.target.value)} />
                  </div>
                  <div>
                    <h3 className="text-red-700 font-black text-[10px] uppercase mb-2">Actual Result</h3>
                    <textarea className="w-full bg-white rounded-xl p-3 min-h-60px outline-none text-xs text-gray-500 border-none resize-none" value={actual} onChange={(e)=>setActual(e.target.value)} />
                  </div>
                </div>
              </div>

              {/* AUDIT TRAIL */}
              <div className="p-4 border-t border-gray-100 mt-4">
                <h3 className="text-blue-900 font-black text-[11px] uppercase tracking-widest mb-6 flex items-center gap-2">⌘ Audit Trail</h3>
                <div className="grid grid-cols-4 gap-6">
                  <div>
                    <div className="text-[10px] font-black text-gray-400 uppercase mb-1">Created By</div>
                    <div className="text-sm font-bold text-gray-700">{audit.createdBy}</div>
                    <div className="text-[10px] text-gray-400 font-bold">{audit.createdAt}</div>
                  </div>
                  <div>
                    <div className="text-[10px] font-black text-blue-600 uppercase mb-1">Assigned To</div>
                    <div className="text-sm font-bold text-blue-700 underline decoration-2 underline-offset-4">{audit.assignedTo}</div>
                  </div>
                  <div>
                    <div className="text-[10px] font-black text-gray-400 uppercase mb-1">Submitted By</div>
                    <div className="text-sm font-bold text-gray-700">{audit.submittedBy}</div>
                  </div>
                  <div>
                    <div className="text-[10px] font-black text-gray-400 uppercase mb-1">Accepted By</div>
                    <div className="text-sm font-bold text-gray-700">{audit.acceptedBy}</div>
                  </div>
                </div>
              </div>
            </div>

            {/* RIGHT COLUMN */}
            <div className="flex-1 space-y-4">
              <div className="bg-[#E2E8F0]/40 p-6 rounded-2xl border border-white space-y-6">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-[10px] font-black text-gray-400 uppercase mb-2">Severity</label>
                    <select className="w-full bg-white rounded-lg px-2 py-2 text-[11px] font-bold text-red-600 outline-none shadow-sm" value={severity} onChange={(e)=>setSeverity(e.target.value)}>
                      <option>Major</option><option>Minor</option><option>Critical</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-[10px] font-black text-gray-400 uppercase mb-2">Priority</label>
                    <select className="w-full bg-white rounded-lg px-2 py-2 text-[11px] font-bold text-blue-600 outline-none shadow-sm" value={priority} onChange={(e)=>setPriority(e.target.value)}>
                      <option>High</option><option>Medium</option><option>Low</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-[10px] font-black text-gray-400 uppercase mb-3">Target Platform</label>
                  <div className="flex flex-wrap gap-2">
                    {['Android', 'iOS', 'Web', 'macOS', 'Windows', 'Linux'].map(p => (
                      <button 
                        key={p} onClick={() => togglePlatform(p)}
                        className={`px-3 py-1.5 rounded-full text-[10px] font-black transition-all ${platforms.includes(p) ? 'bg-blue-700 text-white shadow-md' : 'bg-blue-100 text-blue-700 hover:bg-blue-200'}`}
                      >
                        {p}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-[10px] font-black text-gray-400 uppercase mb-2">Version</label>
                    <input className="w-full bg-white rounded-lg px-3 py-2 text-[11px] font-bold text-gray-700 outline-none" value={version} onChange={(e)=>setVersion(e.target.value)} />
                  </div>
                  <div>
                    <label className="block text-[10px] font-black text-gray-400 uppercase mb-2">Status</label>
                    <select className="w-full bg-white rounded-lg px-2 py-2 text-[11px] font-bold text-gray-700 outline-none shadow-sm" value={status} onChange={(e)=>setStatus(e.target.value)}>
                      <option>New</option><option>In Progress</option><option>Resolved</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}