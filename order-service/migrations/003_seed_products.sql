-- +migrate Up
INSERT INTO products (name, description, price) VALUES
('Мыло', 'Обычное мыло', 100.00),
('Верёвка', 'Прочная верёвка', 200.50),
('Стул', 'Деревянный стул', 1500.75),
('VIP-статус', 'Эксклюзивный доступ ко всем возможностям', 9999999.00);

-- +migrate Down
TRUNCATE TABLE products RESTART IDENTITY CASCADE; 