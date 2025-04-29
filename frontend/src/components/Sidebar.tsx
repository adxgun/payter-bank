import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Home, PlusCircle, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
export const Sidebar: React.FC = () => {
  const location = useLocation();
  const {
    logout
  } = useAuth();
  const isActive = (path: string) => {
    return location.pathname === path;
  };
  const navItems = [{
    name: 'Dashboard',
    path: '/dashboard',
    icon: <Home size={20} />
  }, {
    name: 'Create Account',
    path: '/create-account',
    icon: <PlusCircle size={20} />
  }];
  return <div className="h-screen w-64 bg-white border-r border-gray-200 flex flex-col">
      <div className="p-6">
        <h2 className="text-2xl font-bold text-blue-800">BankAdmin</h2>
      </div>
      <nav className="flex-1 px-4 py-4">
        <ul className="space-y-2">
          {navItems.map(item => <li key={item.path}>
              <Link to={item.path} className={`flex items-center px-4 py-3 text-sm rounded-lg ${isActive(item.path) ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-100'}`}>
                <span className="mr-3">{item.icon}</span>
                {item.name}
              </Link>
            </li>)}
        </ul>
      </nav>
      <div className="p-4 border-t border-gray-200">
        <button onClick={() => logout()} className="flex items-center px-4 py-3 text-sm text-gray-700 rounded-lg hover:bg-gray-100 w-full">
          <LogOut size={20} className="mr-3" />
          Logout
        </button>
      </div>
    </div>;
};