import React, { useState, useEffect } from 'react';
import TaskModal from './TaskModal';
import BugsModal from './BugsModal';
import BugDetailEditor from './BugDetailEditor';
import { useNavigate } from 'react-router-dom';

function Dashboard() {
    // 1. Состояния для модальных окон и выбора
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const [selectedTask, setSelectedTask] = useState<any>(null);
    const [selectedBugId, setSelectedBugId] = useState<string | undefined>();

    // 2. Состояния для данных из БД
    const [tasks, setTasks] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    // 3. ЗАГРУЗКА ТАСОК С БЭКЕНДА
    const fetchTasks = async () => {
        setIsLoading(true);
        try {
            const baseUrl = (import.meta as any).env.VITE_API_URL;
            const response = await fetch(`${baseUrl}/tasks`);
            
            if (!response.ok) throw new Error('Ошибка при получении списка задач');
            
            const data = await response.json();
            // Go может вернуть null, если слайс пустой, поэтому подстраховываемся через || []
            setTasks(data || []);
        } catch (err) {
            console.error("Fetch error:", err);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchTasks();
    }, []);

    // 4. Логика создания новой Таски в БД
    const handleCreateTask = async (data: { name: string; desc: string }) => {
        try {
            const userId = localStorage.getItem('userId');
            const baseUrl = (import.meta as any).env.VITE_API_URL;

            const response = await fetch(`${baseUrl}/users`, { // Путь /users согласно твоему HandleCreateUser
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    title: data.name,        // Поле в Go: Title
                    description: data.desc,  // Поле в Go: Description
                    owner_id: Number(userId) // Поле в Go: OwnerId
                }),
            });

            if (!response.ok) throw new Error('Не удалось создать задачу');

            // Обновляем список после успешного создания
            await fetchTasks();
            setIsModalOpen(false);
        } catch (err) {
            alert("Ошибка при создании задачи. Проверьте соединение с бэкендом.");
            console.error(err);
        }
    };

    // 5. Логика для багов (оставляем пока локальную или адаптируй под API позже)
    const handleCreateBug = (newBugData: any) => {
        setTasks(prev => prev.map(task => {
            if (task.id === selectedTask?.id) {
                return {
                    ...task,
                    bugs: [...(task.bugs || []), { id: `B-${Date.now()}`, ...newBugData }]
                };
            }
            return task;
        }));
    };

    const handleUpdateBug = (updatedBugData: any) => {
        setTasks(prev => prev.map(task => {
            if (task.id === selectedTask?.id) {
                return {
                    ...task,
                    bugs: (task.bugs || []).map((bug: any) => 
                        bug.id === selectedBugId ? { ...bug, ...updatedBugData } : bug
                    )
                };
            }
            return task;
        }));
    };

    return (
        <div className="p-8 w-full">
            {/* ЗАГОЛОВОК */}
            <div className="flex justify-between items-start mb-8">
                <div>
                    <h1 className="text-3xl font-black text-gray-900">Active Tasks</h1>
                    <p className="text-gray-500 text-sm">Managing prioritized issues from database</p>
                </div>
                <button 
                    onClick={() => setIsModalOpen(true)} 
                    className="bg-blue-600 text-white px-6 py-2.5 rounded-lg font-bold shadow-md hover:bg-blue-700 transition-all"
                >
                    + Create New Task
                </button>
            </div>

            {/* ОСНОВНОЙ КОНТЕНТ */}
            {isLoading ? (
                // Индикатор загрузки
                <div className="flex flex-col items-center justify-center py-20">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
                    <p className="text-gray-400">Загрузка задач...</p>
                </div>
            ) : tasks.length === 0 ? (
                // СОСТОЯНИЕ: ТАСОК НЕТ
                <div className="flex flex-col items-center justify-center py-20 bg-gray-50 rounded-3xl border-2 border-dashed border-gray-200">
                    <div className="w-20 h-20 bg-white rounded-full shadow-sm flex items-center justify-center mb-4">
                        <svg className="w-10 h-10 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
                        </svg>
                    </div>
                    <h3 className="text-xl font-bold text-gray-900">Задачи отсутствуют</h3>
                    <p className="text-gray-500 mt-1">В базе данных пока нет записей для отображения.</p>
                </div>
            ) : (
                // ВЫВОД ТАСОК
                <>
                    <div className="grid grid-cols-12 px-4 mb-4 text-[10px] font-bold text-gray-400 uppercase tracking-widest">
                        <div className="col-span-2">Task ID</div>
                        <div className="col-span-3">Name</div>
                        <div className="col-span-7">Description</div>
                    </div>

                    <div className="space-y-3">
                        {tasks.map((task) => (
                            <div 
                                key={task.id} 
                                onClick={() => setSelectedTask(task)} 
                                className="grid grid-cols-12 items-center bg-white p-4 cursor-pointer hover:shadow-md hover:border-blue-400 rounded-xl transition-all shadow-sm border border-gray-100"
                            >
                                <div className="col-span-2 text-sm font-bold text-blue-600">
                                    T-{task.id}
                                </div>
                                <div className="col-span-3 font-bold text-gray-900 truncate pr-4">
                                    {task.title}
                                </div>
                                <div className="col-span-7 text-sm text-gray-500 truncate">
                                    {task.description || <span className="text-gray-300 italic">No description provided</span>}
                                </div>
                            </div>
                        ))}
                    </div>
                </>
            )}

            {/* МОДАЛКИ */}
            <TaskModal 
                isOpen={isModalOpen} 
                onClose={() => setIsModalOpen(false)} 
                onCreate={handleCreateTask} 
            />

            {selectedTask && !isEditorOpen && (
                <BugsModal 
                    task={tasks.find(t => t.id === selectedTask.id) || selectedTask} 
                    onClose={() => setSelectedTask(null)} 
                    onAddBug={handleCreateBug} 
                    setIsEditorOpen={setIsEditorOpen}
                    setSelectedBugId={setSelectedBugId}
                />
            )}

            {isEditorOpen && selectedTask && (
                <BugDetailEditor 
                    isOpen={isEditorOpen} 
                    onClose={() => {
                        setIsEditorOpen(false);
                        setSelectedBugId(undefined);
                    }} 
                    task={tasks.find(t => t.id === selectedTask.id) || selectedTask} 
                    bugId={selectedBugId}
                    onCreateBug={handleCreateBug}
                    onUpdateBug={handleUpdateBug}
                />
            )}
        </div>
    );
}

