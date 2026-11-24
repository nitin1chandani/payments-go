package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nitin1chandani/payments/internal/wallet/walletdb"
)

// posting is one line of transaction: money in or out of one account.
type Posting struct {
	AccountID uuid.UUID
	Direction Direction
	Amount    int64
	Currency  string
}

// NewTransaction is a transaction we want to record.
type NewTransaction struct {
	Description string
	Postings    []Posting
}

// Transaction is a posted journal entry, together with its entries
type TransactionWithEntries struct {
	Transaction
	Entries []Entry `json:"entries"`
}

// inTx runs fn inside a database transaction, handing it queries bound to that
// transaction. It commits when fn returns nil, and rolls back otherwise.
func (s *Store) inTx(ctx context.Context, fn func(q *walletdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	defer tx.Rollback(ctx)

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// PostTransaction records a transaction and creating entries and return entries
func (s *Store) PostTransaction(ctx context.Context, nt NewTransaction) (TransactionWithEntries, error) {
	var posted TransactionWithEntries
	err := s.inTx(ctx, func(q *walletdb.Queries) error {
		journal, err := q.CreateTransaction(ctx, nt.Description)
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}

		posted.Transaction = journal

		for _, p := range nt.Postings {
			entry, err := q.CreateEntry(ctx, walletdb.CreateEntryParams{
				TransactionID: journal.ID,
				AccountID:     p.AccountID,
				Direction:     p.Direction,
				Amount:        p.Amount,
				Currency:      p.Currency,
			})

			if err != nil {
				return fmt.Errorf("insert entry: %w", err)
			}

			posted.Entries = append(posted.Entries, entry)
		}

		return nil
	})

	if err != nil {
		return TransactionWithEntries{}, err
	}

	return posted, nil
}
