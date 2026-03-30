import React, { useState } from 'react';
import { Navigate, BrowserRouter, Routes, Route } from 'react-router-dom';
import AuthPage from './AuthPage';
import MainPage from './MainPage';
import AnalyticsPage from './AnalyticsPage';


export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(
    localStorage.getItem('isAuthenticated') === 'true' && !!localStorage.getItem('jwtToken')
  );

  return (
    <BrowserRouter>
      <main className="">
        <Routes>
          <Route path="/" element={isAuthenticated ? <MainPage/> : <Navigate to="/login" replace/>} />
          <Route path="/login" element={<AuthPage/>} />
          <Route path="/analytics" element={isAuthenticated ? <AnalyticsPage onBack={() => window.history.back()} /> : <Navigate to="/login" replace/>} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}