function Header() {
  const [isOpen, setIsOpen] = useState(false);
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.clear(); // Очищаем всё сразу
    navigate('/login');
    window.location.reload();
  };

  return (
    <header className="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200 w-full">
      <div className="flex items-center space-x-8">
        <div className="text-blue-600 font-bold text-xl tracking-tight">BugTracker</div>
        <nav className="flex space-x-6 text-sm font-medium text-gray-500">
          <button className="text-blue-600 border-b-2 border-blue-600">Dashboard</button>
        </nav>
      </div>

      {/* КНОПКА ПРОФИЛЯ */}
      <div className="relative">
        <button 
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center justify-center w-10 h-10 rounded-xl bg-blue-100/70 border-2 border-blue-600 text-blue-600"
        >
          <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
        </button>

        {isOpen && (
          <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50">
            <div className="px-4 py-2 text-xs text-gray-400 border-b mb-1">
              {localStorage.getItem('userEmail')}
            </div>
            <button
              onClick={handleLogout}
              className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 font-medium"
            >
              Выйти
            </button>
          </div>
        )}
      </div>
    </header>
  );
}

function Sidebar() {
  return (
    <aside className="w-64 bg-[#F8FAFC] border-r border-gray-200 p-5 flex flex-col gap-y-4">
      <button className="flex items-center gap-x-3 w-full p-4 bg-white rounded-lg shadow-sm text-blue-600 hover:bg-blue-50 transition-colors">
        <div className="w-10 h-10 rounded-xl flex items-center justify-center">
            <svg className="w-6 h-6 text-blue-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
              <polyline points="14 2 14 8 20 8" />
              <path d="M10 13l2 2 4-4" />
            </svg>
        </div>
        <span className="font-semibold text-lg">My Tasks</span>
      </button>
    </aside>
  );
}

// Убедись, что MainPage правильно собирает структуру
export default function MainPage() {
    return (
        <div className="flex flex-col h-screen bg-white">
            <Header /> 
            <div className="flex flex-1 overflow-hidden">
                <Sidebar />
                <main className="flex-1 bg-white overflow-y-auto">
                    <Dashboard />
                </main>
            </div>
        </div>
    );
}

