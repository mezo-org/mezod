-- Migration number: 0000 	 2024-11-14T13:02:39.927Z

CREATE TABLE IF NOT EXISTS activity (
  address VARCHAR(42) PRIMARY KEY,
  tx_count INTEGER NOT NULL DEFAULT 0,
  deployed_contracts INTEGER NOT NULL DEFAULT 0,
  deployed_contracts_tx_count INTEGER NOT NULL DEFAULT 0,
  claimed_btc TEXT NOT NULL DEFAULT '0'
);