CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_account_number VARCHAR(255),
    transaction_type VARCHAR (255),
    transaction_amount  DECIMAL(27,8),
    recipient_account_number VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);