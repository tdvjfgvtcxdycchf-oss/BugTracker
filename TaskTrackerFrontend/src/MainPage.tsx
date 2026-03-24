import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import TaskModal from './TaskModal';
import BugsModal from './BugsModal';
import BugDetailEditor from './BugDetailEditor';

function Dashboard() {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const [selectedTask, setSelectedTask] = useState<any>(null);
    const [selectedBugId, setSelectedBugId] = useState<string | undefined>();
    const [tasks, setTasks] = useState<any[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const currentUserId = Number(localStorage.getItem('userId') || '0');

    const fetchTasks = async () => {
        setIsLoading(true);
        try {
            const baseUrl = (import.meta as any).env.VITE_API_URL;
            const response = await fetch(`${baseUrl}/tasks`);
            const data = await response.json();
            setTasks(data || []);
        } catch (err) { console.error("Fetch tasks error:", err); }
        finally { setIsLoading(false); }
    };

    useEffect(() => { fetchTasks(); }, []);

    const handleDeleteTask = async (taskId: number) => {
        if (!taskId || !currentUserId) return;
        if (!confirm('Удалить задачу?')) return;
        try {
            const baseUrl = (import.meta as any).env.VITE_API_URL;
            const res = await fetch(`${baseUrl}/tasks/${taskId}`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ owner_id: currentUserId }),
            });
            if (!res.ok) throw new Error(`Delete failed: ${res.status}`);
            if (selectedTask?.id === taskId) {
                setSelectedTask(null);
                setIsEditorOpen(false);
                setSelectedBugId(undefined);
            }
            await fetchTasks();
        } catch (err: any) {
            alert('Не удалось удалить задачу');
        }
    };

    const handleBugSavedInState = (updatedBugs: any[]) => {
        setTasks(prev => prev.map(t =>
            t.id === selectedTask?.id ? { ...t, bugs: updatedBugs } : t
        ));
    };

    const handleCreateTask = async (data: { name: string; desc: string }) => {
        try {
            const userId = localStorage.getItem('userId');
            const baseUrl = (import.meta as any).env.VITE_API_URL;
            if (!userId) return alert("Ошибка: Авторизуйтесь снова");
            const response = await fetch(`${baseUrl}/tasks`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title: data.name, description: data.desc, owner_id: Number(userId) }),
            });
            if (!response.ok) throw new Error('Не удалось создать задачу');
            await fetchTasks();
            setIsModalOpen(false);
        } catch (err: any) { alert(err.message); }
    };

    const colors = ['#6366f1', '#f59e0b', '#10b981', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6'];
    const getColor = (id: number) => colors[id % colors.length];

    return (
        <div className="p-4 sm:p-8 w-full max-w-4xl mx-auto">
            <div className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Задачи</h1>
                    <p className="text-sm text-gray-400 mt-0.5">{tasks.length} активных задач</p>
                </div>
                <button
                    onClick={() => setIsModalOpen(true)}
                    className="flex items-center gap-2 bg-indigo-600 text-white px-5 py-2.5 rounded-xl font-semibold shadow-lg shadow-indigo-200 hover:bg-indigo-700 transition-all text-sm"
                >
                    <span className="text-lg leading-none">+</span> Создать задачу
                </button>
            </div>

            {isLoading ? (
                <div className="flex justify-center py-20">
                    <div className="animate-spin rounded-full h-8 w-8 border-2 border-indigo-600 border-t-transparent"></div>
                </div>
            ) : tasks.length === 0 ? (
                <div className="text-center py-24">
                    <div className="text-6xl mb-4">📋</div>
                    <p className="text-lg font-medium text-gray-400">Задач пока нет</p>
                    <p className="text-sm text-gray-300 mt-1">Нажмите «Создать задачу», чтобы начать</p>
                </div>
            ) : (
                <div className="grid gap-3">
                    {tasks.map(task => (
                        <div
                            key={task.id}
                            onClick={() => setSelectedTask(task)}
                            className="group flex items-center bg-white rounded-2xl border border-gray-100 hover:border-indigo-200 hover:shadow-lg hover:shadow-indigo-50 cursor-pointer transition-all duration-200 overflow-hidden"
                        >
                            <div className="w-1 self-stretch shrink-0" style={{ backgroundColor: getColor(task.id) }} />
                            <div className="flex items-center gap-4 flex-1 px-5 py-4 min-w-0">
                                <span className="text-xs font-bold px-2.5 py-1 rounded-lg text-white shrink-0" style={{ backgroundColor: getColor(task.id) }}>
                                    #{task.id}
                                </span>
                                <div className="flex-1 min-w-0">
                                    <p className="font-semibold text-gray-900 truncate">{task.title}</p>
                                    <p className="text-sm text-gray-400 truncate mt-0.5">{task.description || 'Без описания'}</p>
                                </div>
                                {task.owner_id === currentUserId && (
                                    <button
                                        type="button"
                                        onClick={(e) => { e.stopPropagation(); handleDeleteTask(task.id); }}
                                        className="shrink-0 opacity-0 group-hover:opacity-100 text-xs font-semibold text-red-500 bg-red-50 border border-red-100 px-3 py-1.5 rounded-lg hover:bg-red-100 transition-all"
                                    >
                                        Удалить
                                    </button>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            <TaskModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} onCreate={handleCreateTask} />

            {selectedTask && !isEditorOpen && (
                <BugsModal
                    task={tasks.find(t => t.id === selectedTask.id) || selectedTask}
                    onClose={() => setSelectedTask(null)}
                    setIsEditorOpen={setIsEditorOpen}
                    setSelectedBugId={setSelectedBugId}
                    onBugsLoaded={handleBugSavedInState}
                />
            )}

            {isEditorOpen && selectedTask && (
                <BugDetailEditor
                    isOpen={isEditorOpen}
                    onClose={() => { setIsEditorOpen(false); setSelectedBugId(undefined); }}
                    task={tasks.find(t => t.id === selectedTask.id) || selectedTask}
                    currentBug={tasks.find(t => t.id === selectedTask.id)?.bugs?.find((b: any) => b.id === Number(selectedBugId))}
                    bugId={selectedBugId}
                    onBugSaved={handleBugSavedInState}
                />
            )}
        </div>
    );
}

function Header() {
    const [isOpen, setIsOpen] = useState(false);
    const navigate = useNavigate();
    const userEmail = localStorage.getItem('userEmail') || 'Guest';
    const initials = userEmail.slice(0, 2).toUpperCase();

    const handleLogout = () => { localStorage.clear(); navigate('/login'); window.location.reload(); };

    return (
        <header className="flex items-center justify-between px-6 py-3.5 bg-white border-b border-gray-100 w-full">
            <div className="flex items-center gap-2.5">
                <div className="w-7 h-7 rounded-lg bg-indigo-600 flex items-center justify-center">
                    <span className="text-white text-xs font-black">B</span>
                </div>
                <span className="font-bold text-gray-900 tracking-tight">BugTracker</span>
            </div>
            <div className="relative">
                <button
                    onClick={() => setIsOpen(!isOpen)}
                    className="w-9 h-9 rounded-xl bg-indigo-50 text-indigo-600 font-bold text-xs flex items-center justify-center border border-indigo-100 hover:bg-indigo-100 transition-colors"
                >
                    {initials}
                </button>
                {isOpen && (
                    <div className="absolute right-0 mt-2 w-52 bg-white border border-gray-100 rounded-xl shadow-xl py-1 z-50">
                        <div className="px-4 py-2.5 text-xs text-gray-400 border-b border-gray-50 truncate">{userEmail}</div>
                        <button onClick={handleLogout} className="w-full text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 font-medium transition-colors">
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
        <aside className="hidden sm:block w-56 bg-gray-50 border-r border-gray-100 p-4">
            <button className="flex items-center gap-3 w-full px-3 py-2.5 bg-white rounded-xl shadow-sm text-indigo-600 font-semibold text-sm border border-indigo-50">
                <span>📋</span> Мои задачи
            </button>
        </aside>
    );
}

export default function MainPage() {
    return (
        <div className="flex flex-col h-screen bg-gray-50">
            <Header />
            <div className="flex flex-1 overflow-hidden">
                <Sidebar />
                <main className="flex-1 bg-gray-50 overflow-y-auto">
                    <Dashboard />
                </main>
            </div>
        </div>
    );
}
