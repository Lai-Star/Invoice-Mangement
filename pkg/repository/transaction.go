package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

type TransactionUpdateId struct {
	TransactionId uint64 `pg:"transaction_id"`
	BankAccountId uint64 `pg:"bank_account_id"`
	Amount        int64  `pg:"amount"`
}

func (r *repositoryBase) InsertTransactions(transactions []models.Transaction) error {
	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
	}
	_, err := r.txn.Model(&transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to insert transactions")
}

func (r *repositoryBase) GetTransactionsByPlaidId(linkId uint64, plaidTransactionIds []string) (map[string]TransactionUpdateId, error) {
	type WithPlaidId struct {
		tableName          string `pg:"transactions"`
		plaidTransactionId string `pg:"plaid_transaction_id"`
		TransactionUpdateId
	}
	var items []WithPlaidId
	err := r.txn.Model(&items).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."bank_account_id" = "transaction"."bank_account_id" AND "bank_account"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		WhereIn(`"transaction"."plaid_transaction_id" IN (?)`, plaidTransactionIds).
		Select(&items)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transaction Ids for plaid Ids")
	}

	result := map[string]TransactionUpdateId{}
	for _, item := range items {
		result[item.plaidTransactionId] = item.TransactionUpdateId
	}

	return result, nil
}
