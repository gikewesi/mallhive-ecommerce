-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    slug TEXT UNIQUE NOT NULL
);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    price DECIMAL(10,2) NOT NULL,
    available BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,
    imageURL TEXT
);

-- Create inventory table
CREATE TABLE IF NOT EXISTS inventories (
    id SERIAL PRIMARY KEY,
    product_id INTEGER UNIQUE NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 0
);

-- Insert categories (with slugs)
INSERT INTO categories (name, slug) VALUES
('Electronics', 'electronics'),
('Clothing', 'clothing'),
('Books', 'books'),
('Furniture', 'furniture'),
('Fashion', 'fashion');

-- Insert products using category IDs
INSERT INTO products (name, slug, category_id, price, available, description, imageURL)
VALUES
('iPhone 15', 'iphone-15', (SELECT id FROM categories WHERE name = 'Electronics'), 999.99, TRUE, 'Latest Apple iPhone.', 'https://example.com/iphone15.jpg'),
('Nike Shoes', 'nike-shoes', (SELECT id FROM categories WHERE name = 'Fashion'), 120.00, FALSE, 'Trendy running shoes.', 'https://example.com/nikeshoes.jpg'),
('Dell Laptop', 'dell-laptop', (SELECT id FROM categories WHERE name = 'Electronics'), 850.50, TRUE, 'High performance laptop.', 'https://example.com/dell.jpg'),
('Wooden Dining Table', 'wooden-dining-table', (SELECT id FROM categories WHERE name = 'Furniture'), 550.00, FALSE, 'Elegant wood table.', 'https://example.com/diningtable.jpg'),
('Laptop', 'laptop', (SELECT id FROM categories WHERE name = 'Electronics'), 999.99, TRUE, 'Basic business laptop.', 'https://example.com/laptop.jpg'),
('Smartphone', 'smartphone', (SELECT id FROM categories WHERE name = 'Electronics'), 699.99, TRUE, 'Android smartphone.', 'https://example.com/smartphone.jpg'),
('T-Shirt', 't-shirt', (SELECT id FROM categories WHERE name = 'Clothing'), 19.99, TRUE, 'Comfortable cotton shirt.', 'https://example.com/tshirt.jpg'),
('Fiction Book', 'fiction-book', (SELECT id FROM categories WHERE name = 'Books'), 12.99, FALSE, 'Bestselling fiction novel.', 'https://example.com/book.jpg');

-- Insert inventory (with example stock quantities)
INSERT INTO inventories (product_id, quantity)
SELECT id, FLOOR(RANDOM() * 100)::INT FROM products;
