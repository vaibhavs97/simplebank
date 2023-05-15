package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	testStore := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">>>before: ", acc1.Balance, acc2.Balance)

	n := 5
	amount := int64(10)

	errChan := make(chan error)
	resultChan := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})

			errChan <- err
			resultChan <- result
		}()
	}

	exist := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errChan
		require.NoError(t, err)

		result := <-resultChan
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = testStore.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, acc1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, acc2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, acc1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, acc2.ID, toAccount.ID)

		// check account balance
		fmt.Println(">>>tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := acc1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - acc2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, exist, k)
		exist[k] = true
	}

	// check final updated balances
	updatedAcc1, err := testStore.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc1)

	updatedAcc2, err := testStore.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc2)

	fmt.Println(">>>after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, acc1.Balance-int64(n)*amount, updatedAcc1.Balance)
	require.Equal(t, acc2.Balance+int64(n)*amount, updatedAcc2.Balance)
}

func TestTransferTxDeadLock(t *testing.T) {
	testStore := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">>>before: ", acc1.Balance, acc2.Balance)

	n := 10
	amount := int64(10)

	errChan := make(chan error)

	for i := 0; i < n; i++ {
		fromAccID := acc1.ID
		toAccID := acc2.ID

		if i%2 == 1 {
			fromAccID = acc2.ID
			toAccID = acc1.ID
		}

		go func() {
			_, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccID,
				ToAccountID:   toAccID,
				Amount:        amount,
			})

			errChan <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	// check final updated balances
	updatedAcc1, err := testStore.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc1)

	updatedAcc2, err := testStore.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc2)

	fmt.Println(">>>after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, acc1.Balance, updatedAcc1.Balance)
	require.Equal(t, acc2.Balance, updatedAcc2.Balance)
}
