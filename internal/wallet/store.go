// Package wallet is the wallet service: accounts, balance and the ledger
package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nitin1chandani/payments/internal/wallet/walletdb"
)

// Database generate modeles serve both as database models and domain models
// The alias means callers only ever import package wallet, never wallet db
type (
	Account                      = walletdb.Account
	AccountBalance               = walletdb.AccountBalance
	Entry                        = walletdb.Entry
	Transaction                  = walletdb.Transaction
	NewAccount                   = walletdb.CreateAccountParams
	Direction                    = walletdb.EntryDirection
	ListEntriesByAccountIDParams = walletdb.ListEntriesByAccountParams
)

const (
	Debit  = walletdb.EntryDirectionDebit
	Credit = walletdb.EntryDirectionCredit
)

// ErrAccountNotFound is returned when account does not exist
var ErrAccountNotFound = errors.New("wallet: account not found")

// Store is the wallet's data access layer, often called as repository.
// It wraps the sqlc queries, and owns database connections
// This way of doing helps in future if we decide to move away from sqlc
type Store struct {
	pool *pgxpool.Pool
	q    *walletdb.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: walletdb.New(pool)}
}

// CreateAccount opens a new account. OwnerID is nil for system accounts.
func (s *Store) CreateAccount(ctx context.Context, account NewAccount) (Account, error) {
	acc, err := s.q.CreateAccount(ctx, account)
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}
	return acc, nil
}

// GetAccountByID returns account with the given id.
func (s *Store) GetAccountByID(ctx context.Context, id uuid.UUID) (Account, error) {
	acc, err := s.q.GetAccount(ctx, id)

	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}

	if err != nil {
		return Account{}, fmt.Errorf("get account %s: %w", id, err)
	}
	return acc, nil
}

// GetBalanceByID returns an account's balance
func (s *Store) GetBalanceByID(ctx context.Context, id uuid.UUID) (AccountBalance, error) {
	balance, err := s.q.GetAccountBalance(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountBalance{}, ErrAccountNotFound
	}

	if err != nil {
		return AccountBalance{}, err
	}

	return balance, err
}

// ListEntriesByAccountID returns a page of an account's entries, newest first.
func (s *Store) ListEntriesByAccountID(ctx context.Context, id uuid.UUID, before int64, limit int32) ([]Entry, error) {
	if before <= 0 {
		before = math.MaxInt64
	}

	entries, err := s.q.ListEntriesByAccount(ctx, ListEntriesByAccountIDParams{
		AccountID: id,
		BeforeID:  before,
		PageSize:  limit,
	})

	if err != nil {
		return nil, fmt.Errorf("list entries %s: %w", id, err)
	}
	return entries, nil
}
