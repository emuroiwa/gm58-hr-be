-- Insert default currencies
INSERT INTO currencies (code, name, symbol, is_active, is_base_currency) VALUES
('USD', 'US Dollar', ', true, true),
('ZWL', 'Local Dollar', 'Z, true, false),
('ZAR', 'South African Rand', 'R', true, false),
('GBP', 'British Pound', '£', true, false),
('EUR', 'Euro', '€', true, false);

-- Insert default leave types
INSERT INTO leave_types (name, description, days_per_year, is_paid, carry_forward, max_carry_days) VALUES
('Annual Leave', 'Annual vacation leave', 22, true, true, 5),
('Sick Leave', 'Medical leave', 10, true, false, 0),
('Maternity Leave', 'Maternity leave for new mothers', 98, true, false, 0),
('Paternity Leave', 'Paternity leave for new fathers', 10, true, false, 0),
('Compassionate Leave', 'Leave for family emergencies', 5, true, false, 0),
('Study Leave', 'Educational leave', 0, false, false, 0);

-- Insert default departments
INSERT INTO departments (name, description) VALUES
('Human Resources', 'Employee management and HR services'),
('Finance', 'Financial planning and accounting'),
('Information Technology', 'IT support and development'),
('Sales', 'Sales and customer relations'),
('Marketing', 'Marketing and advertising'),
('Operations', 'Day-to-day operations');

-- Insert default admin user (password: admin123)
INSERT INTO users (username, email, password, role) VALUES
('admin', 'admin@company.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin');
