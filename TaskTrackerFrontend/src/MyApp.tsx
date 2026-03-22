import React, { useState } from 'react';
import { Navigate, BrowserRouter, Routes, Route } from 'react-router-dom';
import AuthPage from './AuthPage';
import MainPage from './MainPage';


export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(
    localStorage.getItem('isAuthenticated') === 'false'
  );

  return (
    <BrowserRouter>
      <main className="">
        <Routes>
          <Route path="/" element={isAuthenticated ? <MainPage/> : <Navigate to="/login" replace/>} />
          <Route path="/login" element={<AuthPage/>} />
        </Routes>
      </main> 
    </BrowserRouter>
  );
}


