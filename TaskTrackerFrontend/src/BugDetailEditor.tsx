import React, { useState, useEffect } from 'react';

// Интерфейсы данных
interface Bug {
  // Основные поля из API (json tags в Go бэкенде)
  id?: number;
  task_id?: number;

  // Старые/альтернативные имена, которые могут встречаться в уже загруженных данных
  id_pk?: number;
  task_id_fk?: number;
  severity: string;
  priority: string;
  status: string;
  description: string;
  playback_description: string;
  expected_result: string; // Ожидаемый результат
  actual_result: string;   // Фактический результат
  version_product: string;
  os: string;
  // Кто создал
  created_by?: number;
  created_by_fk?: number;
  created_time?: string;

  // Кто закреплён
  assigned_to?: number | null;
  assigned_to_fk?: number | null;
  assigned_time?: string | null;

  // Кто сдал
  passed_by?: number | null;
  passed_by_fk?: number | null;
  passed_time?: string | null;

  // Кто принял
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

const BugDetailEditor: React.FC<Props> = ({ isOpen, onClose, task, currentBug, onBugSaved, bugId }) => {
  const isEditing = !!currentBug || !!bugId;
  
  // Состояния
  const [users, setUsers] = useState<User[]>([]);
  const currentUserEmail = localStorage.getItem('userEmail') || 'Guest';
  const currentUserId = Number(localStorage.getItem('userId') || '0');
  const [lifecyclePending, setLifecyclePending] = useState(false);
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

  const [severity, setSeverity] = useState('Low');
  const [priority, setPriority] = useState('Low');
  const [status, setStatus] = useState('Open');
  const [description, setDescription] = useState('');
  const [steps, setSteps] = useState('');
  const [expected, setExpected] = useState(''); // Состояние для ожидаемого
  const [actual, setActual] = useState('');     // Состояние для фактического
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

  // Разрешаем id -> email для жизненного цикла.
  // В бэкенде нет endpoint "GET /users" (список всех), зато есть "GET /users/{id}"
  // которое возвращает все email, кроме текущего id. Мы реконструируем нужный email как
  // "тот, которого нет в списке исключённых".
  useEffect(() => {
    if (!isOpen || !currentBug) return;

    const baseUrl = (import.meta as any).env.VITE_API_URL;

    const ids = [
      currentBug.created_by ?? currentBug.created_by_fk,
      currentBug.assigned_to ?? currentBug.assigned_to_fk,
      currentBug.passed_by ?? currentBug.passed_by_fk,
      currentBug.accepted_by ?? currentBug.accepted_by_fk,
    ].filter((x): x is number => typeof x === 'number' && x > 0);

    // Чтобы реконструкция id->email работала даже когда есть только один id,
    // добавим текущего пользователя (если он не в списке).
    if (currentUserId > 0 && !ids.includes(currentUserId)) ids.push(currentUserId);

    if (ids.length === 0) return;

    let cancelled = false;

    (async () => {
      try {
        const results = await Promise.all(
          ids.map(async (id) => {
            const res = await fetch(`${baseUrl}/users/${id}`);
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
          return {
            id_pk: id,
            email: missing ?? `User #${id}`,
          };
        });

        if (!cancelled) setUsers(resolvedUsers);
      } catch (err) {
        console.error('Ошибка резолвинга email:', err);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [isOpen, currentBug]);

  // Список email для "закрепить за" (только создателю и только если ещё не закреплено).
  useEffect(() => {
    if (!isOpen || !currentBug) return;

    const creatorId = currentBug.created_by ?? currentBug.created_by_fk;
    const assignedId = currentBug.assigned_to ?? currentBug.assigned_to_fk;

    if (!creatorId) return;
    if (assignedId == null) {
      // ok, ещё не закреплено
    } else {
      return;
    }

    if (currentUserId !== creatorId) return;

    const baseUrl = (import.meta as any).env.VITE_API_URL;

    let cancelled = false;
    setAssignableEmailsPending(true);

    (async () => {
      try {
        const res = await fetch(`${baseUrl}/users/${creatorId}`);
        const emails = (await res.json()) as string[];
        if (!cancelled) {
          setAssignableEmails(Array.isArray(emails) ? emails : []);
          setAssignedToEmailChoice('');
        }
      } catch (err) {
        console.error('Ошибка загрузки email для закрепления:', err);
      } finally {
        if (!cancelled) setAssignableEmailsPending(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [isOpen, currentBug, currentUserId]);

  // Загружаем фото с сервера при открытии бага
  useEffect(() => {
    setPendingPhotoFile(null);
    if (isOpen && currentBug) {
      const baseUrl = (import.meta as any).env.VITE_API_URL;
      const id = currentBug.id ?? currentBug.id_pk;
      setPhoto(`${baseUrl}/bugs/${id}/photo?t=${Date.now()}`);
      setPhotoName('фото');
    } else {
      setPhoto(undefined);
      setPhotoName('');
    }
  }, [isOpen, currentBug]);

  // Заполнение полей данными
  useEffect(() => {
    if (isOpen && currentBug) {
      setSeverity(currentBug.severity || 'Low');
      setPriority(currentBug.priority || 'Low');
      setStatus(currentBug.status || 'Open');
      setDescription(currentBug.description || '');
      setSteps(currentBug.playback_description || '');
      setExpected(currentBug.expected_result || ''); // Заполняем ожидаемый
      setActual(currentBug.actual_result || '');     // Заполняем фактический
      setVersion(currentBug.version_product || '');
      setSelectedOS(currentBug.os ? currentBug.os.split(', ') : []);
    }
  }, [currentBug, isOpen]);

  // Поиск почты по ID (created_by_fk)
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
    const baseUrl = (import.meta as any).env.VITE_API_URL;
    const res = await fetch(`${baseUrl}/bugs/${task.id}`);
    const updatedBugs = await res.json().catch(() => []);
    onBugSaved(updatedBugs);
  };

  const handleSave = async (opts?: { closeOnSuccess?: boolean }) => {
    const userId = Number(localStorage.getItem('userId'));
    const payload: any = {
      created_by: userId,
      severity, priority, status, description,
      playback_description: steps,
      expected_result: expected, // Отправляем ожидаемый
      actual_result: actual,     // Отправляем фактический
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
      const baseUrl = (import.meta as any).env.VITE_API_URL;
      const url = isEditing 
        ? `${baseUrl}/bugs/${getBugId()}`
        : `${baseUrl}/bugs/${task?.id}`;
      
      const res = await fetch(url, {
        method: isEditing ? 'PATCH' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        await refreshBugs();
        if (pendingPhotoFile) {
          // Для нового бага берём максимальный ID из обновлённого списка
          const bugsRes = await fetch(`${baseUrl}/bugs/${task.id}`);
          const bugs = await bugsRes.json().catch(() => []);
          const targetId = isEditing ? getBugId() : Math.max(...bugs.map((b: any) => b.id));
          if (targetId) {
            const form = new FormData();
            form.append('photo', pendingPhotoFile);
            await fetch(`${baseUrl}/bugs/${targetId}/photo`, { method: 'POST', body: form });
            setPendingPhotoFile(null);
          }
        }
        if (opts?.closeOnSuccess !== false) onClose();
      }
    } catch (err) { console.error(err); }
  };

  const handlePass = async () => {
    const bugPk = getBugId();
    if (!bugPk) return;

    setLifecyclePending(true);
    try {
      const baseUrl = (import.meta as any).env.VITE_API_URL;
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
        passed_by: currentUserId || null,
        passed_time: nowIso,
        accepted_by: getAcceptedById() ?? null,
        accepted_time: acceptedTime ?? null,
      };

      const res = await fetch(`${baseUrl}/bugs/${bugPk}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        await refreshBugs();
        // Оставляем модал открытым, чтобы сразу увидеть обновлённый audit trail.
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLifecyclePending(false);
    }
  };

  const handleAccept = async () => {
    const bugPk = getBugId();
    if (!bugPk) return;

    setLifecyclePending(true);
    try {
      const baseUrl = (import.meta as any).env.VITE_API_URL;
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
        passed_by: getPassedById() ?? null,
        passed_time: passedTime ?? null,
        accepted_by: currentUserId || null,
        accepted_time: nowIso,
      };

      const res = await fetch(`${baseUrl}/bugs/${bugPk}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        await refreshBugs();
        // Оставляем модал открытым, чтобы сразу увидеть обновлённый audit trail.
      }
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
  const canPass = isAssignee && (passedById == null);
  const canAccept = isCreator && (passedById != null) && (acceptedById == null);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-slate-900/40 backdrop-blur-md flex items-center justify-center p-4 z-[1000]">
      <div className="bg-white w-full max-w-[900px] max-h-[90vh] rounded-[32px] shadow-2xl overflow-hidden flex flex-col p-10">
        
        <h1 className="text-3xl font-black text-slate-900 mb-8">
          {isEditing ? 'Редактировать баг' : 'Создать баг'}
        </h1>
            <h1 className="text-xl font-black text-slate-900 mb-8">
              Bug ID: {currentBug?.id}
            </h1>
        <div className="flex-1 overflow-y-auto pr-2 space-y-6">
          {/* Селекторы */}
          <div className="grid grid-cols-3 gap-6">
            
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Серьёзность</label>
              <select value={severity} onChange={e => setSeverity(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none">
                <option value="Low">Низкий</option><option value="Medium">Средний</option><option value="High">Высокий</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Приоритет</label>
              <select value={priority} onChange={e => setPriority(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none">
                <option value="Low">Низкий</option><option value="Medium">Средний</option><option value="High">Высокий</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Статус</label>
              <select value={status} onChange={e => setStatus(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none">
                <option value="Open">Открыт</option><option value="In Progress">В работе</option><option value="Closed">Закрыт</option>
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
                    <option key={e} value={e}>
                      {e}
                    </option>
                  ))}
                </select>
              )}
            </div>
          )}

          {/* Описания */}
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Описание</label>
              <textarea value={description} onChange={e => setDescription(e.target.value)} className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[120px] outline-none" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Шаги воспроизведения</label>
              <textarea value={steps} onChange={e => setSteps(e.target.value)} className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[120px] outline-none" />
            </div>
          </div>

          {/* ВЕРНУЛИ: Ожидаемый и фактический результаты */}
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Ожидаемый результат</label>
              <textarea 
                value={expected} 
                onChange={e => setExpected(e.target.value)} 
                className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[80px] outline-none" 
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Фактический результат</label>
              <textarea 
                value={actual} 
                onChange={e => setActual(e.target.value)} 
                className="w-full p-4 rounded-xl border border-slate-300 bg-slate-50 min-h-[80px] outline-none" 
              />
            </div>
          </div>

          {/* Версия и OS */}
          <div className="grid grid-cols-2 gap-6 items-end">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">Версия</label>
              <input value={version} onChange={e => setVersion(e.target.value)} className="w-full p-3 rounded-xl border border-slate-100 bg-slate-50 outline-none" placeholder="1.0.0" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-900">ОС</label>
              <div className="flex gap-2">
                {['Win', 'Mac', 'Linux', 'iOS', 'Android'].map(os => (
                  <button key={os} onClick={() => setSelectedOS(prev => prev.includes(os) ? prev.filter(o => o !== os) : [...prev, os])}
                    className={`px-4 py-2 rounded-lg text-xs font-bold border transition-all ${selectedOS.includes(os) ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-slate-600 border-slate-100'}`}>
                    {os}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Фото */}
          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-900">Фото</label>
            <label className="flex items-center gap-3 cursor-pointer w-full p-3 rounded-xl border border-dashed border-slate-200 bg-slate-50 hover:border-blue-400 transition-colors">
              <span className="text-lg">📎</span>
              <span className="text-sm text-slate-500 truncate">{photoName || 'Прикрепить изображение...'}</span>
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

          {/* Lightbox */}
          {lightbox && photo && (
            <div
              className="fixed inset-0 z-[2000] bg-black/80 flex items-center justify-center p-4"
              onClick={() => setLightbox(false)}
            >
              <img src={photo} alt="full" className="max-w-full max-h-full rounded-xl shadow-2xl" />
            </div>
          )}

          {/* Audit Trail */}
          <div className="pt-6 border-t border-slate-50">
            <h2 className="text-xl font-black text-slate-900 mb-6">История жизненного цикла</h2>
            <div className="grid grid-cols-2 gap-6">
              <div className="p-6 rounded-3xl border border-slate-100 bg-slate-50/50">
                <p className="text-[10px] font-black text-slate-400 uppercase mb-2">СОЗДАН</p>
                <p className="font-bold text-slate-900">
                  {/* Если ID создателя совпадает с твоим — пишем почту из localStorage напрямую */}
                  {createdById === currentUserId ? currentUserEmail : getUserEmail(createdById)}
                </p>
                {isEditing && createdTime && (
                  <p className="text-[10px] text-slate-400 mt-1">
                    {formatDate(createdTime)}
                  </p>
                )}
              </div>
              <div className="p-6 rounded-3xl border border-slate-100 bg-slate-50/50">
                <p className="text-[10px] font-black text-slate-400 uppercase mb-2">ЗАКРЕПЛЕН ЗА</p>
                <p className="font-bold text-slate-900">
                  {assignedToId != null ? getUserEmail(assignedToId) : '—'}
                </p>
                {assignedTime && (
                  <p className="text-[10px] text-slate-400 mt-1">{formatDate(assignedTime)}</p>
                )}
              </div>

              <div className="p-6 rounded-3xl border border-slate-100 bg-slate-50/50">
                <p className="text-[10px] font-black text-slate-400 uppercase mb-2">СДАЛ</p>
                <p className="font-bold text-slate-900">
                  {passedById != null ? getUserEmail(passedById) : '—'}
                </p>
                {passedTime && (
                  <p className="text-[10px] text-slate-400 mt-1">{formatDate(passedTime)}</p>
                )}
              </div>

              <div className="p-6 rounded-3xl border border-slate-100 bg-slate-50/50">
                <p className="text-[10px] font-black text-slate-400 uppercase mb-2">ПРИНЯЛ</p>
                <p className="font-bold text-slate-900">
                  {acceptedById != null ? getUserEmail(acceptedById) : '—'}
                </p>
                {acceptedTime && (
                  <p className="text-[10px] text-slate-400 mt-1">{formatDate(acceptedTime)}</p>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Футер с кнопками */}
        <div className="mt-8 flex justify-end items-center gap-6">
          <button onClick={onClose} className="text-sm font-bold text-slate-500 uppercase">Отмена</button>
          {isCreator && assignedToId == null && (
            <button
              type="button"
              disabled={lifecyclePending || !assignedToEmailChoice}
              onClick={() => handleSave({ closeOnSuccess: false })}
              className="bg-slate-900 text-white px-10 py-4 rounded-3xl font-black shadow-lg shadow-slate-200 hover:scale-[1.02] transition-all disabled:opacity-60"
            >
              Закрепить
            </button>
          )}
          {canPass && (
            <button
              type="button"
              disabled={lifecyclePending}
              onClick={handlePass}
              className="bg-emerald-600 text-white px-10 py-4 rounded-3xl font-black shadow-lg shadow-emerald-100 hover:scale-[1.02] transition-all disabled:opacity-60"
            >
              Сдать
            </button>
          )}
          {canAccept && (
            <button
              type="button"
              disabled={lifecyclePending}
              onClick={handleAccept}
              className="bg-purple-600 text-white px-10 py-4 rounded-3xl font-black shadow-lg shadow-purple-100 hover:scale-[1.02] transition-all disabled:opacity-60"
            >
              Принять
            </button>
          )}
          <button onClick={() => handleSave()} className="bg-blue-600 text-white px-10 py-4 rounded-3xl font-black shadow-lg shadow-blue-100 hover:scale-[1.02] transition-all">
            {isEditing ? 'Обновить баг' : 'Создать баг'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default BugDetailEditor;