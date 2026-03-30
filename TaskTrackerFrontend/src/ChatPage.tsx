import { useEffect, useMemo, useRef, useState } from 'react';
import { API_URL } from './config';

type Thread = {
  id: number;
  scope: 'org' | 'project' | 'dm';
  peer_email?: string;
  title?: string;
  last_message?: string;
  last_message_at?: string;
  unread_count?: number;
};
type Msg = {
  id: number;
  user_id: number;
  user_email: string;
  body: string;
  created_at: string;
  edited_at?: string;
  deleted_at?: string;
};

export default function ChatPage({ onBack }: { onBack: () => void }) {
  const jwtToken = localStorage.getItem('jwtToken') || '';
  const authHeaders = useMemo(() => (jwtToken ? { Authorization: `Bearer ${jwtToken}` } : {}), [jwtToken]);
  const currentUserId = Number(localStorage.getItem('userId') || '0');

  const selectedOrgId = Number(localStorage.getItem('selectedOrgId') || '0');
  const selectedProjectId = Number(localStorage.getItem('selectedProjectId') || '0');

  const [mode, setMode] = useState<'org' | 'project' | 'dm'>('org');
  const [threads, setThreads] = useState<Thread[]>([]);
  const [activeThreadId, setActiveThreadId] = useState<number>(0);
  const [messages, setMessages] = useState<Msg[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const [editingMessageId, setEditingMessageId] = useState<number>(0);
  const [editingBody, setEditingBody] = useState('');
  const [search, setSearch] = useState('');
  const [body, setBody] = useState('');
  const [dmEmail, setDmEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const listRef = useRef<HTMLDivElement | null>(null);
  const lastMsgIdRef = useRef<number>(0);

  const loadThreads = async () => {
    setLoading(true);
    try {
      let url = `${API_URL}/chat/threads?scope=${mode}`;
      if (mode === 'org') url += `&org_id=${selectedOrgId}`;
      if (mode === 'project') url += `&project_id=${selectedProjectId}`;
      const res = await fetch(url, { headers: authHeaders });
      const data = await res.json().catch(() => []);
      const list = Array.isArray(data) ? data : [];
      setThreads(list);
      if (list[0]?.id && !list.some((t: Thread) => t.id === activeThreadId)) {
        setActiveThreadId(list[0].id);
      }
    } finally {
      setLoading(false);
    }
  };

  const ensureScopeThread = async () => {
    if (mode === 'dm') return;
    const payload = mode === 'org' ? { scope: 'org', org_id: selectedOrgId } : { scope: 'project', project_id: selectedProjectId };
    const res = await fetch(`${API_URL}/chat/threads`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify(payload),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok && data?.id) setActiveThreadId(data.id);
  };

  const loadMessages = async (beforeId?: number) => {
    if (!activeThreadId) {
      setMessages([]);
      setHasMore(false);
      return;
    }
    const url = new URL(`${API_URL}/chat/threads/${activeThreadId}/messages`);
    url.searchParams.set('limit', '40');
    if (beforeId) url.searchParams.set('before_id', String(beforeId));
    const res = await fetch(url.toString(), { headers: authHeaders });
    const data = await res.json().catch(() => []);
    const list = (Array.isArray(data) ? data : []) as Msg[];
    const ordered = [...list].reverse();
    setHasMore(list.length >= 40);
    if (beforeId) {
      setMessages(prev => [...ordered, ...prev]);
    } else {
      setMessages(ordered);
    }
  };

  const markRead = async (threadId: number) => {
    if (!threadId) return;
    await fetch(`${API_URL}/chat/threads/${threadId}/read`, {
      method: 'POST',
      headers: authHeaders,
    });
  };

  useEffect(() => { loadThreads(); }, [mode]);
  useEffect(() => { ensureScopeThread(); }, [mode, selectedOrgId, selectedProjectId]);
  useEffect(() => {
    setEditingMessageId(0);
    setEditingBody('');
    setSearch('');
    loadMessages();
    markRead(activeThreadId);
  }, [activeThreadId]);

  useEffect(() => {
    if (!messages.length) return;
    const lastId = messages[messages.length - 1].id;
    if (lastId !== lastMsgIdRef.current) {
      lastMsgIdRef.current = lastId;
      listRef.current?.scrollTo({ top: listRef.current.scrollHeight, behavior: 'smooth' });
    }
  }, [messages]);

  useEffect(() => {
    const id = window.setInterval(() => { loadMessages(); }, 3000);
    return () => window.clearInterval(id);
  }, [activeThreadId]);

  useEffect(() => {
    const id = window.setInterval(() => { loadThreads(); }, 5000);
    return () => window.clearInterval(id);
  }, [mode, selectedOrgId, selectedProjectId]);

  useEffect(() => {
    if (!activeThreadId) return;
    const id = window.setInterval(async () => {
      const res = await fetch(`${API_URL}/chat/threads/${activeThreadId}/typing`, { headers: authHeaders });
      const data = await res.json().catch(() => []);
      setTypingUsers(Array.isArray(data) ? data : []);
    }, 1500);
    return () => window.clearInterval(id);
  }, [activeThreadId]);

  const createDM = async () => {
    if (!dmEmail.trim()) return;
    const res = await fetch(`${API_URL}/chat/threads`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify({ scope: 'dm', email: dmEmail.trim() }),
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      alert(data?.error || 'Не удалось создать личку');
      return;
    }
    setDmEmail('');
    await loadThreads();
    if (data?.id) setActiveThreadId(data.id);
  };

  const sendMessage = async () => {
    if (!activeThreadId || !body.trim()) return;
    const res = await fetch(`${API_URL}/chat/threads/${activeThreadId}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify({ body: body.trim() }),
    });
    if (!res.ok) return;
    setBody('');
    await updateTyping(false);
    await loadMessages();
    await markRead(activeThreadId);
    await loadThreads();
  };

  const visibleMessages = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return messages;
    return messages.filter(m =>
      (m.user_email || '').toLowerCase().includes(q) ||
      (m.body || '').toLowerCase().includes(q)
    );
  }, [messages, search]);

  const renderWithDateSeparators = (items: Msg[]) => {
    const out: any[] = [];
    let lastDate = '';
    for (const m of items) {
      const day = new Date(m.created_at).toLocaleDateString();
      if (day !== lastDate) {
        lastDate = day;
        out.push(
          <div key={`sep-${m.id}`} className="text-center text-[10px] font-bold text-slate-400 py-1">
            {day}
          </div>
        );
      }
      const mine = m.user_id === currentUserId;
      out.push(
        <div key={m.id} className={`max-w-[80%] rounded-2xl px-3 py-2 ${mine ? 'ml-auto bg-indigo-600 text-white' : 'bg-slate-100 text-slate-800'}`}>
          <div className={`text-[10px] mb-1 ${mine ? 'text-indigo-100' : 'text-slate-500'}`}>{m.user_email}</div>
          {editingMessageId === m.id ? (
            <div className="space-y-2">
              <textarea
                value={editingBody}
                onChange={(e) => setEditingBody(e.target.value)}
                className="w-full text-sm rounded-lg border border-indigo-200 text-slate-900 p-2"
              />
              <div className="flex gap-2 justify-end">
                <button onClick={() => { setEditingMessageId(0); setEditingBody(''); }} className="text-xs font-bold text-slate-500">Отмена</button>
                <button onClick={saveEdit} className="text-xs font-bold text-indigo-100 underline">Сохранить</button>
              </div>
            </div>
          ) : (
            <>
              <div className={`text-sm whitespace-pre-wrap ${m.deleted_at ? 'italic opacity-75' : ''}`}>{m.deleted_at ? 'Сообщение удалено' : m.body}</div>
              <div className={`text-[10px] mt-1 flex items-center gap-2 ${mine ? 'text-indigo-100' : 'text-slate-400'}`}>
                {m.edited_at && !m.deleted_at && <span>изменено</span>}
                {mine && !m.deleted_at && (
                  <>
                    <button onClick={() => { setEditingMessageId(m.id); setEditingBody(m.body); }} className="underline">ред.</button>
                    <button onClick={() => deleteMessage(m.id)} className="underline">удал.</button>
                  </>
                )}
              </div>
            </>
          )}
        </div>
      );
    }
    return out;
  };

  const updateTyping = async (isTyping: boolean) => {
    if (!activeThreadId) return;
    await fetch(`${API_URL}/chat/threads/${activeThreadId}/typing`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify({ is_typing: isTyping }),
    });
  };

  const saveEdit = async () => {
    if (!editingMessageId || !editingBody.trim()) return;
    const res = await fetch(`${API_URL}/chat/messages/${editingMessageId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify({ body: editingBody.trim() }),
    });
    if (!res.ok) return;
    setEditingMessageId(0);
    setEditingBody('');
    await loadMessages();
    await loadThreads();
  };

  const deleteMessage = async (id: number) => {
    if (!confirm('Удалить сообщение?')) return;
    const res = await fetch(`${API_URL}/chat/messages/${id}`, {
      method: 'DELETE',
      headers: authHeaders,
    });
    if (!res.ok) return;
    await loadMessages();
    await loadThreads();
  };

  return (
    <div className="min-h-screen bg-slate-50 p-6">
      <div className="max-w-5xl mx-auto space-y-4">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="text-slate-500 hover:text-slate-900 text-2xl leading-none">←</button>
          <div>
            <h1 className="text-3xl font-black text-slate-900">Чаты</h1>
            <p className="text-sm text-slate-500">Организация, проект и личные сообщения</p>
          </div>
        </div>

        <div className="flex gap-2">
          {(['org', 'project', 'dm'] as const).map(m => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={`px-4 py-2 rounded-xl text-sm font-bold ${mode === m ? 'bg-indigo-600 text-white' : 'bg-white border border-slate-200 text-slate-700'}`}
            >
              {m === 'org' ? 'Организация' : m === 'project' ? 'Проект' : 'Личка'}
            </button>
          ))}
        </div>

        {mode === 'dm' && (
          <div className="flex gap-2">
            <input
              value={dmEmail}
              onChange={(e) => setDmEmail(e.target.value)}
              placeholder="email для лички"
              className="flex-1 px-3 py-2.5 rounded-xl border border-slate-200 bg-white text-sm outline-none"
            />
            <button onClick={createDM} className="px-4 py-2.5 rounded-xl bg-slate-900 text-white font-bold text-sm">Открыть</button>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-[280px_1fr] gap-4">
          <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-3 space-y-2 h-[65vh] overflow-y-auto">
            {loading && <p className="text-xs text-slate-400 px-2">Загрузка...</p>}
            {threads.map(t => (
              <button
                key={t.id}
                onClick={() => setActiveThreadId(t.id)}
                className={`w-full text-left px-3 py-2.5 rounded-xl text-sm font-semibold ${activeThreadId === t.id ? 'bg-indigo-50 text-indigo-700' : 'hover:bg-slate-50 text-slate-700'}`}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="truncate">{t.scope === 'dm' ? (t.peer_email || `DM #${t.id}`) : (t.title || `${t.scope === 'org' ? 'Org' : 'Project'} #${t.id}`)}</p>
                    {t.last_message && <p className="text-xs text-slate-400 truncate mt-0.5">{t.last_message}</p>}
                  </div>
                  {!!t.unread_count && t.unread_count > 0 && (
                    <span className="shrink-0 text-[10px] font-black bg-indigo-600 text-white rounded-full px-2 py-0.5">{t.unread_count}</span>
                  )}
                </div>
              </button>
            ))}
            {threads.length === 0 && <p className="text-sm text-slate-400 px-2">Нет чатов</p>}
          </div>

          <div className="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col h-[65vh]">
            {hasMore && messages.length > 0 && (
              <button
                onClick={() => loadMessages(messages[0].id)}
                className="mb-2 self-center text-xs font-bold text-indigo-600 hover:text-indigo-700"
              >
                Загрузить более старые
              </button>
            )}
            <div className="mb-2">
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Поиск в сообщениях..."
                className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm outline-none"
              />
            </div>
            <div ref={listRef} className="flex-1 overflow-y-auto space-y-2 pr-1">
              {renderWithDateSeparators(visibleMessages)}
              {visibleMessages.length === 0 && <p className="text-sm text-slate-400">Сообщений не найдено</p>}
            </div>
            {typingUsers.length > 0 && (
              <div className="text-xs text-slate-400 mb-2">{typingUsers.join(', ')} печатает...</div>
            )}
            <div className="pt-3 border-t border-slate-100 flex gap-2">
              <input
                value={body}
                onChange={(e) => { setBody(e.target.value); updateTyping(e.target.value.trim().length > 0); }}
                placeholder={activeThreadId ? 'Введите сообщение...' : 'Выберите чат'}
                className="flex-1 px-3 py-2.5 rounded-xl border border-slate-200 text-sm outline-none"
                disabled={!activeThreadId}
                onKeyDown={(e) => { if (e.key === 'Enter') sendMessage(); }}
                onBlur={() => updateTyping(false)}
              />
              <button
                onClick={sendMessage}
                disabled={!activeThreadId || !body.trim()}
                className="px-4 py-2.5 rounded-xl bg-indigo-600 text-white font-bold text-sm disabled:opacity-50"
              >
                Отправить
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

