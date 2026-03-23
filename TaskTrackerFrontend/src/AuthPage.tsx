import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';



export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true); // Переключатель Вход/Регистрация
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setPending(true);

    const userData = { email, password };
    const baseUrl = (import.meta as any).env.VITE_API_URL;
  // Выбираем URL в зависимости от режима: вход или регистрация
    const url = isLogin 
    ? `${baseUrl}/login` // Замени на свой путь для логина
    : `${baseUrl}/users`;

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userData),
    });

    const data = await response.json();

    if (!response.ok) {
      // Выводим ошибку из бэкенда (например, "invalid password" или "user already exists")
      throw new Error(data.error || 'Ошибка при выполнении запроса');
    }

    // Проверяем наличие ID (бэкенд теперь возвращает {"id": ...})
    const userId = data.id || data.Id;

    if (userId) {
      // СОХРАНЯЕМ ДАННЫЕ
      localStorage.setItem('isAuthenticated', 'true');
      localStorage.setItem('userId', userId.toString());
      localStorage.setItem('userEmail', email);

      // ПЕРЕХОДИМ НА ГЛАВНУЮ
      navigate('/');
      window.location.reload(); 
    } else {
      throw new Error("Сервер не вернул ID пользователя");
    }

    } catch (err: any) {
      setFormError(err.message);
      console.error('Auth error:', err);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full space-y-8 bg-white p-8 rounded-2xl shadow-xl border border-gray-100">
        <div>
          <h2 className="text-center text-3xl font-extrabold text-gray-900">
            {isLogin ? 'Войти в аккаунт' : 'Создать аккаунт'}
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Или{' '}
            <button 
              onClick={() => setIsLogin(!isLogin)}
              className="font-medium text-blue-600 hover:text-blue-500 underline"
            >
              {isLogin ? 'зарегистрироваться' : 'войти в существующий'}
            </button>
          </p>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="rounded-md shadow-sm space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Email</label>
              <input
                type="email"
                required
                className="appearance-none relative block w-full px-3 py-3 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-lg focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="email@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Пароль</label>
              <input
                type="password"
                required
                className="appearance-none relative block w-full px-3 py-3 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-lg focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
          </div>

          {formError && (
            <div className="text-red-500 text-sm text-center bg-red-50 p-2 rounded-lg">
              {formError}
            </div>
          )}
          <button
            type="submit"
            disabled={pending}
            className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-lg text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
          >
            {pending ? 'Загрузка...' : (isLogin ? 'Войти' : 'Зарегистрироваться')}
          </button>
        </form>
      </div>
    </div>
  );
}
