-- Drop company-related constraints and indexes
ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_company_name_unique;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_company_employee_number_unique;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_company_email_unique;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_company_national_id_unique;
ALTER TABLE payroll_periods DROP CONSTRAINT IF EXISTS payroll_periods_company_year_month_unique;
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_company_employee_period_unique;
ALTER TABLE tax_certificates DROP CONSTRAINT IF EXISTS tax_certificates_company_employee_year_unique;
ALTER TABLE leave_types DROP CONSTRAINT IF EXISTS leave_types_company_name_unique;

-- Drop indexes
DROP INDEX IF EXISTS idx_departments_company_id;
DROP INDEX IF EXISTS idx_positions_company_id;
DROP INDEX IF EXISTS idx_employees_company_id;
DROP INDEX IF EXISTS idx_payroll_periods_company_id;
DROP INDEX IF EXISTS idx_payslips_company_id;
DROP INDEX IF EXISTS idx_allowances_company_id;
DROP INDEX IF EXISTS idx_deductions_company_id;
DROP INDEX IF EXISTS idx_leave_types_company_id;
DROP INDEX IF EXISTS idx_leave_requests_company_id;
DROP INDEX IF EXISTS idx_tax_certificates_company_id;
DROP INDEX IF EXISTS idx_audit_logs_company_id;

-- Remove company_id from tables
ALTER TABLE audit_logs DROP COLUMN IF EXISTS company_id;
ALTER TABLE tax_certificates DROP COLUMN IF EXISTS company_id;
ALTER TABLE leave_requests DROP COLUMN IF EXISTS company_id;
ALTER TABLE leave_types DROP COLUMN IF EXISTS company_id;
ALTER TABLE deductions DROP COLUMN IF EXISTS company_id;
ALTER TABLE allowances DROP COLUMN IF EXISTS company_id;
ALTER TABLE payslips DROP COLUMN IF EXISTS company_id;
ALTER TABLE payroll_periods DROP COLUMN IF EXISTS company_id;
ALTER TABLE employees DROP COLUMN IF EXISTS company_id;
ALTER TABLE positions DROP COLUMN IF EXISTS company_id;
ALTER TABLE departments DROP COLUMN IF EXISTS company_id;

-- Restore original unique constraints
ALTER TABLE departments DROP CONSTRAINT IF EXISTS uni_departments_name;
ALTER TABLE departments ADD CONSTRAINT departments_name_key UNIQUE(name);
ALTER TABLE employees ADD CONSTRAINT employees_employee_number_key UNIQUE(employee_number);
ALTER TABLE employees ADD CONSTRAINT employees_email_key UNIQUE(email);
ALTER TABLE employees ADD CONSTRAINT employees_national_id_key UNIQUE(national_id);
ALTER TABLE payroll_periods ADD CONSTRAINT payroll_periods_year_month_key UNIQUE(year, month);
ALTER TABLE payslips ADD CONSTRAINT payslips_employee_id_payroll_period_id_key UNIQUE(employee_id, payroll_period_id);
ALTER TABLE tax_certificates ADD CONSTRAINT tax_certificates_employee_id_year_key UNIQUE(employee_id, year);
ALTER TABLE leave_types ADD CONSTRAINT leave_types_name_key UNIQUE(name);

-- Revert users role constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'hr', 'employee'));

-- Drop company-related tables
DROP TABLE IF EXISTS company_settings;
DROP TABLE IF EXISTS company_users;
DROP TABLE IF EXISTS companies;