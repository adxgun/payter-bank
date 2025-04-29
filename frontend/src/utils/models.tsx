export interface Account {
  account_id: string;
  account_number: string;
  account_type: string;
  balance: {
    amount: number;
  };
  status: string;
  currency: string;
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface AccountsStats {
  closed: number;
  suspended: number;
  total: number;
  total_users: number;
}

export interface Transaction {
  id: string;
  from_account_id: string;
  to_account_id: string;
  amount: {
    amount: number;
    currency: string;
  };
  reference_number: string;
  description: string;
  status: string;
  currency: string;
  created_at: string;
  updated_at: string;
}

export interface AuditLog {
  account_id: string;
  action: string;
  action_code: string;
  action_by: string;
  amount: {
    amount: number;
    currency: string;
  };;
  current_status: string;
  new_status: string;
  old_status: string;
  created_at: string;
}