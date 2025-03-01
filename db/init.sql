DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateways') THEN
        CREATE TABLE gateways (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            data_format_supported VARCHAR(50) NOT NULL,  
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'countries') THEN
        CREATE TABLE countries (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL UNIQUE,
            code CHAR(2) NOT NULL UNIQUE,
            currency CHAR(3) NOT NULL,
            created_at TIMESTAMP     DEFAULT CURRENT_TIMESTAMP, 
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'gateway_countries') THEN
        CREATE TABLE gateway_countries (
            gateway_id INT NOT NULL, 
            country_id INT NOT NULL,
            PRIMARY KEY (gateway_id, country_id)
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'transactions') THEN
        CREATE TABLE transactions (
            id SERIAL PRIMARY KEY,
            order_id VARCHAR(255) NOT NULL,
            amount DECIMAL(10, 2) NOT NULL,
            type VARCHAR(50) NOT NULL,
            status VARCHAR(50) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  
            gateway_id INT NOT NULL,  
            country_id INT NOT NULL,  
            user_id INT NOT NULL,
            currency VARCHAR(10) NOT NULL
        );
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'ledger') THEN
        CREATE TABLE ledger (
            id SERIAL PRIMARY KEY,
            user_id VARCHAR(255) NOT NULL,
            currency VARCHAR(10) NOT NULL,
            amount DECIMAL(15, 2) NOT NULL,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            CONSTRAINT unique_user_currency UNIQUE (user_id, currency)
        );
        
        CREATE INDEX idx_ledger_user_id ON ledger(user_id);
        CREATE INDEX idx_ledger_currency ON ledger(currency);
    END IF;
END $$;


DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'ledger') THEN
        CREATE TABLE ledger (
            id SERIAL PRIMARY KEY,              -- Added as a best practice for unique row identification
            user_id VARCHAR(255) NOT NULL,     -- Changed to user_id for consistency
            currency VARCHAR(10) NOT NULL,     -- Limited to 10 chars for typical currency codes
            amount DECIMAL(15, 2) NOT NULL,    -- Changed to DECIMAL instead of VARCHAR for proper numeric operations
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        -- Optional: Add indexes for better query performance
        CREATE INDEX idx_ledger_user_id ON ledger(user_id);
        CREATE INDEX idx_ledger_currency ON ledger(currency);
    END IF;
END $$;


-- Insert initial data into countries table
INSERT INTO countries (name, code, currency)
VALUES
 ('INDIA', 'IN', 'INR'),
 ('UAE', 'AE', 'AED')
ON CONFLICT (name) DO NOTHING;


-- Insert initial data into gateways table
INSERT INTO gateways (name, data_format_supported)
VALUES 
  ('STRIPE', 'JSON'),
  ('RAZORPAY', 'JSON'),
  ('DEFAULT_GATEWAY', 'JSON')
ON CONFLICT (name) DO NOTHING;

INSERT INTO gateway_countries (gateway_id, country_id)
VALUES 
  (1, 3),
  (2, 1),
  (2, 3),
  (7, 1),
  (7, 3)
ON CONFLICT DO NOTHING;

INSERT INTO users (username, email, password, country_id)
VALUES 
    ('atharva', 'atharvaajgaonkar29@gmail.com', '$2a$10$hashedpasswordhere', 1),
    ('john_doe', 'johndoe@example.com', '$2a$10$anotherhashedpasswordhere', 2)
ON CONFLICT (name) DO NOTHING;

