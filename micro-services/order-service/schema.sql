CREATE DATABASE orders_db;

\c orders_db;

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    product_ids INT[] NOT NULL,
    total NUMERIC(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
