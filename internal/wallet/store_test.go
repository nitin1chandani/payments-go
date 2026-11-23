package wallet_test

import (
	"testing"

	"github.com/nitin1chandani/payments/internal/wallet"
)

// newStore connects to database in WALLET_TEST_DATABASE_URL
// Without it, test is skipped, so go test ./... still passes on the machine without docker

func newStore(t *testing.T) *wallet.Store {
	return nil
}
