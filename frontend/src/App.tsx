import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { CreateAccount } from './pages/CreateAccount';
import { AccountDetails } from './pages/AccountDetails';
import { AuthProvider } from './context/AuthContext';
export function App() {
  return <AuthProvider>
      <BrowserRouter>
        <div className="min-h-screen bg-gray-50">
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/create-account" element={<CreateAccount />} />
            <Route path="/accounts/:id" element={<AccountDetails />} />
            <Route path="/" element={<Navigate to="/login" replace />} />
          </Routes>
        </div>
      </BrowserRouter>
    </AuthProvider>;
}