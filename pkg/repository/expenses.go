package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
	"time"
)

func (r *repositoryBase) GetSpending(bankAccountId uint64) ([]models.Spending, error) {
	var result []models.Spending
	err := r.txn.Model(&result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve spending")
	}

	return result, nil
}

func (r *repositoryBase) GetSpendingByFundingSchedule(bankAccountId, fundingScheduleId uint64) ([]models.Spending, error) {
	result := make([]models.Spending, 0)
	err := r.txn.Model(&result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."funding_schedule_id" = ?`, fundingScheduleId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses for funding schedule")
	}

	return result, nil
}

func (r *repositoryBase) CreateSpending(spending *models.Spending) error {
	spending.AccountId = r.AccountId()
	spending.DateCreated = time.Now().UTC()

	_, err := r.txn.Model(spending).Insert(spending)
	return errors.Wrap(err, "failed to create spending")
}

// UpdateExpenses should only be called with complete expense models. Do not use partial models with missing data for
// this action.
func (r *repositoryBase) UpdateExpenses(bankAccountId uint64, updates []models.Spending) error {
	for i := range updates {
		updates[i].AccountId = r.AccountId()
		updates[i].BankAccountId = bankAccountId
	}

	_, err := r.txn.Model(&updates).
		Update(&updates)
	if err != nil {
		return errors.Wrap(err, "failed to update expenses")
	}

	return nil
}

func (r *repositoryBase) GetSpendingById(bankAccountId, spendingId uint64) (*models.Spending, error) {
	var result models.Spending
	err := r.txn.Model(&result).
		Relation("FundingSchedule").
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_id" = ?`, spendingId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expense")
	}

	return &result, nil
}
