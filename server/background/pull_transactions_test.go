package background

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/plaid/plaid-go/v14/plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullTransactionsJob_Run(t *testing.T) {
	t.Run("invalid link", func(t *testing.T) {
		clock := clock.NewMock()
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    math.MaxInt64,
			Start:     clock.Now(),
			End:       clock.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.EqualError(t, err, "failed to retrieve link to pull transactions: failed to get link: pg: no rows in result set")
		assert.Equal(t, "ROLLBACK", hook.LastEntry().Message, "Should rollback")
	})

	t.Run("manual link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    link.LinkId,
			Start:     clock.Now(),
			End:       clock.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should not return an error, this way the job is not retried")

		// Make sure that if the job fails, nothing changes.
		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)
	})

	t.Run("no bank accounts", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    plaidLink.LinkId,
			Start:     clock.Now(),
			End:       clock.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "no error should be returned as the test should not be retried")
	})

	t.Run("missing access token", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		// Need at least one bank account.
		fixtures.GivenIHaveABankAccount(t, clock, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := platypus.NewPlaid(log, nil, nil, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		args := PullTransactionsArguments{
			AccountId: user.AccountId,
			LinkId:    plaidLink.LinkId,
			Start:     clock.Now(),
			End:       clock.Now().Add(-1 * 24 * time.Hour),
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.EqualError(t, err, "failed to retrieve access token for plaid link: failed to retrieve access token for plaid link: pg: no rows in result set")
	})

	t.Run("happy path, pull new transactions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		accessToken := gofakeit.UUID()
		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, accessToken))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Eq(accessToken),
				gomock.Eq(plaidLink.PlaidLink.ItemId),
			).
			Return(plaidClient, nil).
			Times(1)

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
			).
			Return([]platypus.BankAccount{
				platypus.PlaidBankAccount{
					AccountId: plaidBankAccount.PlaidAccountId,
					Balances: platypus.PlaidBankAccountBalances{
						Available: 100,
						Current:   100,
					},
					Mask:         plaidBankAccount.Mask,
					Name:         plaidBankAccount.Name,
					OfficialName: plaidBankAccount.PlaidOfficialName,
					Type:         "depository",
					SubType:      "checking",
				},
			}, nil).
			Times(1)

		plaidClient.EXPECT().
			GetAllTransactions(
				gomock.Any(),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.Eq([]string{
					plaidBankAccount.PlaidAccountId,
				}),
			).
			Return([]platypus.Transaction{
				platypus.PlaidTransaction{
					Amount:               1,
					BankAccountId:        plaidBankAccount.PlaidAccountId,
					Category:             []string{},
					Date:                 clock.Now(),
					ISOCurrencyCode:      "USD",
					IsPending:            false,
					MerchantName:         "Amazon",
					Name:                 "Amazon",
					OriginalDescription:  "AMZN MARKETPLACE",
					PendingTransactionId: nil,
					TransactionId:        gofakeit.UUID(),
				},
			}, nil).
			Times(1)

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		// Should not have any transactions before this job runs.
		count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.Zero(t, count, "should not have any transactions yet")

		{ // Do our first pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     clock.Now().Add(-1 * 24 * time.Hour),
				End:       clock.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count = fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.NotZero(t, count, "should have more than zero transactions now")
	})

	t.Run("happy path, pull new transactions subsequent", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		accessToken := gofakeit.UUID()
		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, accessToken))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Eq(accessToken),
				gomock.Eq(plaidLink.PlaidLink.ItemId),
			).
			Return(plaidClient, nil).
			Times(2)

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
			).
			Return([]platypus.BankAccount{
				platypus.PlaidBankAccount{
					AccountId: plaidBankAccount.PlaidAccountId,
					Balances: platypus.PlaidBankAccountBalances{
						Available: 100,
						Current:   100,
					},
					Mask:         plaidBankAccount.Mask,
					Name:         plaidBankAccount.Name,
					OfficialName: plaidBankAccount.PlaidOfficialName,
					Type:         "depository",
					SubType:      "checking",
				},
			}, nil).
			Times(2)

		transactions := []platypus.Transaction{
			platypus.PlaidTransaction{
				Amount:               1395,
				BankAccountId:        plaidBankAccount.PlaidAccountId,
				Category:             []string{},
				Date:                 clock.Now(),
				ISOCurrencyCode:      "USD",
				IsPending:            false,
				MerchantName:         "Hulu",
				Name:                 "Hulu",
				OriginalDescription:  "HULU",
				PendingTransactionId: nil,
				TransactionId:        gofakeit.UUID(),
			},
			platypus.PlaidTransaction{
				Amount:               1654,
				BankAccountId:        plaidBankAccount.PlaidAccountId,
				Category:             []string{},
				Date:                 clock.Now().Add(-1 * 24 * time.Hour),
				ISOCurrencyCode:      "USD",
				IsPending:            false,
				MerchantName:         "Amazon",
				Name:                 "Amazon",
				OriginalDescription:  "AMZN MARKETPLACE",
				PendingTransactionId: nil,
				TransactionId:        gofakeit.UUID(),
			},
		}

		plaidClient.EXPECT().
			GetAllTransactions(
				gomock.Any(),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.Eq([]string{
					plaidBankAccount.PlaidAccountId,
				}),
			).
			Return([]platypus.Transaction{
				transactions[0],
			}, nil)

		plaidClient.EXPECT().
			GetAllTransactions(
				gomock.Any(),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.Eq([]string{
					plaidBankAccount.PlaidAccountId,
				}),
			).
			Return(transactions, nil)

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		{ // Do our first pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     clock.Now().Add(-1 * 24 * time.Hour),
				End:       clock.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.NotZero(t, count, "should have more than zero transactions now")

		{ // Do our second pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     clock.Now().Add(-1 * 24 * time.Hour),
				End:       clock.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have two transactions now.
		count = fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.EqualValues(t, 2, count, "should have 2 transactions now")
	})

	t.Run("clear pending status on existing transaction", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		accessToken := gofakeit.UUID()
		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, accessToken))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Eq(accessToken),
				gomock.Eq(plaidLink.PlaidLink.ItemId),
			).
			Return(plaidClient, nil).
			Times(2)

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
			).
			Return([]platypus.BankAccount{
				platypus.PlaidBankAccount{
					AccountId: plaidBankAccount.PlaidAccountId,
					Balances: platypus.PlaidBankAccountBalances{
						Available: 100,
						Current:   100,
					},
					Mask:         plaidBankAccount.Mask,
					Name:         plaidBankAccount.Name,
					OfficialName: plaidBankAccount.PlaidOfficialName,
					Type:         "depository",
					SubType:      "checking",
				},
			}, nil).
			Times(2)

		end := clock.Now()
		start := end.Add(-2 * 24 * time.Hour)
		txn := mock_plaid.GenerateTransactions(t, start, end, 10, []string{
			plaidBankAccount.PlaidAccountId,
		})

		transactions := myownsanity.Map(txn, func(item plaid.Transaction) platypus.Transaction {
			result, err := platypus.NewTransactionFromPlaid(item)
			require.NoError(t, err, "must be able to create plaid transaction")
			return result
		})

		// Start with a pending transaction.
		testTxn := transactions[0].(platypus.PlaidTransaction)
		testTxn.IsPending = true
		transactions[0] = testTxn

		plaidClient.EXPECT().
			GetAllTransactions(
				gomock.Any(),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.Eq([]string{
					plaidBankAccount.PlaidAccountId,
				}),
			).
			Return(transactions, nil)

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		{ // Do our first pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     clock.Now().Add(-7 * 24 * time.Hour),
				End:       clock.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.NotZero(t, count, "should have more than zero transactions now")

		pendingCount := fixtures.CountPendingTransactions(t, user.AccountId)
		assert.EqualValues(t, 1, pendingCount, "there should be a single pending transaction")

		// Update the transactions returned from the API, now it will not be pending.
		testTxn.IsPending = false
		transactions[0] = testTxn

		plaidClient.EXPECT().
			GetAllTransactions(
				gomock.Any(),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.Eq([]string{
					plaidBankAccount.PlaidAccountId,
				}),
			).
			Return(transactions, nil)

		{ // Now try to retrieve a few more.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     clock.Now().Add(-7 * 24 * time.Hour),
				End:       clock.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		newCount := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.EqualValues(t, newCount, count, "must have more transactions after second run")

		pendingCount = fixtures.CountPendingTransactions(t, user.AccountId)
		assert.Zero(t, pendingCount, "there should not be a pending transaction anymore")
	})

	t.Run("closed account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		accessToken := gofakeit.UUID()
		require.NoError(t, provider.UpdateAccessTokenForPlaidLinkId(context.Background(), plaidLink.AccountId, plaidLink.PlaidLink.ItemId, accessToken))

		plaidBankAccount := fixtures.GivenIHaveABankAccount(t, clock, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fixtures.GivenIHaveABankAccount(t, clock, &plaidLink, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		plaidPlatypus := mockgen.NewMockPlatypus(ctrl)
		plaidClient := mockgen.NewMockClient(ctrl)
		plaidPlatypus.EXPECT().
			NewClient(
				gomock.Any(),
				gomock.AssignableToTypeOf(new(models.Link)),
				gomock.Eq(accessToken),
				gomock.Eq(plaidLink.PlaidLink.ItemId),
			).
			Return(plaidClient, nil).
			Times(1)

		plaidClient.EXPECT().
			GetAccounts(
				gomock.Any(),
			).
			Return([]platypus.BankAccount{
				platypus.PlaidBankAccount{
					AccountId: plaidBankAccount.PlaidAccountId,
					Balances: platypus.PlaidBankAccountBalances{
						Available: 100,
						Current:   100,
					},
					Mask:         plaidBankAccount.Mask,
					Name:         plaidBankAccount.Name,
					OfficialName: plaidBankAccount.PlaidOfficialName,
					Type:         "depository",
					SubType:      "checking",
				},
			}, nil).
			Times(1)

		plaidClient.EXPECT().
			GetAllTransactions(
				gomock.Any(),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.AssignableToTypeOf(time.Time{}),
				gomock.Eq([]string{
					plaidBankAccount.PlaidAccountId, // Does not include the second account.
				}),
			).
			Return([]platypus.Transaction{
				platypus.PlaidTransaction{
					Amount:               1,
					BankAccountId:        plaidBankAccount.PlaidAccountId,
					Category:             []string{},
					Date:                 clock.Now(),
					ISOCurrencyCode:      "USD",
					IsPending:            false,
					MerchantName:         "Amazon",
					Name:                 "Amazon",
					OriginalDescription:  "AMZN MARKETPLACE",
					PendingTransactionId: nil,
					TransactionId:        gofakeit.UUID(),
				},
			}, nil).
			Times(1)

		handler := NewPullTransactionsHandler(log, db, clock, provider, plaidPlatypus, publisher)

		// Should not have any transactions before this job runs.
		count := fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.Zero(t, count, "should not have any transactions yet")

		{ // Do our first pull of transactions.
			args := PullTransactionsArguments{
				AccountId: user.AccountId,
				LinkId:    plaidLink.LinkId,
				Start:     clock.Now().Add(-1 * 24 * time.Hour),
				End:       clock.Now(),
			}
			argsEncoded, err := DefaultJobMarshaller(args)
			assert.NoError(t, err, "must be able to marshal arguments")

			// Make sure that before we start there isn't anything in the database.
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

			err = handler.HandleConsumeJob(context.Background(), argsEncoded)
			assert.NoError(t, err, "must process job successfully")
		}

		// We should have a few transactions now.
		count = fixtures.CountNonDeletedTransactions(t, user.AccountId)
		assert.NotZero(t, count, "should have more than zero transactions now")
	})
}
