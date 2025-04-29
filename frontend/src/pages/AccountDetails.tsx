import React, {useEffect, useState} from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Sidebar } from '../components/Sidebar';
import {Ban, XCircle, FileText, Clock, ArrowUpRight, ArrowDownLeft, Check} from 'lucide-react';
import {useAuth} from "@/context/AuthContext.tsx";
import {Account, AuditLog, Transaction} from "@/utils/models.tsx";
import {httpClient} from "@/lib/httpClient.ts";

export const AccountDetails: React.FC = () => {
  const {
    id
  } = useParams<{
    id: string;
  }>();
  const navigate = useNavigate();
  const [account, setAccount] = useState<Account>();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('transactions');
  const [transactionType, setTransactionType] = useState('');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');

  useEffect(() => {
    async function fetchAccount(){
      const result = (await httpClient.get(`accounts/${id}`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${user?.token}`,
        }
      })).data.data as Account;
      console.log(JSON.stringify(result))
      setAccount(result);
    }

    async function fetchTransactions(){
      const result = (await httpClient.get(`accounts/${id}/transactions`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${user?.token}`,
        }
      })).data.data;
      console.log(JSON.stringify(result))
      setTransactions(result);
    }

    async function fetchAuditLogs(){
      const result = (await httpClient.get(`accounts/${id}/logs`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${user?.token}`,
        }
      })).data.data;
      console.log(JSON.stringify(result))
      setAuditLogs(result);
    }

    fetchAccount();
    fetchTransactions();
    fetchAuditLogs();
  }, [id, httpClient, user]);

  const handleTransaction = (e: React.FormEvent) => {
    e.preventDefault();
    const endpoint = transactionType === 'credit' ? 'credit' : 'debit';
    const fromAccountID = transactionType == 'debit' ? account?.account_id : null;
    const toAccountID = transactionType == 'credit' ? account?.account_id : null;
    const payload = {
      from_account_id: fromAccountID,
      to_account_id: toAccountID,
      amount: parseFloat(amount),
      description: description,
      type: transactionType,
    }

    httpClient.post(endpoint, payload, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${user?.token}`,
      }
    })
    alert(`${transactionType} transaction of £${amount} processed successfully!`);
    setAmount('');
    setDescription('');
    setTransactionType('');
  };
  const handleAccountAction = (action: string) => {
    let endpoint = '';
    switch (action) {
      case 'activated':
        endpoint = `accounts/${account?.account_id}/activate`;
        break;
      case 'suspended':
          endpoint = `accounts/${account?.account_id}/suspend`;
          break;
      case 'closed':
        endpoint = `accounts/${account?.account_id}/close`;
        break;
      default:
        throw new Error('Invalid action');
    }
    httpClient.patch(endpoint, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${user?.token}`,
      }
    })

    alert(`Account ${action} successfully!`);
    navigate('/dashboard');
  };
  return <div className="flex h-screen bg-gray-50">
      <Sidebar />
      <div className="flex-1 overflow-auto">
        <header className="bg-white shadow">
          <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center">
              <h1 className="text-2xl font-semibold text-gray-900">
                Account Details
              </h1>
              <div className="flex space-x-3">
                <button onClick={() => handleAccountAction('activated')} className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-green-600 hover:bg-red-700">
                  <Check className="mr-2 h-4 w-4" />
                  Activate
                </button>
                <button onClick={() => handleAccountAction('suspended')} className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-yellow-500 hover:bg-yellow-600">
                  <Ban className="mr-2 h-4 w-4" />
                  Suspend
                </button>
                <button onClick={() => handleAccountAction('closed')} className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700">
                  <XCircle className="mr-2 h-4 w-4" />
                  Close
                </button>
              </div>
            </div>
          </div>
        </header>
        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          {/* Account Overview */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Account Information
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                Account details and customer information.
              </p>
            </div>
            <div className="border-t border-gray-200">
              <dl>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">
                    Account Number
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {account?.account_number}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">
                    Account Type
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {account?.account_type}
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">
                    Current Balance
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    £
                    {account?.balance.amount.toLocaleString('en-US', {
                    minimumFractionDigits: 2
                  })}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Status</dt>
                  <dd className="mt-1 text-sm sm:mt-0 sm:col-span-2">
                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                      ${account?.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : account?.status === 'SUSPENDED' ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'}`}>
                      {account?.status}
                    </span>
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">
                    Account Holder
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {account?.first_name + ' ' + account?.last_name}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">
                    Email Address
                  </dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {account?.email}
                  </dd>
                </div>
              </dl>
            </div>
          </div>
          {/* Transaction Form */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                New Transaction
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                Process a new transaction for this account.
              </p>
            </div>
            <div className="border-t border-gray-200 px-4 py-5 sm:px-6">
              <form onSubmit={handleTransaction}>
                <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div className="sm:col-span-2">
                    <label htmlFor="transaction-type" className="block text-sm font-medium text-gray-700">
                      Transaction Type
                    </label>
                    <select id="transaction-type" value={transactionType} onChange={e => setTransactionType(e.target.value)} required className="mt-1 block w-full bg-white border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                      <option value="">Select type</option>
                      <option value="credit">Credit (Deposit)</option>
                      <option value="debit">Debit (Withdrawal)</option>
                    </select>
                  </div>
                  <div className="sm:col-span-2">
                    <label htmlFor="amount" className="block text-sm font-medium text-gray-700">
                      Amount (£)
                    </label>
                    <input type="number" name="amount" id="amount" min="0.01" step="0.01" value={amount} onChange={e => setAmount(e.target.value)} required className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                  </div>
                  <div className="sm:col-span-2">
                    <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                      Description
                    </label>
                    <input type="text" name="description" id="description" value={description} onChange={e => setDescription(e.target.value)} required className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                  </div>
                </div>
                <div className="mt-5 flex justify-end">
                  <button type="submit" disabled={!transactionType} className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-blue-300 disabled:cursor-not-allowed">
                    Process Transaction
                  </button>
                </div>
              </form>
            </div>
          </div>
          {/* Tabs */}
          <div className="border-b border-gray-200">
            <nav className="-mb-px flex">
              <button onClick={() => setActiveTab('transactions')} className={`w-1/4 py-4 px-1 text-center border-b-2 font-medium text-sm
                  ${activeTab === 'transactions' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}>
                <FileText className="w-5 h-5 mx-auto mb-1" />
                Transactions
              </button>
              <button onClick={() => setActiveTab('audit')} className={`w-1/4 py-4 px-1 text-center border-b-2 font-medium text-sm
                  ${activeTab === 'audit' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}>
                <Clock className="w-5 h-5 mx-auto mb-1" />
                Audit Log
              </button>
            </nav>
          </div>
          {/* Tab Content */}
          <div className="mt-6">
            {activeTab === 'transactions' ? <div className="bg-white shadow overflow-hidden sm:rounded-md">
                <ul className="divide-y divide-gray-200">
                  {transactions.length > 0 ? transactions.map(transaction => <li key={transaction.id} className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        {transaction.to_account_id === account?.account_id && (
                            <ArrowDownLeft className="h-5 w-5 text-green-500 mr-2" />
                        )}
                        {transaction.from_account_id === account?.account_id && (
                            <ArrowUpRight className="h-5 w-5 text-red-500 mr-2" />
                        )}
                        <div>
                          <p className="text-sm font-medium text-gray-900">
                            {transaction.description}
                          </p>
                          <p className="text-sm text-gray-500">
                            {new Date(transaction.created_at).toLocaleString()}
                          </p>
                        </div>
                      </div>
                      <div
                          className={`text-sm font-medium ${transaction.to_account_id === account?.account_id ? 'text-green-600' : transaction.from_account_id === account?.account_id ? 'text-red-600' : 'text-blue-600'}`}
                      >
                        {transaction.to_account_id === account?.account_id
                            ? '+'
                            : transaction.from_account_id === account?.account_id
                                ? '-'
                                : ''}
                        $
                        {transaction.amount.amount.toLocaleString('en-US', {
                          minimumFractionDigits: 2,
                        })}
                      </div>
                    </div>
                    <p className="mt-1 text-xs text-gray-500">
                      Ref: {transaction.reference_number}
                    </p>
                      </li>) : <li className="px-4 py-6 text-center text-sm text-gray-500">
                      No transactions found for this account.
                    </li>}
                </ul>
              </div> : <div className="bg-white shadow overflow-hidden sm:rounded-md">
                <ul className="divide-y divide-gray-200">
                  {auditLogs.length > 0 ? auditLogs.map(log => <li key={log.created_at} className="px-4 py-4 sm:px-6">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="text-sm font-medium text-gray-900">
                              {
                                log.action_code == 'account_status_change' && log.new_status == 'ACTIVE' ? 'Activated account' : log.action
                              }
                            </p>
                            {
                              log.action_code === 'account_credit' || log.action_code === 'account_debit' ? <h4 className="text-sm text-gray-500">
                                    {log.amount.amount} {log.amount.currency}
                              </h4> : <p className="text-sm text-gray-500"></p>
                            }
                          </div>
                          <div className="text-sm text-gray-500">
                            {new Date(log.created_at).toLocaleString()}
                          </div>
                        </div>
                        <p className="mt-1 text-xs text-gray-500">
                          By: {log.action_by}
                        </p>
                      </li>) : <li className="px-4 py-6 text-center text-sm text-gray-500">
                      No audit logs found for this account.
                    </li>}
                </ul>
              </div>}
          </div>
        </main>
      </div>
    </div>;
};