import React, { useState, useEffect } from 'react';
import { API_URL } from './config';
import { apiFetch } from './api';

interface Bug {
  id?: number;
  task_id?: number;
  id_pk?: number;
  task_id_fk?: number;
  severity: string;
  priority: string;
  status: string;
  description: string;
  playback_description: string;
  expected_result: string;
  actual_result: string;
  version_product: string;
  os: string;
  created_by?: number;
  created_by_fk?: number;
  created_time?: string;
  assigned_to?: number | null;
  assigned_to_fk?: number | null;
  assigned_time?: string | null;
  passed_by?: number | null;
  passed_by_fk?: number | null;
  passed_time?: string | null;
  accepted_by?: number | null;
  accepted_by_fk?: number | null;
  accepted_time?: string | null;
}

interface User {
  id_pk?: number;
  id?: number;
  email: string;
}

interface Props {
  isOpen: boolean;
  onClose: () => void;
  task: { id: number };
  currentBug: Bug | null;
  onBugSaved: (data: any) => void;
  bugId?: string | number | null;
}

function AuditEntry({ label, email, date }: { label: string; email: string; date?: string | null }) {
  return (
    <div className="p-6 rounded-3xl border border-slate-100 bg-slate-50/50">
      <p className="text-[10px] font-black text-slate-400 uppercase mb-2">{label}</p>
      <p className="font-bold text-slate-900">{email}</p>
      {date && <p className="text-[10px] text-slate-400 mt-1">{date}</p>}
    </div>
  );
}

