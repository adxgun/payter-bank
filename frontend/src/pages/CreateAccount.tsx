import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Sidebar } from '../components/Sidebar';
import { User, Save } from 'lucide-react';
import {httpClient, SuccessResponse} from "@/lib/httpClient.ts";
import {useAuth} from "@/context/AuthContext.tsx";
export const CreateAccount: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [formData, setFormData] = useState({
    // User details
    first_name: '',
    last_name: '',
    password: '',
    email: '',
    accountType: 'Current',
    initialDeposit: 0,
    user_role: 'CUSTOMER'
  });
  const [step, setStep] = useState(1);
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const {
      name,
      value
    } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name === 'initialDeposit' ? parseFloat(value) || 0 : value
    }));
  };
  const handleSubmit = async(e: React.FormEvent) => {
    e.preventDefault();
    const createUserPayload = {
      first_name: formData.first_name,
      last_name: formData.last_name,
      password: formData.password,
      email: formData.email,
      user_type: formData.user_role,
    }

    const createUserResult = await httpClient.post('users', createUserPayload, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${user?.token}`,
        }
    });

    const createAccountPayload = {
      account_type: formData.accountType,
      initial_deposit: formData.initialDeposit,
      currency: 'GBP',
      user_id: createUserResult.data.data.user_id,
    }
    const createAccountResult = await httpClient.post('accounts', createAccountPayload, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${user?.token}`,
      }
    })

    console.log(JSON.stringify(createUserResult))
    console.log(JSON.stringify(createAccountResult))
    alert('Account created successfully!');
    navigate('/dashboard');
  };
  return <div className="flex h-screen bg-gray-50">
      <Sidebar />
      <div className="flex-1 overflow-auto">
        <header className="bg-white shadow">
          <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
            <h1 className="text-2xl font-semibold text-gray-900">
              Create User & Account
            </h1>
          </div>
        </header>
        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Create a new user and associated account
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                Fill out all required information below
              </p>
            </div>
            <div className="border-t border-gray-200">
              <div className="px-4 py-5 sm:px-6">
                <div className="flex justify-center mb-8">
                  <div className="flex items-center">
                    <div className={`flex items-center justify-center w-10 h-10 rounded-full ${step === 1 ? 'bg-blue-600' : 'bg-blue-200'}`}>
                      <User className="h-6 w-6 text-white" />
                    </div>
                    <div className={`w-24 h-1 ${step === 1 ? 'bg-blue-600' : 'bg-blue-200'}`}></div>
                    <div className={`flex items-center justify-center w-10 h-10 rounded-full ${step === 2 ? 'bg-blue-600' : 'bg-blue-200'}`}>
                      <Save className="h-6 w-6 text-white" />
                    </div>
                  </div>
                </div>
                <form onSubmit={handleSubmit}>
                  {step === 1 ? <div className="space-y-6">
                      <h4 className="text-lg font-medium text-gray-900">
                        User Information
                      </h4>
                      <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-2">
                        <div>
                          <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                            First Name
                          </label>
                          <input type="text" name="first_name" id="first_name" value={formData.first_name} onChange={handleChange} required className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                        </div>
                        <div>
                          <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                            Last Name
                          </label>
                          <input type="text" name="last_name" id="last_name" value={formData.last_name} onChange={handleChange} required className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                        </div>
                        <div>
                          <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                            Email Address
                          </label>
                          <input type="email" name="email" id="email" value={formData.email} onChange={handleChange} required className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                        </div>
                        <div>
                          <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                            Password
                          </label>
                          <input type="password" name="password" id="password" value={formData.password} onChange={handleChange} required className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                        </div>
                        <div>
                          <label htmlFor="user_role" className="block text-sm font-medium text-gray-700">
                            Account Type
                          </label>
                          <select id="user_role" name="user_role" value={formData.user_role} onChange={handleChange} className="mt-1 block w-full bg-white border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                            <option>ADMIN</option>
                            <option>CUSTOMER</option>
                          </select>
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <button type="button" onClick={() => setStep(2)} className="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                          Next
                        </button>
                      </div>
                    </div> : <div className="space-y-6">
                      <h4 className="text-lg font-medium text-gray-900">
                        Account Information
                      </h4>
                      <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-2">
                        <div>
                          <label htmlFor="accountType" className="block text-sm font-medium text-gray-700">
                            Account Type
                          </label>
                          <select id="accountType" name="accountType" value={formData.accountType} onChange={handleChange} className="mt-1 block w-full bg-white border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
                            <option>Current</option>
                          </select>
                        </div>
                        <div>
                          <label htmlFor="initialDeposit" className="block text-sm font-medium text-gray-700">
                            Initial Deposit Amount (£)
                          </label>
                          <input type="number" name="initialDeposit" id="initialDeposit" min="0" step="0.01" value={formData.initialDeposit} onChange={handleChange} className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
                        </div>
                      </div>
                      <div className="flex justify-between">
                        <button type="button" onClick={() => setStep(1)} className="inline-flex justify-center py-2 px-4 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                          Back
                        </button>
                        <button type="submit" className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                          Create Account
                        </button>
                      </div>
                    </div>}
                </form>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>;
};