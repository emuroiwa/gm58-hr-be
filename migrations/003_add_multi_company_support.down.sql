-- Create companies table
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    website VARCHAR(255),
    tax_number VARCHAR(50),
    registration_no VARCHAR(50),
    industry VARCHAR(100),
    size VARCHAR(20),
    base_currency_id INTEGER REFERENCES currencies(id),
    
    -- Billing
    billing_plan VARCHAR(20) DEFAULT 'free',
    billing_cycle VARCHAR(20) DEFAULT 'monthly',
    subscription_end TIMESTAMP,
    max_employees INTEGER DEFAULT 10,
    
    -- Settings
    logo_url VARCHAR(255),
    payroll_cycle VARCHAR(20) DEFAULT 'monthly',
    work_week_days INTEGER DEFAULT 5,
    work_day_hours DECIMAL(4,2) DEFAULT 8.0,
    overtime_rate DECIMAL(4,2) DEFAULT 1.5,
    weekend_rate DECIMAL(4,2) DEFAULT 2.0,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    
    -- Audit
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create company_users table
CREATE TABLE company_users (
    id SERIAL PRIMARY KEY,
    company_id INTEGER REFERENCES companies(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'employee',
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(company_id, user_id)
);

-- Create company_settings table
CREATE TABLE company_settings (
    id SERIAL PRIMARY KEY,
    company_id INTEGER UNIQUE REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Tax Settings
    enable_paye BOOLEAN DEFAULT true,
    enable_nssa BOOLEAN DEFAULT true,
    enable_aids_levy BOOLEAN DEFAULT true,
    custom_tax_rates JSONB,
    
    -- Leave Settings
    leave_year_start DATE,
    allow_negative_leave BOOLEAN DEFAULT false,
    
    -- Payroll Settings
    payroll_approval_levels INTEGER DEFAULT 1,
    require_timesheet BOOLEAN DEFAULT false,
    
    -- Notification Settings
    email_notifications BOOLEAN DEFAULT true,
    sms_notifications BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add company_id to existing tables
ALTER TABLE departments ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE positions ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE employees ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE payroll_periods ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE payslips ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE allowances ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE deductions ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE leave_types ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE leave_requests ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE tax_certificates ADD COLUMN company_id INTEGER REFERENCES companies(id);
ALTER TABLE audit_logs ADD COLUMN company_id INTEGER REFERENCES companies(id);

-- Update users table to add super_admin role option
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('super_admin', 'company_admin', 'hr', 'manager', 'employee'));

-- Create indexes for better performance
CREATE INDEX idx_companies_code ON companies(code);
CREATE INDEX idx_companies_is_active ON companies(is_active);
CREATE INDEX idx_company_users_company_id ON company_users(company_id);
CREATE INDEX idx_company_users_user_id ON company_users(user_id);
CREATE INDEX idx_company_users_is_active ON company_users(is_active);

-- Add company_id indexes to all tables
CREATE INDEX idx_departments_company_id ON departments(company_id);
CREATE INDEX idx_positions_company_id ON positions(company_id);
CREATE INDEX idx_employees_company_id ON employees(company_id);
CREATE INDEX idx_payroll_periods_company_id ON payroll_periods(company_id);
CREATE INDEX idx_payslips_company_id ON payslips(company_id);
CREATE INDEX idx_allowances_company_id ON allowances(company_id);
CREATE INDEX idx_deductions_company_id ON deductions(company_id);
CREATE INDEX idx_leave_types_company_id ON leave_types(company_id);
CREATE INDEX idx_leave_requests_company_id ON leave_requests(company_id);
CREATE INDEX idx_tax_certificates_company_id ON tax_certificates(company_id);
CREATE INDEX idx_audit_logs_company_id ON audit_logs(company_id);

-- Update unique constraints to include company_id
-- ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_name_key;
ALTER TABLE departments ADD CONSTRAINT departments_company_name_unique UNIQUE(company_id, name);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_employee_number_key;
ALTER TABLE employees ADD CONSTRAINT employees_company_employee_number_unique UNIQUE(company_id, employee_number);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_email_key;
ALTER TABLE employees ADD CONSTRAINT employees_company_email_unique UNIQUE(company_id, email);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_national_id_key;
ALTER TABLE employees ADD CONSTRAINT employees_company_national_id_unique UNIQUE(company_id, national_id);

ALTER TABLE payroll_periods DROP CONSTRAINT IF EXISTS payroll_periods_year_month_key;
ALTER TABLE payroll_periods ADD CONSTRAINT payroll_periods_company_year_month_unique UNIQUE(company_id, year, month);

ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_employee_id_payroll_period_id_key;
ALTER TABLE payslips ADD CONSTRAINT payslips_company_employee_period_unique UNIQUE(company_id, employee_id, payroll_period_id);

ALTER TABLE tax_certificates DROP CONSTRAINT IF EXISTS tax_certificates_employee_id_year_key;
ALTER TABLE tax_certificates ADD CONSTRAINT tax_certificates_company_employee_year_unique UNIQUE(company_id, employee_id, year);

ALTER TABLE leave_types DROP CONSTRAINT IF EXISTS leave_types_name_key;
ALTER TABLE leave_types ADD CONSTRAINT leave_types_company_name_unique UNIQUE(company_id, name);

-- Create default company for existing data (if any)
DO $$
DECLARE
    default_company_id INTEGER;
    default_user_id INTEGER;
BEGIN
    -- Check if there's existing data
    IF EXISTS (SELECT 1 FROM employees LIMIT 1) THEN
        -- Create a default company
        INSERT INTO companies (name, code, email, billing_plan, max_employees, base_currency_id)
        VALUES ('Default Company', 'DEFAULT', 'admin@defaultcompany.com', 'free', 1000, 
                (SELECT id FROM currencies WHERE is_base_currency = true LIMIT 1))
        RETURNING id INTO default_company_id;
        
        -- Update all existing records with the default company
        UPDATE departments SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE positions SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE employees SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE payroll_periods SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE payslips SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE allowances SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE deductions SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE leave_types SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE leave_requests SET company_id = default_company_id WHERE company_id IS NULL;
        UPDATE tax_certificates SET company_id = default_company_id WHERE company_id IS NULL;
        
        -- Create company_users records for existing users
        INSERT INTO company_users (company_id, user_id, role, is_default, is_active)
        SELECT default_company_id, id, 
               CASE 
                   WHEN role = 'admin' THEN 'company_admin'
                   ELSE role
               END,
               true, true
        FROM users;
    END IF;
END $$;

-- Make company_id NOT NULL after migration
ALTER TABLE departments ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE positions ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE employees ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE payroll_periods ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE payslips ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE allowances ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE deductions ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE leave_types ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE leave_requests ALTER COLUMN company_id SET NOT NULL;
ALTER TABLE tax_certificates ALTER COLUMN company_id SET NOT NULL;