-- Drop tables in reverse order to avoid foreign key conflicts
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS tax_certificates;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS leave_types;
DROP TABLE IF EXISTS deductions;
DROP TABLE IF EXISTS allowances;
DROP TABLE IF EXISTS payslips;
DROP TABLE IF EXISTS payroll_periods;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS currencies;
DROP TABLE IF EXISTS users;
