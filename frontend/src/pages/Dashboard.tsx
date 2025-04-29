import React, {useEffect, useState} from 'react';
import { Link } from 'react-router-dom';
import { Sidebar } from '../components/Sidebar';
import {Users, CreditCard, AlertCircle, User} from 'lucide-react';
import {useAuth} from "@/context/AuthContext.tsx";
import {httpClient} from "@/lib/httpClient.ts";
import {Account, AccountsStats} from "@/utils/models.tsx";

export const Dashboard: React.FC = () => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [stats, setStats] = useState<AccountsStats>();
  const { user } = useAuth();
  useEffect(() => {
    async function loadAccounts() {
      try {
        const accounts = await (httpClient.get<Account[]>(
            'accounts', {
              'headers': {
                'Authorization': `Bearer ${user?.token}`
              }
            }
        ));
        setAccounts(accounts.data.data);
      } catch (error) {
        console.log(error);
      }
    }

    async function loadStats() {
      try {
        const stats = await (httpClient.get<AccountsStats>(
            'accounts/stats', {
              'headers': {
                'Authorization': `Bearer ${user?.token}`
              }
            }
        ));
        setStats(stats.data.data);
      } catch (e) {
        console.log(e);
      }
    }
    loadAccounts();
    loadStats();
  }, [httpClient, user])

  return <div className="flex h-screen bg-gray-50">
      <Sidebar />
      <div className="flex-1 overflow-auto">
        <header className="bg-white shadow">
          <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
            <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
          </div>
        </header>
        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          {/* Stats */}
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0 bg-blue-100 rounded-md p-3">
                    <Users className="h-6 w-6 text-blue-600" />
                  </div>
                  <div className="ml-5">
                    <p className="text-sm font-medium text-gray-500 truncate">
                      Total Users
                    </p>
                    <p className="mt-1 text-3xl font-semibold text-gray-900">
                      {stats?.total_users}
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0 bg-green-100 rounded-md p-3">
                    <CreditCard className="h-6 w-6 text-green-600" />
                  </div>
                  <div className="ml-5">
                    <p className="text-sm font-medium text-gray-500 truncate">
                      Total Accounts
                    </p>
                    <p className="mt-1 text-3xl font-semibold text-gray-900">
                      {stats?.total}
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0 bg-yellow-100 rounded-md p-3">
                    <User className="h-6 w-6 text-yellow-600" />
                  </div>
                  <div className="ml-5">
                    <p className="text-sm font-medium text-gray-500 truncate">
                      Closed Accounts
                    </p>
                    <p className="mt-1 text-3xl font-semibold text-gray-900">
                      {stats?.closed}
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0 bg-red-100 rounded-md p-3">
                    <AlertCircle className="h-6 w-6 text-red-600" />
                  </div>
                  <div className="ml-5">
                    <p className="text-sm font-medium text-gray-500 truncate">
                      Suspended Accounts
                    </p>
                    <p className="mt-1 text-3xl font-semibold text-gray-900">
                      {stats?.suspended}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* Accounts List */}
          <div className="mt-8">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-medium text-gray-900">
                Recent Accounts
              </h2>
              <Link to="/create-account" className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                Create New
              </Link>
            </div>
            <div className="bg-white shadow overflow-hidden sm:rounded-md">
              <ul className="divide-y divide-gray-200">
                {accounts.map(account => {
                return <li key={account.account_id}>
                      <Link to={`/accounts/${account.account_id}`} className="block hover:bg-gray-50">
                        <div className="px-4 py-4 sm:px-6">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center">
                              <p className="text-sm font-medium text-blue-600 truncate">
                                {account.account_number}
                              </p>
                              <p className={`ml-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                                ${account.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : account.status === 'suspended' ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'}`}>
                                {account.status}
                              </p>
                            </div>
                            <div className="text-sm text-gray-500">
                              Balance: £
                              {account.balance.amount.toLocaleString('en-US', {
                            minimumFractionDigits: 2
                          })}
                            </div>
                          </div>
                          <div className="mt-2 sm:flex sm:justify-between">
                            <div className="sm:flex">
                              <p className="flex items-center text-sm text-gray-500">
                                {account.first_name + ' ' + account.last_name || 'Unknown User'}
                              </p>
                            </div>
                            <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                              <p>{account.account_type}</p>
                            </div>
                          </div>
                        </div>
                      </Link>
                    </li>;
              })}
              </ul>
            </div>
          </div>
        </main>
      </div>
    </div>;
};