package models

import (
	"github.com/nleeper/goment"
	"github.com/pkg/errors"
	"time"
)

type Expense struct {
	tableName string `pg:"expenses"`

	ExpenseId              uint64           `json:"expenseId" pg:"expense_id,notnull,pk,type:'bigserial'"`
	AccountId              uint64           `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                *Account         `json:"-" pg:"rel:has-one"`
	BankAccountId          uint64           `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank,on_delete:CASCADE,type:'bigint'"`
	BankAccount            *BankAccount     `json:"bankAccount,omitempty" pg:"rel:has-one"`
	FundingScheduleId      *uint64          `json:"fundingScheduleId" pg:"funding_schedule_id,on_delete:SET NULL"`
	FundingSchedule        *FundingSchedule `json:"fundingSchedule,omitempty" pg:"rel:has-one"`
	Name                   string           `json:"name" pg:"name,notnull,unique:per_bank"`
	Description            string           `json:"description,omitempty" pg:"description"`
	TargetAmount           int64            `json:"targetAmount" pg:"target_amount,notnull,use_zero"`
	CurrentAmount          int64            `json:"currentAmount" pg:"current_amount,notnull,use_zero"`
	RecurrenceRule         *Rule            `json:"recurrenceRule" pg:"recurrence_rule,notnull,type:'text'"`
	LastRecurrence         *time.Time       `json:"lastRecurrence" pg:"last_recurrence,type:'date'"`
	NextRecurrence         time.Time        `json:"nextRecurrence" pg:"next_recurrence,notnull,type:'date'"`
	NextContributionAmount int64            `json:"nextContributionAmount" pg:"next_contribution_amount,notnull,use_zero"`
	IsBehind               bool             `json:"isBehind" pg:"is_behind,notnull,use_zero"`
}

func midnightInLocal(input time.Time, timezone *time.Location) time.Time {
	midnight := time.Date(
		input.Year(),  // Year
		input.Month(), // Month
		input.Day(),   // Day
		0,             // Hours
		0,             // Minutes
		0,             // Seconds
		0,             // Nano seconds
		timezone,      // The account's time zone.
	)

	return midnight
}

func (e *Expense) CalculateNextContribution(
	accountTimezone string,
	nextContributionDate time.Time,
	nextContributionRule *Rule,
) error {
	timezone, err := time.LoadLocation(accountTimezone)
	if err != nil {
		return errors.Wrap(err, "failed to parse account's timezone")
	}

	nextContributionDate = midnightInLocal(nextContributionDate, timezone)

	// If we have achieved our expense then we don't need to do anything.
	if e.TargetAmount <= e.CurrentAmount {
		e.IsBehind = false
		e.NextContributionAmount = 0
	}

	nextDueDate := midnightInLocal(e.NextRecurrence, timezone)
	if time.Now().After(nextDueDate) {
		e.LastRecurrence = &nextDueDate
		e.NextRecurrence = e.RecurrenceRule.After(nextDueDate, false)
		nextDueDate = midnightInLocal(e.NextRecurrence, timezone)
	}

	// If the next time we would contribute to this expense is after the next time the expense is due, then the expense
	// has fallen behind. Mark it as behind and set the contribution to be the difference.
	if nextContributionDate.After(nextDueDate) {
		e.IsBehind = true
		e.NextContributionAmount = e.TargetAmount - e.CurrentAmount
		return nil
	} else if nextContributionDate.Equal(nextDueDate) {
		// If the next time we would contribute is the same day it's due, this is okay. The user could change the due
		// date if they want a bit of a buffer and we would plan it differently. But we don't want to consider this
		// "behind".
		e.IsBehind = false
		e.NextContributionAmount = e.TargetAmount - e.CurrentAmount
		return nil
	}

	// If the next time we would contribute to this expense is not behind and has more than one contribution to meet its
	// target then we need to calculate a partial contribution.
	numberOfContributions := 0
	if nextContributionDate.Before(nextDueDate) {
		numberOfContributions++
	}
	moment, _ := goment.New(midnightInLocal(time.Now(), timezone))
	now := moment.StartOf("day").ToTime()
	nextContributionRule.DTStart(now)
	contributionDateX := nextContributionDate
	for {
		contributionDateX = nextContributionRule.After(contributionDateX, false)
		if nextDueDate.Before(contributionDateX) {
			break
		}

		numberOfContributions++
	}

	totalNeeded := e.TargetAmount - e.CurrentAmount
	perContribution := totalNeeded / int64(numberOfContributions)

	e.NextContributionAmount = perContribution
	return nil
}
