CREATE TABLE IF NOT EXISTS transactions
(
    id       VARCHAR(64) PRIMARY KEY,
    amount   DECIMAL(10, 2)                          NOT NULL,
    currency VARCHAR(10)                             NOT NULL,
    sender   VARCHAR(255)                            NOT NULL,
    receiver VARCHAR(255)                            NOT NULL,
    status   ENUM ('pending', 'completed', 'failed') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
