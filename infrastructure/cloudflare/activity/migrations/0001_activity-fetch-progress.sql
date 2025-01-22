-- Migration number: 0001 	 2025-01-21T11:54:20.927Z

CREATE TABLE IF NOT EXISTS fetch_progress (
    id TEXT PRIMARY KEY,
    page INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

INSERT INTO fetch_progress (id)
VALUES ("address"), ("tx_count"), ("address_txs"), ("contract_txs");