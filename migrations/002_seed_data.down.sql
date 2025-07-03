-- Remove seed data
DELETE FROM users WHERE username = 'admin';
DELETE FROM departments WHERE name IN ('Human Resources', 'Finance', 'Information Technology', 'Sales', 'Marketing', 'Operations');
DELETE FROM leave_types WHERE name IN ('Annual Leave', 'Sick Leave', 'Maternity Leave', 'Paternity Leave', 'Compassionate Leave', 'Study Leave');
DELETE FROM currencies WHERE code IN ('USD', 'ZWL', 'ZAR', 'GBP', 'EUR');
