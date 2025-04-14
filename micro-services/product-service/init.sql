CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    availability BOOLEAN NOT NULL DEFAULT TRUE
);

INSERT INTO categories (name) VALUES ('Electronics'), ('Clothing'), ('Books'), ('Furniture'), ('Fashion');

INSERT INTO products (name, category, price, availability) VALUES 
('iPhone 15', 'Electronics', 999.99, TRUE),
('Nike Shoes', 'Fashion', 120.00, FALSE),
('Dell Laptop', 'Electronics', 850.50, TRUE),
('Wooden Dining Table', 'Furniture', 550.00, FALSE)
('Laptop', 'Electronics', 999.99, TRUE),
('Smartphone', 'Electronics', 699.99, TRUE),
('T-Shirt', 'Clothing', 19.99, TRUE),
('Fiction Book', 'Books', 12.99, FALSE);
