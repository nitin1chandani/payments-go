BEGIN;

DROP VIEW IF EXISTS account_balance;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS account;
DROP FUNCTION IF EXISTS forbid_ledger_changes();
DROP FUNCTION IF EXISTS check_transaction_balanced();
DROP TYPE IF EXISTS entry_direction;

COMMIT;
