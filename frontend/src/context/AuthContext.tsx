import React, {useState, createContext, useContext, ReactNode, useEffect} from 'react';
import {httpClient} from '@/lib/httpClient.ts';

interface Profile {
  user_id: string;
  account_id: string;
  account_type: string;
  first_name: string;
  last_name: string;
  email: string;
  token: string;
}

interface AccessToken {
  token: string;
}

interface AuthContextType {
  user: Profile | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const STORAGE_KEY = 'auth_profile';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{
  children: ReactNode;
}> = ({
  children
}) => {
  useEffect(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const parsed = JSON.parse(stored) as Profile;
        setUser(parsed);
      } catch (e) {
        console.error('Failed to parse stored profile:', e);
        localStorage.removeItem(STORAGE_KEY);
      }
    }
  }, []);

  const [user, setUser] = useState<Profile | null>(null);
  const isAuthenticated = user !== null;
  const login = async (email: string, password: string) => {
    try {
      const result = (await httpClient.post('users/authenticate', {
        email: email,
        password: password
      })).data.data;
      const profile = (await httpClient.get<Profile>('me', {
        headers: {
          'Authorization': `Bearer ${result.token}`
        }
      })).data.data as Profile;

      profile.token = result.token;
      setUser(profile);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(profile));
      return Promise.resolve();
    } catch (err) {
      return Promise.reject(err);
    }
  };
  const logout = () => {
    setUser(null);
    localStorage.removeItem(STORAGE_KEY)
  };
  return <AuthContext.Provider value={{
    user,
    isAuthenticated,
    login,
    logout
  }}>
      {children}
    </AuthContext.Provider>;
};
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};