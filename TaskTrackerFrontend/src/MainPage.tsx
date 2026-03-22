import React, { useState } from 'react';
import { useNavigate, BrowserRouter, Routes, Route } from 'react-router-dom';
import TaskModal from './TaskModal';
import BugsModal from './BugsModal';
import BugDetailEditor from './BugDetailEditor';

// 1. Выносим Header наружу (хорошая практика)
function Header() {
    const [isOpen, setIsOpen] = useState(false);
    const navigate = useNavigate();

    const handleLogout = () => {
    
    localStorage.removeItem('isAuthenticated');
    navigate('/login');
    window.location.reload();
  };

  return (
    <header className="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200 relative">
      <div className="flex items-center space-x-8">
        <div className="text-blue-600 font-bold text-xl tracking-tight">BugTracker</div>
        <nav className="flex space-x-6 text-sm font-medium text-gray-500">
          <button className="text-blue-600 border-b-2 border-blue-600 pb-4 -mb-4.25">Dashboard</button>
          
        </nav>
      </div>

      <div className="relative">
        <button 
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center justify-center w-10 h-10 rounded-xl bg-blue-100/70 border-2 border-blue-600 text-blue-600 hover:bg-blue-200 transition-all focus:outline-none"
        >
          <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
        </button>

        {isOpen && (
          <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50">
            <button
              onClick={handleLogout}
              className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors font-medium"
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
    
    // 1. Контейнер боковой панели. 
    // `h-screen` - во весь экран, `sticky` - прижат при скролле, `border-r` - граница справа.
    <aside className="w-64 h-screen bg-[#F8FAFC] border-r border-gray-200 sticky top-0 left-0 p-5 flex flex-col gap-y-4">
      
      {/* 2. Кнопка "My Tasks" (как на скрине) */}
      <button className="flex items-center gap-x-3 w-full p-4 bg-white rounded-lg shadow-sm text-blue-600 hover:bg-blue-50 transition-colors">
        
        {/* SVG Иконка "Задача/Заметки" */}
        <div className="w-10 h-10 rounded-xl flex items-center justify-center">
            <svg 
              className="w-6 h-6 text-blue-600" 
              viewBox="0 0 24 24" 
              fill="none" 
              stroke="currentColor" 
              strokeWidth="2.5" 
              strokeLinecap="round" 
              strokeLinejoin="round"
            >
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <path d="M10 13l2 2 4-4"></path>
            </svg>
        </div>

        {/* Текст кнопки */}
        <span className="font-semibold text-lg text-blue-600">My Tasks</span>
      </button>
    </aside>
  );
};

const tasks = [
  { id: '1', name: 'API Gateway Refactoring', desc: 'Updating internal microservices to use gRPC protocol instead of RES...' },
  { id: '2', name: 'Auth Component Overhaul', desc: 'Implementing multi-factor authentication and biometrics for mobile cli...' },
  { id: '3', name: 'Database Migration 4.0', desc: 'Migrating legacy Postgres instance to distributed CockroachDB clust...' },
  { id: '4', name: 'Load Balancer Optimization', desc: 'Fine-tuning weighted round-robin algorithms for edge clusters...' },
  { id: '5', name: 'CI/CD Pipeline Security', desc: 'Hardening GitHub Action runners with isolated environments...' },
];

function Dashboard() {
    // 1. Состояния для модальных окон
    const [isModalOpen, setIsModalOpen] = useState(false);       // Окно создания новой таски
    const [isEditorOpen, setIsEditorOpen] = useState(false);     // Большое окно редактора бага (скриншот 2)
    const [selectedTask, setSelectedTask] = useState<any>(null); // Какая таска сейчас выбрана
    const [selectedBugId, setSelectedBugId] = useState<string | undefined>(); // Какой баг открыт (если есть)

    // 2. Данные (список таск)
    const [tasks, setTasks] = useState([
        { 
            id: 'T-001', 
            name: 'Initial Setup', 
            desc: 'Project initialization', 
            email: 'admin@tracker.com', 
            bugs: [{ id: 'B-1', title: 'Token mismatch' }] 
        }
    ]);

    // 3. Логика создания новой Таски
    const handleCreateTask = (data: { name: string; desc: string }) => {
  // Получаем email текущего пользователя или ставим заглушку
        const userEmail = localStorage.getItem('userEmail') || 'guest@tracker.com';

        const newTask = {
            id: `T-00${tasks.length + 1}`,
            name: data.name,
            desc: data.desc,
            email: userEmail, // ТЕПЕРЬ ПОЛЕ ЕСТЬ, ОШИБКА ИСЧЕЗНЕТ
            bugs: []
        };
        
        setTasks([newTask, ...tasks]);
        setIsModalOpen(false); // Закрываем модалку после создания
    };
    // 4. Логика создания Бага (внутри выбранной таски)
    const handleCreateBug = (newBugData: any) => {
        setTasks(prev => prev.map(task => {
            if (task.id === selectedTask?.id) {
                return {
                    ...task,
                    bugs: [...(task.bugs || []), { 
                        id: `B-NEW-${Date.now()}`, 
                        ...newBugData 
                    }]
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
                bugs: task.bugs.map((bug: any) => 
                bug.id === selectedBugId ? { ...bug, ...updatedBugData } : bug
                )
            };
            }
            return task;
        }));
    };

    // Находим актуальные данные выбранной задачи из общего списка
    const currentTaskData = tasks.find(t => t.id === selectedTask?.id);

    return (
        <div className="p-8 w-full">
            {/* ШАПКА И КНОПКА СОЗДАНИЯ ТАСКИ */}
            <div className="flex justify-between items-start mb-8">
                <div>
                    <h1 className="text-3xl font-black text-gray-900">Active Tasks</h1>
                    <p className="text-gray-500 text-sm">Managing prioritized issues for the current sprint</p>
                </div>
                <button 
                    onClick={() => setIsModalOpen(true)} 
                    className="bg-blue-600 text-white px-6 py-2.5 rounded-lg font-bold shadow-md hover:bg-blue-700 transition-all"
                >
                    + Create New Task
                </button>
            </div>

            {/* Заголовки таблицы */}
            <div className="grid grid-cols-12 px-4 mb-4 text-[10px] font-bold text-gray-400 uppercase tracking-widest">
            <div className="col-span-2">Task ID</div>
            <div className="col-span-3">Name</div>
            <div className="col-span-5">Description</div> {/* Дали 5 колонок под текст */}
            </div>

            <div className="space-y-3">
            {tasks.map((task) => (
                <div 
                key={task.id} 
                onClick={() => setSelectedTask(task)} 
                className="grid grid-cols-12 items-center bg-white p-4 cursor-pointer hover:border-blue-400 rounded-xl transition-all shadow-sm border border-transparent"
                >
                {/* ID */}
                <div className="col-span-2 text-sm font-medium text-gray-400">{task.id}</div>
                
                {/* Название */}
                <div className="col-span-3 font-bold text-gray-900">{task.name}</div>
                
                {/* ОПИСАНИЕ (Теперь оно здесь отображается) */}
                <div className="col-span-5 text-sm text-gray-500 truncate pr-4">
                    {task.desc || <span className="text-gray-300 italic">No description provided</span>}
                </div>
                
                
                </div>
            ))}
            </div>

            {/* --- ВСЕ МОДАЛЬНЫЕ ОКНА В ОДНОМ МЕСТЕ --- */}

            {/* 1. Окно создания Таски */}
            <TaskModal 
                isOpen={isModalOpen} 
                onClose={() => setIsModalOpen(false)} 
                onCreate={handleCreateTask} 
            />

        {/* 1. Список багов */}
        {selectedTask && !isEditorOpen && (
            <BugsModal 
                task={tasks.find(t => t.id === selectedTask.id) || selectedTask} 
                onClose={() => setSelectedTask(null)} 
                onAddBug={handleCreateBug} // Теперь TypeScript не ругается
                setIsEditorOpen={setIsEditorOpen}
                setSelectedBugId={setSelectedBugId}
            />
        )}

        {/* 2. Редактор */}
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
            onUpdateBug={handleUpdateBug} // ПЕРЕДАЕМ НОВУЮ ФУНКЦИЮ
        />
        )}

        </div>
    );
};

export default function MainPage() {
    return (
        <div className=" bg-white">
              {/* 1. Хедер (верхняя полоса) */}
              <Header /> 
              <div className="flex flex-1">
                {/* 2. Сайдбар (левая колонка) */}
                <Sidebar />
                {/* 3. Основная рабочая область (правая часть) */}
                <main className="flex-1 bg-white overflow-y-auto">
                  {/* Вместо обычного h1 вставляем наш полноценный компонент */}
                  <Dashboard />
                </main>
              </div>
        </div>
    )
}