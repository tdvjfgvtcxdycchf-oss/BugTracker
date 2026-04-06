import { useNavigate, useLocation } from 'react-router-dom';

const P = '#7C5CBF';
const PL = '#EDE9F7';

function IconPencil() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
    </svg>
  );
}

function IconGrid() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/>
      <rect x="14" y="14" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/>
    </svg>
  );
}

function IconChat() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
    </svg>
  );
}

function IconChart() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/>
      <line x1="6" y1="20" x2="6" y2="14"/><line x1="2" y1="20" x2="22" y2="20"/>
    </svg>
  );
}

function IconGear() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="3"/>
      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
    </svg>
  );
}

function IconMenu() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
    </svg>
  );
}

const NAV = [
  { path: '/new', label: 'Новая задача', Icon: IconPencil },
  { path: '/', label: 'Мои задачи', Icon: IconGrid },
  { path: '/chat', label: 'Чаты', Icon: IconChat },
  { path: '/analytics', label: 'Аналитика', Icon: IconChart },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const userEmail = localStorage.getItem('userEmail') || '';
  const initial = userEmail.slice(0, 1).toUpperCase();

  const isActive = (path: string) =>
    path === '/' ? pathname === '/' : pathname.startsWith(path);

  return (
    <div className="flex h-screen" style={{ background: '#F5F4FA' }}>
      {/* Sidebar */}
      <aside className="w-[80px] bg-white border-r border-gray-100 flex flex-col items-center pt-4 pb-4 shrink-0">
        {NAV.map(({ path, label, Icon }) => {
          const active = isActive(path);
          return (
            <button
              key={path}
              onClick={() => navigate(path)}
              className="flex flex-col items-center gap-1 w-full px-1 py-3 transition-colors"
              style={{ color: active ? P : '#9CA3AF' }}
            >
              <span style={active ? { background: PL, borderRadius: 10, padding: 6 } : { padding: 6 }}>
                <Icon />
              </span>
              <span className="text-[9px] font-medium leading-tight text-center">{label}</span>
            </button>
          );
        })}
        <div className="mt-auto">
          <button
            onClick={() => navigate('/admin')}
            className="flex flex-col items-center gap-1 w-full px-1 py-3 transition-colors"
            style={{ color: pathname === '/admin' ? P : '#9CA3AF' }}
          >
            <span style={pathname === '/admin' ? { background: PL, borderRadius: 10, padding: 6 } : { padding: 6 }}>
              <IconGear />
            </span>
          </button>
        </div>
      </aside>

      {/* Right side */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* Header */}
        <header className="flex items-center justify-between px-5 py-3 bg-white border-b border-gray-100 shrink-0 h-14">
          <div className="flex items-center gap-3">
            <span className="text-gray-400"><IconMenu /></span>
            <div className="flex items-center gap-1.5">
              <span className="font-bold text-gray-900 text-[15px]">🐛 BugTracker</span>
            </div>
          </div>
          <button onClick={() => navigate('/profile')} className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <div
              className="w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-bold shrink-0"
              style={{ background: P }}
            >
              {initial}
            </div>
            <span className="text-sm text-gray-700 font-medium hidden sm:block max-w-[120px] truncate">{userEmail}</span>
          </button>
        </header>

        {/* Content */}
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
}
