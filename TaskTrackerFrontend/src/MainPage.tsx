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
        if (!confirm('Удалить таску?')) return;

        try {
            const baseUrl = (import.meta as any).env.VITE_API_URL;
            const res = await fetch(`${baseUrl}/tasks/${taskId}`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ owner_id: currentUserId }),
            });

            if (!res.ok) throw new Error(`Delete failed: ${res.status}`);

            // Сбрасываем открытые модалки/редактор если удалили текущую
            if (selectedTask?.id === taskId) {
                setSelectedTask(null);
                setIsEditorOpen(false);
                setSelectedBugId(undefined);
            }

            await fetchTasks();
        } catch (err: any) {
            console.error(err);
            alert('Не удалось удалить таску');
        }
    };

    // Обновление списка багов в конкретной задаче
    const handleBugSavedInState = (updatedBugs: any[]) => {
        setTasks(prev => prev.map(t => 
            t.id === selectedTask?.id ? { ...t, bugs: updatedBugs } : t
        ));
    };

    // Реальная логика создания таски
    const handleCreateTask = async (data: { name: string; desc: string }) => {
        try {
            const userId = localStorage.getItem('userId');
            const baseUrl = (import.meta as any).env.VITE_API_URL;
            if (!userId) return alert("Ошибка: Авторизуйтесь снова");

            const response = await fetch(`${baseUrl}/tasks`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    title: data.name,
                    description: data.desc,
                    owner_id: Number(userId)
                }),
            });

            if (!response.ok) throw new Error('Не удалось создать задачу');
            
            await fetchTasks(); // Перезагружаем список
            setIsModalOpen(false);
        } catch (err: any) { alert(err.message); }
    };

    return (
        <div className="p-8 w-full">
            <div className="flex justify-between items-start mb-8">
                <div>
                    <h1 className="text-3xl font-black text-gray-900">Active Tasks</h1>
                    
                </div>
                <button onClick={() => setIsModalOpen(true)} className="bg-blue-600 text-white px-6 py-2.5 rounded-lg font-bold shadow-md hover:bg-blue-700 transition-all">
                    + Create New Task
                </button>
            </div>

            {isLoading ? (
                <div className="flex justify-center py-20"><div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600"></div></div>
            ) : (
                <div className="space-y-3">
                    {tasks.map(task => (
                        <div
                            key={task.id}
                            onClick={() => setSelectedTask(task)}
                            className="relative grid grid-cols-12 items-center bg-white p-4 cursor-pointer hover:shadow-md rounded-xl border border-gray-100 transition-all"
                        >
                            <div className="col-span-2 text-sm font-bold text-blue-600">ID-{task.id}</div>
                            <div className="col-span-3 font-bold text-gray-900 truncate pr-4">{task.title}</div>
                            <div className="col-span-7 text-sm text-gray-500 truncate">{task.description || "Нет описания"}</div>
                            {task.owner_id === currentUserId && (
                                <button
                                    type="button"
                                    onClick={(e) => { e.stopPropagation(); handleDeleteTask(task.id); }}
                                    className="absolute right-4 top-1/2 -translate-y-1/2 text-sm font-bold text-red-600 bg-red-50 border border-red-100 px-3 py-1 rounded-lg hover:bg-red-100"
                                >
                                    Удалить
                                </button>
                            )}
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
                    onBugsLoaded={handleBugSavedInState} // Важно: обновляем баги в родителе
                />
            )}

            {isEditorOpen && selectedTask && (
    <BugDetailEditor 
        isOpen={isEditorOpen} 
        onClose={() => {
            setIsEditorOpen(false);
            setSelectedBugId(undefined);
        }} 
        // Ищем задачу
        task={tasks.find(t => t.id === selectedTask.id) || selectedTask} 
        // Ищем конкретный баг внутри этой задачи по ID
        currentBug={tasks
            .find(t => t.id === selectedTask.id)?.bugs
            ?.find((b: any) => b.id === Number(selectedBugId))
        }
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

    const handleLogout = () => { localStorage.clear(); navigate('/login'); window.location.reload(); };

    return (
        <header className="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200 w-full">
            <div className="text-blue-600 font-bold text-xl tracking-tight">BugTracker</div>
            <div className="relative">
                <button onClick={() => setIsOpen(!isOpen)} className="flex items-center justify-center w-10 h-10 rounded-xl bg-blue-100 border-2 border-blue-600 text-blue-600">
                    <span className="text-xs font-bold">UI</span>
                </button>
                {isOpen && (
                    <div className="absolute right-0 mt-2 w-48 bg-white border rounded-lg shadow-lg py-1 z-50">
                        <div className="px-4 py-2 text-xs text-gray-400 border-b">{userEmail}</div>
                        <button onClick={handleLogout} className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 font-medium">Выйти</button>
                    </div>
                )}
            </div>
        </header>
    );
}

function Sidebar() {
    return (
        <aside className="w-64 bg-[#F8FAFC] border-r border-gray-200 p-5">
            <button className="flex items-center gap-x-3 w-full p-4 bg-white rounded-lg shadow-sm text-blue-600 font-semibold">
                📋 My Tasks
            </button>
        </aside>
    );
}

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