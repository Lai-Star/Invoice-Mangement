import FundingSchedule from "models/FundingSchedule";
import { Map } from 'immutable';
import { createSelector } from "reselect";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";

const fundingSchedulesByBankAccount = state => state.fundingSchedules.items;

export const getFundingSchedules = createSelector<any, any, Map<number, FundingSchedule>>(
  [getSelectedBankAccountId, fundingSchedulesByBankAccount],
  (selectedBankAccountId, fundingSchedulesByBank) => {
    return fundingSchedulesByBank.get(selectedBankAccountId, Map<number, FundingSchedule>());
  },
);