const BugDetailEditor: React.FC<Props> = ({ isOpen, onClose, task, currentBug, onBugSaved, bugId }) => {
  const isEditing = !!currentBug || !!bugId;

  const [users, setUsers] = useState<User[]>([]);
  const currentUserEmail = localStorage.getItem('userEmail') || 'Guest';
  const currentUserId = Number(localStorage.getItem('userId') || '0');
  const currentUserRole = localStorage.getItem('userRole') || 'qa';
  const [lifecyclePending, setLifecyclePending] = useState(false);

  // Comments
  const [comments, setComments] = useState<{ id_pk: number; user_id_fk: number; body: string; created_at: string }[]>([]);
  const [commentBody, setCommentBody] = useState('');
  const [commentPending, setCommentPending] = useState(false);


  const [photo, setPhoto] = useState<string | undefined>();
  const [photoName, setPhotoName] = useState('');
  const [lightbox, setLightbox] = useState(false);
  const [pendingPhotoFile, setPendingPhotoFile] = useState<File | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setPhotoName(file.name);
    setPendingPhotoFile(file);
    setPhoto(URL.createObjectURL(file));
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    const item = Array.from(e.clipboardData.items).find(i => i.type.startsWith('image/'));
    if (!item) return;
    const file = item.getAsFile();
    if (!file) return;
    setPhotoName('скриншот.png');
    setPendingPhotoFile(file);
    setPhoto(URL.createObjectURL(file));
  };

  const [severity, setSeverity] = useState('Low');
  const [priority, setPriority] = useState('Low');
  const [status, setStatus] = useState('Open');
  const [description, setDescription] = useState('');
  const [steps, setSteps] = useState('');
  const [expected, setExpected] = useState('');
  const [actual, setActual] = useState('');
  const [version, setVersion] = useState('');
  const [selectedOS, setSelectedOS] = useState<string[]>([]);
  const [assignableEmails, setAssignableEmails] = useState<string[]>([]);
  const [assignableEmailsPending, setAssignableEmailsPending] = useState(false);
  const [assignedToEmailChoice, setAssignedToEmailChoice] = useState('');

  useEffect(() => {
    if (isOpen && currentUserId > 0 && currentUserEmail) {
      setUsers(prev => {
        const exists = prev.some(u => u.id_pk === currentUserId || u.id === currentUserId);
        if (exists) return prev;
        return [...prev, { id_pk: currentUserId, email: currentUserEmail }];
      });
    }
  }, [isOpen, currentUserId, currentUserEmail]);

  useEffect(() => {
    if (!isOpen || !currentBug) return;

    const ids = [
      currentBug.created_by ?? currentBug.created_by_fk,
      currentBug.assigned_to ?? currentBug.assigned_to_fk,
      currentBug.passed_by ?? currentBug.passed_by_fk,
      currentBug.accepted_by ?? currentBug.accepted_by_fk,
    ].filter((x): x is number => typeof x === 'number' && x > 0);

    if (currentUserId > 0 && !ids.includes(currentUserId)) ids.push(currentUserId);
    if (ids.length === 0) return;

    let cancelled = false;

    (async () => {
      try {
        const results = await Promise.all(
          ids.map(async (id) => {
            const res = await apiFetch(`${API_URL}/users/${id}`);
            const emails = (await res.json()) as string[];
            return { id, emails };
          })
        );

        const union = new Set<string>();
        const excludedById = new Map<number, Set<string>>();

        for (const { id, emails } of results) {
          const excluded = new Set(emails);
          excludedById.set(id, excluded);
          emails.forEach((e) => union.add(e));
        }

        const resolvedUsers: User[] = ids.map((id) => {
          const excluded = excludedById.get(id) ?? new Set<string>();
          const missing = [...union].find((e) => !excluded.has(e));
          return { id_pk: id, email: missing ?? `User #${id}` };
        });

        if (!cancelled) setUsers(resolvedUsers);
      } catch (err) {
        console.error(err);
      }
    })();

    return () => { cancelled = true; };
  }, [isOpen, currentBug]);

  useEffect(() => {
    if (!isOpen || !currentBug) return;

    const creatorId = currentBug.created_by ?? currentBug.created_by_fk;
    const assignedId = currentBug.assigned_to ?? currentBug.assigned_to_fk;

    if (!creatorId || assignedId != null || currentUserId !== creatorId) return;

    let cancelled = false;
    setAssignableEmailsPending(true);

    (async () => {
      try {
        const res = await apiFetch(`${API_URL}/users/${creatorId}`);
        const emails = (await res.json()) as string[];
        if (!cancelled) {
          setAssignableEmails(Array.isArray(emails) ? emails : []);
          setAssignedToEmailChoice('');
        }
      } catch (err) {
        console.error(err);
      } finally {
        if (!cancelled) setAssignableEmailsPending(false);
      }
    })();

    return () => { cancelled = true; };
  }, [isOpen, currentBug, currentUserId]);

  useEffect(() => {
    setPendingPhotoFile(null);
    if (isOpen && currentBug) {
      const id = currentBug.id ?? currentBug.id_pk;
      if (!id) return;
      let cancelled = false;
      (async () => {
        try {
          const res = await apiFetch(`${API_URL}/bugs/${id}/photo?t=${Date.now()}`);
          if (!res.ok) return;
          const blob = await res.blob();
          const url = URL.createObjectURL(blob);
          if (!cancelled) {
            setPhoto(url);
            setPhotoName('фото');
          } else {
            URL.revokeObjectURL(url);
          }
        } catch (e) {
          // ignore
        }
      })();
      return () => { cancelled = true; };
    } else {
      setPhoto(undefined);
      setPhotoName('');
    }
  }, [isOpen, currentBug]);

  useEffect(() => {
    if (isOpen && currentBug) {
      setSeverity(currentBug.severity || 'Low');
      setPriority(currentBug.priority || 'Low');
      setStatus(currentBug.status || 'Open');
      setDescription(currentBug.description || '');
      setSteps(currentBug.playback_description || '');
      setExpected(currentBug.expected_result || '');
      setActual(currentBug.actual_result || '');
      setVersion(currentBug.version_product || '');
      setSelectedOS(currentBug.os ? currentBug.os.split(', ') : []);
    }
  }, [currentBug, isOpen]);

  useEffect(() => {
    if (!isOpen || !currentBug) { setComments([]); return; }
    const id = currentBug.id ?? currentBug.id_pk;
    if (!id) return;
    let cancelled = false;
    (async () => {
      try {
        const cRes = await apiFetch(`${API_URL}/bugs/${id}/comments`);
        const cData = await cRes.json().catch(() => []);
        if (!cancelled) setComments(Array.isArray(cData) ? cData : []);
      } catch (err) { console.error(err); }
    })();
    return () => { cancelled = true; };
  }, [isOpen, currentBug]);


  const handleAddComment = async () => {
    const trimmed = commentBody.trim();
    if (!trimmed) return;
    const id = currentBug?.id ?? currentBug?.id_pk;
    if (!id) return;
    setCommentPending(true);
    try {
      const res = await apiFetch(`${API_URL}/bugs/${id}/comments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ body: trimmed }),
      });
      if (res.ok) {
        const newComment = await res.json();
        setComments(prev => [...prev, newComment]);
        setCommentBody('');
      }
    } catch (err) { console.error(err); }
    finally { setCommentPending(false); }
  };

  const getUserEmail = (id?: number) => {
    if (!id) return '—';
    const found = users.find(u => u.id_pk === id || u.id === id);
    return found ? found.email : `User #${id}`;
  };

  const getBugId = () => currentBug?.id ?? currentBug?.id_pk ?? bugId;
  const getCreatedById = () => currentBug?.created_by ?? currentBug?.created_by_fk;
  const getAssignedToId = () => currentBug?.assigned_to ?? currentBug?.assigned_to_fk;
  const getPassedById = () => currentBug?.passed_by ?? currentBug?.passed_by_fk;
  const getAcceptedById = () => currentBug?.accepted_by ?? currentBug?.accepted_by_fk;

  const createdTime = currentBug?.created_time;
  const assignedTime = currentBug?.assigned_time;
  const passedTime = currentBug?.passed_time;
  const acceptedTime = currentBug?.accepted_time;

  const formatDate = (value?: string | null) => {
    if (!value) return null;
    const d = new Date(value);
    if (Number.isNaN(d.getTime())) return null;
    return d.toLocaleDateString();
  };

  const refreshBugs = async () => {
    const res = await apiFetch(`${API_URL}/bugs/${task.id}`);
    const updatedBugs = await res.json().catch(() => []);
    onBugSaved(updatedBugs);
  };

  const handleSave = async (opts?: { closeOnSuccess?: boolean }) => {
    const userId = Number(localStorage.getItem('userId'));
    const payload: any = {
      created_by: userId,
      severity, priority, status, description,
      playback_description: steps,
      expected_result: expected,
      actual_result: actual,
      version_product: version,
      os: selectedOS.join(', '),
      task_id: task.id,
      assigned_to_email: assignedToEmailChoice || ''
    };

    if (isEditing && currentBug) {
      payload.assigned_to = getAssignedToId() ?? null;
      payload.assigned_time = assignedTime ?? null;
      payload.passed_by = getPassedById() ?? null;
      payload.passed_time = passedTime ?? null;
      payload.accepted_by = getAcceptedById() ?? null;
      payload.accepted_time = acceptedTime ?? null;
    }

    try {
      const url = isEditing
        ? `${API_URL}/bugs/${getBugId()}`
        : `${API_URL}/bugs/${task?.id}`;

      const res = await apiFetch(url, {
        method: isEditing ? 'PATCH' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        await refreshBugs();
        if (pendingPhotoFile) {
          const bugsRes = await apiFetch(`${API_URL}/bugs/${task.id}`);
          const bugs = await bugsRes.json().catch(() => []);
          const targetId = isEditing ? getBugId() : Math.max(...bugs.map((b: any) => b.id));
          if (targetId) {
            const form = new FormData();
            form.append('photo', pendingPhotoFile);
            await apiFetch(`${API_URL}/bugs/${targetId}/photo`, { method: 'POST', body: form });
            setPendingPhotoFile(null);
          }
        }
        if (opts?.closeOnSuccess !== false) onClose();
      }
    } catch (err) { console.error(err); }
  };

  const handleStatusChange = async (newStatus: string) => {
    const bugPk = getBugId();
    if (!bugPk) return;
    setLifecyclePending(true);
    try {
      const payload: any = {
        task_id: task.id,
        severity, priority, status: newStatus, description,
        playback_description: steps,
        expected_result: expected,
        actual_result: actual,
        version_product: version,
        os: selectedOS.join(', '),
        assigned_to: getAssignedToId() ?? null,
        assigned_time: assignedTime ?? null,
        passed_by: getPassedById() ?? null,
        passed_time: passedTime ?? null,
        accepted_by: getAcceptedById() ?? null,
        accepted_time: acceptedTime ?? null,
      };
      const res = await apiFetch(`${API_URL}/bugs/${bugPk}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (res.ok) { setStatus(newStatus); await refreshBugs(); }
    } catch (err) { console.error(err); }
    finally { setLifecyclePending(false); }
  };

  const handleLifecycle = async (type: 'pass' | 'accept') => {
    const bugPk = getBugId();
    if (!bugPk) return;

    setLifecyclePending(true);
    try {
      const nowIso = new Date().toISOString();
      const payload: any = {
        task_id: task.id,
        severity, priority, status, description,
        playback_description: steps,
        expected_result: expected,
        actual_result: actual,
        version_product: version,
        os: selectedOS.join(', '),
        assigned_to: getAssignedToId() ?? null,
        assigned_time: assignedTime ?? null,
        passed_by: type === 'pass' ? currentUserId : (getPassedById() ?? null),
        passed_time: type === 'pass' ? nowIso : (passedTime ?? null),
        accepted_by: type === 'accept' ? currentUserId : (getAcceptedById() ?? null),
        accepted_time: type === 'accept' ? nowIso : (acceptedTime ?? null),
      };

      const res = await apiFetch(`${API_URL}/bugs/${bugPk}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) await refreshBugs();
    } catch (err) {
      console.error(err);
    } finally {
      setLifecyclePending(false);
    }
  };

  const createdById = getCreatedById();
  const assignedToId = getAssignedToId();
  const passedById = getPassedById();
  const acceptedById = getAcceptedById();

  const isCreator = createdById != null && currentUserId === createdById;
  const isAssignee = assignedToId != null && currentUserId === assignedToId;
  const canPass = isAssignee && passedById == null;
  const canAccept = isCreator && passedById != null && acceptedById == null;

  // Role-based workflow
  const isQA = currentUserRole === 'qa' || currentUserRole === 'admin';
  const isDev = currentUserRole === 'developer' || currentUserRole === 'admin';
  const canReopen = isEditing && isQA && (status === 'Fixed' || status === 'Ready for Retest' || status === 'Verified');
  const canReject = isEditing && isDev && status !== 'Rejected' && status !== 'Can\'t Reproduce' && status !== 'Verified';
  const canMarkFixed = isEditing && isDev && (status === 'Open' || status === 'In Progress' || status === 'Reopened');

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-slate-900/40 backdrop-blur-md flex items-center justify-center p-4 z-[1000]">
      <div className="bg-white w-full max-w-[900px] max-h-[90vh] rounded-2xl sm:rounded-[32px] shadow-2xl overflow-hidden flex flex-col p-4 sm:p-10">

        <h1 className="text-2xl sm:text-3xl font-black text-slate-900 mb-4 sm:mb-8">
          {isEditing ? `Редактировать баг #${currentBug?.id}` : 'Создать баг'}
        </h1>

        <div className="flex-1 overflow-y-auto pr-2 space-y-6">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Серьёзность</label>
              <select value={severity} onChange={e => setSeverity(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none">
                <option value="Blocker">Blocker</option>
                <option value="Critical">Critical</option>
                <option value="Major">Major</option>
                <option value="Minor">Minor</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Приоритет</label>
              <select value={priority} onChange={e => setPriority(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none">
                <option value="High">High</option>
                <option value="Medium">Medium</option>
                <option value="Low">Low</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Статус</label>
              <select value={status} onChange={e => setStatus(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none">
                <option value="New">New</option>
                <option value="Open">Open</option>
                <option value="In Progress">In Progress</option>
                <option value="Fixed">Fixed</option>
                <option value="Ready for Retest">Ready for Retest</option>
                <option value="Verified">Verified</option>
                <option value="Reopened">Reopened</option>
                <option value="Rejected">Rejected</option>
                <option value="Can't Reproduce">Can't Reproduce</option>
              </select>
            </div>
          </div>

          {isCreator && assignedToId == null && (
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Закрепить за (email)</label>
              {assignableEmailsPending ? (
                <div className="text-sm text-slate-500">Загрузка...</div>
              ) : (
                <select
                  value={assignedToEmailChoice}
                  onChange={(e) => setAssignedToEmailChoice(e.target.value)}
                  className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none"
                >
                  <option value="">Выберите email</option>
                  {assignableEmails.map((e) => (
                    <option key={e} value={e}>{e}</option>
                  ))}
                </select>
              )}
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Описание</label>
              <textarea value={description} onChange={e => setDescription(e.target.value)} className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[120px] outline-none" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Шаги воспроизведения</label>
              <textarea value={steps} onChange={e => setSteps(e.target.value)} className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[120px] outline-none" />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Ожидаемый результат</label>
              <textarea value={expected} onChange={e => setExpected(e.target.value)} className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[80px] outline-none" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Фактический результат</label>
              <textarea value={actual} onChange={e => setActual(e.target.value)} className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[80px] outline-none" />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6 items-end">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Версия</label>
              <input value={version} onChange={e => setVersion(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none" placeholder="1.0.0" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">ОС</label>
              <div className="flex flex-wrap gap-2">
                {['Win', 'Mac', 'Linux', 'iOS', 'Android'].map(os => (
                  <button key={os} onClick={() => setSelectedOS(prev => prev.includes(os) ? prev.filter(o => o !== os) : [...prev, os])}
                    className={`px-4 py-2 rounded-lg text-xs font-bold border transition-all ${selectedOS.includes(os) ? 'text-white border-[#7C5CBF]' : 'bg-white text-slate-600 border-slate-100'}`} style={selectedOS.includes(os) ? { background: '#7C5CBF' } : {}}>
                    {os}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="space-y-2" onPaste={handlePaste}>
            <label className="text-xs font-bold text-slate-900">Фото</label>
            <label className="flex items-center gap-3 cursor-pointer w-full p-3 rounded-xl border border-dashed border-slate-200 bg-slate-50 hover:border-blue-400 transition-colors">
              <span className="text-lg">📎</span>
              <span className="text-sm text-slate-500 truncate">{photoName || 'Прикрепить или вставить (Ctrl+V)...'}</span>
              <input type="file" accept="image/*" className="hidden" onChange={handleFileChange} />
            </label>
            {photo && (
              <div className="relative">
                <img
                  src={photo}
                  alt="скриншот"
                  className="w-full max-h-48 object-cover rounded-xl border border-slate-100 cursor-zoom-in"
                  onClick={() => setLightbox(true)}
                  onError={() => { setPhoto(undefined); setPhotoName(''); }}
                />
              </div>
            )}
          </div>

          {lightbox && photo && (
            <div
              className="fixed inset-0 z-[2000] bg-black/80 flex items-center justify-center p-4"
              onClick={() => setLightbox(false)}
            >
              <img src={photo} alt="full" className="max-w-full max-h-full rounded-xl shadow-2xl" />
            </div>
          )}

          <div className="pt-6 border-t border-slate-50">
            <h2 className="text-xl font-black text-slate-900 mb-6">История жизненного цикла</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-6">
              <AuditEntry
                label="СОЗДАН"
                email={createdById === currentUserId ? currentUserEmail : getUserEmail(createdById)}
                date={isEditing ? formatDate(createdTime) : null}
              />
              <AuditEntry
                label="ЗАКРЕПЛЕН ЗА"
                email={assignedToId != null ? getUserEmail(assignedToId) : '—'}
                date={formatDate(assignedTime)}
              />
              <AuditEntry
                label="СДАЛ"
                email={passedById != null ? getUserEmail(passedById) : '—'}
                date={formatDate(passedTime)}
              />
              <AuditEntry
                label="ПРИНЯЛ"
                email={acceptedById != null ? getUserEmail(acceptedById) : '—'}
                date={formatDate(acceptedTime)}
              />
            </div>
          </div>


          {isEditing && (
            <div className="pt-6 border-t border-slate-50">
              <h2 className="text-xl font-black text-slate-900 mb-4">Комментарии</h2>
              <div className="space-y-3 mb-4">
                {comments.length === 0 && (
                  <p className="text-sm text-slate-400">Нет комментариев</p>
                )}
                {comments.map(c => (
                  <div key={c.id_pk} className="p-4 rounded-2xl bg-slate-50 border border-slate-100">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs font-bold text-slate-500">{getUserEmail(c.user_id_fk)}</span>
                      <span className="ml-auto text-[10px] text-slate-400">{new Date(c.created_at).toLocaleString()}</span>
                    </div>
                    <p className="text-sm text-slate-800 whitespace-pre-wrap">{c.body}</p>
                  </div>
                ))}
              </div>
              <div className="flex gap-3">
                <textarea
                  value={commentBody}
                  onChange={e => setCommentBody(e.target.value)}
                  placeholder="Написать комментарий..."
                  className="flex-1 p-3 rounded-xl border border-slate-200 bg-white text-sm outline-none resize-none min-h-[60px]"
                  onKeyDown={e => { if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) handleAddComment(); }}
                />
                <button
                  type="button"
                  disabled={commentPending || !commentBody.trim()}
                  onClick={handleAddComment}
                  className="self-end text-white px-5 py-3 rounded-2xl font-bold text-sm disabled:opacity-50 transition-colors" style={{ background: '#7C5CBF' }}
                >
                  {commentPending ? '...' : 'Отправить'}
                </button>
              </div>
            </div>
          )}
        </div>

        <div className="mt-4 sm:mt-8 flex flex-wrap justify-end items-center gap-3 sm:gap-6">
          <button onClick={onClose} className="text-sm font-bold text-slate-500 uppercase">Отмена</button>
          {isCreator && assignedToId == null && (
            <button
              type="button"
              disabled={lifecyclePending || !assignedToEmailChoice}
              onClick={() => handleSave({ closeOnSuccess: false })}
              className="bg-slate-900 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-slate-200 hover:scale-[1.02] transition-all disabled:opacity-60"
            >
              Закрепить
            </button>
          )}
          {canMarkFixed && (
            <button type="button" disabled={lifecyclePending}
              onClick={() => handleStatusChange('Fixed')}
              className="bg-emerald-600 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-emerald-100 hover:scale-[1.02] transition-all disabled:opacity-60">
              Fixed
            </button>
          )}
          {canReject && (
            <button type="button" disabled={lifecyclePending}
              onClick={() => handleStatusChange('Rejected')}
              className="bg-orange-500 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-orange-100 hover:scale-[1.02] transition-all disabled:opacity-60">
              Reject
            </button>
          )}
          {canReject && (
            <button type="button" disabled={lifecyclePending}
              onClick={() => handleStatusChange("Can't Reproduce")}
              className="bg-yellow-500 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-yellow-100 hover:scale-[1.02] transition-all disabled:opacity-60">
              Can't Reproduce
            </button>
          )}
          {canReopen && (
            <button type="button" disabled={lifecyclePending}
              onClick={() => handleStatusChange('Reopened')}
              className="bg-red-600 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-red-100 hover:scale-[1.02] transition-all disabled:opacity-60">
              Reopen
            </button>
          )}
          {canPass && (
            <button
              type="button"
              disabled={lifecyclePending}
              onClick={() => handleLifecycle('pass')}
              className="bg-emerald-600 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-emerald-100 hover:scale-[1.02] transition-all disabled:opacity-60"
            >
              Сдать
            </button>
          )}
          {canAccept && (
            <button
              type="button"
              disabled={lifecyclePending}
              onClick={() => handleLifecycle('accept')}
              className="bg-purple-600 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-purple-100 hover:scale-[1.02] transition-all disabled:opacity-60"
            >
              Принять
            </button>
          )}
          <button onClick={() => handleSave()} className="bg-blue-600 text-white px-6 sm:px-10 py-3 sm:py-4 rounded-3xl font-black shadow-lg shadow-blue-100 hover:scale-[1.02] transition-all">
            {isEditing ? 'Обновить баг' : 'Создать баг'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default BugDetailEditor;
