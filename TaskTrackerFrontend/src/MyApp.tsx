import { useState } from 'react';
import { Navigate, BrowserRouter, Routes, Route } from 'react-router-dom';
import AuthPage from './AuthPage';
import MainPage from './MainPage';
import AnalyticsPage from './AnalyticsPage';
import AdminPage from './AdminPage';
import ChatPage from './ChatPage';
import ProfilePage from './ProfilePage';


export default function App() {
  const [isAuthenticated] = useState(
    localStorage.getItem('isAuthenticated') === 'true' && !!localStorage.getItem('jwtToken')
  );
  const userRole = localStorage.getItem('userRole') || '';
  const canAccessAdmin = isAuthenticated && (userRole === 'admin' || userRole === 'pm');

  return (
    <BrowserRouter>
      <main className="">
        <Routes>
          <Route path="/" element={isAuthenticated ? <MainPage/> : <Navigate to="/login" replace/>} />
          <Route path="/login" element={<AuthPage/>} />
          <Route path="/analytics" element={isAuthenticated ? <AnalyticsPage onBack={() => window.history.back()} /> : <Navigate to="/login" replace/>} />
          <Route path="/admin" element={canAccessAdmin ? <AdminPage onBack={() => window.history.back()} /> : <Navigate to="/" replace/>} />
          <Route path="/chat" element={isAuthenticated ? <ChatPage onBack={() => window.history.back()} /> : <Navigate to="/login" replace/>} />
          <Route path="/profile" element={isAuthenticated ? <ProfilePage onBack={() => window.history.back()} /> : <Navigate to="/login" replace/>} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}


