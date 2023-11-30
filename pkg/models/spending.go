package models

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/util"
	"github.com/pkg/errors"
)

type SpendingType uint8

const (
	SpendingTypeExpense SpendingType = iota
	SpendingTypeGoal
)

type Spending struct {
	tableName string `pg:"spending"`

	SpendingId             uint64           `json:"spendingId" pg:"spending_id,notnull,pk,type:'bigserial'"`
	AccountId              uint64           `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                *Account         `json:"-" pg:"rel:has-one"`
	BankAccountId          uint64           `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank,on_delete:CASCADE,type:'bigint'"`
	BankAccount            *BankAccount     `json:"bankAccount,omitempty" pg:"rel:has-one" swaggerignore:"true"`
	FundingScheduleId      uint64           `json:"fundingScheduleId" pg:"funding_schedule_id,notnull,on_delete:RESTRICT"`
	FundingSchedule        *FundingSchedule `json:"-" pg:"rel:has-one" swaggerignore:"true"`
	SpendingType           SpendingType     `json:"spendingType" pg:"spending_type,notnull,use_zero,unique:per_bank"`
	Name                   string           `json:"name" pg:"name,notnull,unique:per_bank"`
	Description            string           `json:"description,omitempty" pg:"description"`
	TargetAmount           int64            `json:"targetAmount" pg:"target_amount,notnull,use_zero"`
	CurrentAmount          int64            `json:"currentAmount" pg:"current_amount,notnull,use_zero"`
	UsedAmount             int64            `json:"usedAmount" pg:"used_amount,notnull,use_zero"`
	RecurrenceRule         *Rule            `json:"recurrenceRule" pg:"recurrence_rule,type:'text'" swaggertype:"string"`
	LastRecurrence         *time.Time       `json:"lastRecurrence" pg:"last_recurrence"`
	NextRecurrence         time.Time        `json:"nextRecurrence" pg:"next_recurrence,notnull"`
	NextContributionAmount int64            `json:"nextContributionAmount" pg:"next_contribution_amount,notnull,use_zero"`
	IsBehind               bool             `json:"isBehind" pg:"is_behind,notnull,use_zero"`
	IsPaused               bool             `json:"isPaused" pg:"is_paused,notnull,use_zero"`
	DateCreated            time.Time        `json:"dateCreated" pg:"date_created,notnull"`
}

func (e Spending) GetProgressAmount() int64 {
	switch e.SpendingType {
	case SpendingTypeGoal:
		return e.CurrentAmount + e.UsedAmount
	case SpendingTypeExpense:
		fallthrough
	default:
		return e.CurrentAmount
	}
}

func (e *Spending) CalculateNextContribution(
	ctx context.Context,
	accountTimezone string,
	nextContributionDate time.Time,
	nextContributionRule *Rule,
) error {
	span := sentry.StartSpan(ctx, "CalculateNextContribution")
	defer span.Finish()

	span.SetTag("spendingId", strconv.FormatUint(e.SpendingId, 10))

	timezone, err := time.LoadLocation(accountTimezone)
	if err != nil {
		return errors.Wrap(err, "failed to parse account's timezone")
	}

	// The total needed needs to be calculated differently for goals and expenses. How much expenses need is always a
	// representation of the target amount minus the current amount allocated to the expense. But goals work a bit
	// differently because the allocated amount can fluctuate throughout the life of the goal. When a transaction is
	// spent from a goal it deducts from the current amount, but adds to the used amount. This is to keep track of how
	// much the goal has actually progressed while maintaining existing patterns for calculating allocations. As a
	// result for us to know how much a goal needs, we need to subtract the current amount plus the used amount from the
	// target for goals.
	progressAmount := e.GetProgressAmount()

	nextContributionDate = util.MidnightInLocal(nextContributionDate, timezone)

	// If we have achieved our expense then we don't need to do anything.
	if e.TargetAmount <= progressAmount {
		e.IsBehind = false
		e.NextContributionAmount = 0
	}

	nextDueDate := util.MidnightInLocal(e.NextRecurrence, timezone)
	if time.Now().After(nextDueDate) && e.RecurrenceRule != nil {
		e.LastRecurrence = &nextDueDate
		e.NextRecurrence = util.MidnightInLocal(e.RecurrenceRule.After(nextDueDate, false), timezone)
		nextDueDate = util.MidnightInLocal(e.NextRecurrence, timezone)
	}

	needed := int64(math.Max(float64(e.TargetAmount-progressAmount), 0))

	// If the next time we would contribute to this expense is after the next time the expense is due, then the expense
	// has fallen behind. Mark it as behind and set the contribution to be the difference.
	if nextContributionDate.After(nextDueDate) {
		e.NextContributionAmount = needed
		e.IsBehind = progressAmount < e.TargetAmount
		return nil
	} else if nextContributionDate.Equal(nextDueDate) {
		// If the next time we would contribute is the same day it's due, this is okay. The user could change the due
		// date if they want a bit of a buffer and we would plan it differently. But we don't want to consider this
		// "behind".
		e.IsBehind = false
		e.NextContributionAmount = needed
		return nil
	} else if progressAmount >= e.TargetAmount {
		e.IsBehind = false
	} else {
		// Fix weird edge case where this isn't being unset.
		e.IsBehind = false
	}

	// TODO Handle expenses that recur more frequently than they are funded.
	nowInTimezone := time.Now().In(timezone)
	nextContributionRule.DTStart(nextContributionDate)
	numberOfContributions := len(nextContributionRule.Between(nowInTimezone, nextDueDate, false))

	if numberOfContributions == 0 {
		// This is a bit weird, I'm not sure what causes this yet off the top of my head. I ran into while testing when
		// I made the due date the 29th, and the next payday the 28th. (Note: it was the 28th at the time). And it
		// caused this to break. Pretty sure this is just a bug with the funding schedule next contribution date being
		// in the past, but this is a short term fix for now. This also acts as a slight safety net for a divide by 0
		// error.
		e.NextContributionAmount = needed
	} else {
		perContribution := needed / int64(numberOfContributions)
		e.NextContributionAmount = perContribution
	}

	return nil
}
