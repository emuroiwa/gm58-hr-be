-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'employee',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create currencies table
CREATE TABLE currencies (
    id SERIAL PRIMARY KEY,
    code VARCHAR(3) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(5),
    is_active BOOLEAN DEFAULT true,
    is_base_currency BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create exchange_rates table
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency_id INTEGER REFERENCES currencies(id),
    to_currency_id INTEGER REFERENCES currencies(id),
    rate DECIMAL(15,6) NOT NULL,
    effective_date TIMESTAMP NOT NULL,
    source VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create departments table
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    manager_id INTEGER,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create positions table
CREATE TABLE positions (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    department_id INTEGER REFERENCES departments(id),
    description TEXT,
    min_salary DECIMAL(15,2),
    max_salary DECIMAL(15,2),
    currency_id INTEGER REFERENCES currencies(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create employees table
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    employee_number VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    national_id VARCHAR(50) UNIQUE,
    tax_number VARCHAR(50),
    passport_number VARCHAR(50),
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20),
    alternative_phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100) DEFAULT 'Local',
    
    -- Employment Details
    position_id INTEGER REFERENCES positions(id),
    department_id INTEGER REFERENCES departments(id),
    manager_id INTEGER,
    
    -- Salary Information
    basic_salary DECIMAL(15,2),
    currency_id INTEGER REFERENCES currencies(id),
    payment_method VARCHAR(20) DEFAULT 'bank_transfer',
    payment_schedule VARCHAR(20) DEFAULT 'monthly',
    
    -- Bank Details
    bank_name VARCHAR(100),
    bank_account VARCHAR(50),
    bank_branch VARCHAR(100),
    bank_code VARCHAR(20),
    swift_code VARCHAR(20),
    
    -- Employment Dates
    hire_date DATE,
    probation_end_date DATE,
    contract_end_date DATE,
    termination_date DATE,
    
    -- Status
    employment_type VARCHAR(20) DEFAULT 'permanent',
    employment_status VARCHAR(20) DEFAULT 'active',
    is_active BOOLEAN DEFAULT true,
    
    -- Medical and Emergency
    emergency_contact_name VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    medical_aid_number VARCHAR(50),
    medical_aid_provider VARCHAR(100),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create payroll_periods table
CREATE TABLE payroll_periods (
    id SERIAL PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    description TEXT,
    processed_at TIMESTAMP,
    processed_by INTEGER REFERENCES users(id),
    approved_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year, month)
);

-- Create payslips table
CREATE TABLE payslips (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER REFERENCES employees(id),
    payroll_period_id INTEGER REFERENCES payroll_periods(id),
    currency_id INTEGER REFERENCES currencies(id),
    exchange_rate DECIMAL(15,6),
    
    -- Earnings
    basic_salary DECIMAL(15,2),
    overtime DECIMAL(15,2) DEFAULT 0,
    allowances DECIMAL(15,2) DEFAULT 0,
    bonus DECIMAL(15,2) DEFAULT 0,
    commission DECIMAL(15,2) DEFAULT 0,
    other_earnings DECIMAL(15,2) DEFAULT 0,
    total_earnings DECIMAL(15,2),
    
    -- Deductions
    payee_tax DECIMAL(15,2),
    aids_levy DECIMAL(15,2),
    nssa_contribution DECIMAL(15,2),
    pension_contribution DECIMAL(15,2) DEFAULT 0,
    medical_aid DECIMAL(15,2) DEFAULT 0,
    union_dues DECIMAL(15,2) DEFAULT 0,
    loan_deductions DECIMAL(15,2) DEFAULT 0,
    other_deductions DECIMAL(15,2) DEFAULT 0,
    total_deductions DECIMAL(15,2),
    
    -- Net Pay
    net_pay DECIMAL(15,2),
    
    -- Base Currency Amounts
    total_earnings_base DECIMAL(15,2),
    total_deductions_base DECIMAL(15,2),
    net_pay_base DECIMAL(15,2),
    
    -- Working Days
    working_days INTEGER,
    days_worked INTEGER,
    days_absent INTEGER,
    
    -- Status
    status VARCHAR(20) DEFAULT 'generated',
    payment_reference VARCHAR(100),
    payment_date TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, payroll_period_id)
);

-- Create allowances table
CREATE TABLE allowances (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER REFERENCES employees(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    amount DECIMAL(15,2),
    currency_id INTEGER REFERENCES currencies(id),
    is_fixed BOOLEAN DEFAULT true,
    percentage DECIMAL(5,2),
    is_taxable BOOLEAN DEFAULT true,
    is_recurring BOOLEAN DEFAULT true,
    start_date DATE,
    end_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create deductions table
CREATE TABLE deductions (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER REFERENCES employees(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    amount DECIMAL(15,2),
    currency_id INTEGER REFERENCES currencies(id),
    is_fixed BOOLEAN DEFAULT true,
    percentage DECIMAL(5,2),
    is_statutory BOOLEAN DEFAULT false,
    is_recurring BOOLEAN DEFAULT true,
    start_date DATE,
    end_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create leave_types table
CREATE TABLE leave_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    days_per_year INTEGER,
    is_paid BOOLEAN DEFAULT true,
    carry_forward BOOLEAN DEFAULT false,
    max_carry_days INTEGER,
    requires_approval BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create leave_requests table
CREATE TABLE leave_requests (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER REFERENCES employees(id),
    leave_type_id INTEGER REFERENCES leave_types(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_requested INTEGER NOT NULL,
    reason TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    approved_by INTEGER REFERENCES employees(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create tax_certificates table
CREATE TABLE tax_certificates (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER REFERENCES employees(id),
    year INTEGER NOT NULL,
    total_earnings DECIMAL(15,2),
    total_tax DECIMAL(15,2),
    currency_id INTEGER REFERENCES currencies(id),
    certificate_number VARCHAR(50),
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, year)
);

-- Create audit_logs table
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_employees_employee_number ON employees(employee_number);
CREATE INDEX idx_employees_is_active ON employees(is_active);
CREATE INDEX idx_employees_department ON employees(department_id);
CREATE INDEX idx_employees_manager ON employees(manager_id);
CREATE INDEX idx_payslips_employee_period ON payslips(employee_id, payroll_period_id);
CREATE INDEX idx_payroll_periods_year_month ON payroll_periods(year, month);
CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
CREATE INDEX idx_exchange_rates_currencies ON exchange_rates(from_currency_id, to_currency_id);
CREATE INDEX idx_exchange_rates_date ON exchange_rates(effective_date);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);

-- Add foreign key constraints
ALTER TABLE departments ADD CONSTRAINT fk_departments_manager 
    FOREIGN KEY (manager_id) REFERENCES employees(id);

ALTER TABLE employees ADD CONSTRAINT fk_employees_manager 
    FOREIGN KEY (manager_id) REFERENCES employees(id);
