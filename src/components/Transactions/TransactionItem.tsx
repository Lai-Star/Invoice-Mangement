import { Button, Chip, Divider, ListItem, Typography } from "@material-ui/core";
import classnames from 'classnames';
import Spending from "data/Spending";
import Transaction from "data/Transaction";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import { getSpendingById } from "shared/spending/selectors/getSpendingById";
import selectTransaction from "shared/transactions/actions/selectTransaction";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";

import './styles/TransactionItem.scss';
import { getTransactionIsSelected } from "shared/transactions/selectors/getTransactionIsSelected";
import { ArrowDropDown } from "@material-ui/icons";

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  spending?: Spending;
  isSelected: boolean;
  selectTransaction: { (transactionId: number): void }
}

export class TransactionItem extends Component<WithConnectionPropTypes, {}> {

  getSpentFromString() {
    const { spending, transaction } = this.props;

    if (transaction.getIsAddition()) {
      return null;
    }

    if (!spending) {
      return (
        <Button className="w-full spending-selector">
          <div className="w-full flex justify-start">
            <div className="flex-auto flex justify-start">
              <span className="opacity-50 mr-1">
                Spent From
              </span>
              <span className="opacity-50">
                Safe-To-Spend
              </span>
            </div>
            <div className="flex-none spending-dropdown-icon">
              <ArrowDropDown/>
            </div>
          </div>
        </Button>
      );
    }

    return (
      <Button className="w-full spending-selector">
        <div className="w-full flex justify-start">
          <div className="flex-auto flex justify-start">
            <span className="opacity-50 mr-1">
              Spent From
            </span>
            <span>
              { spending.name }
            </span>
          </div>
          <div className="flex-none spending-dropdown-icon">
            <ArrowDropDown/>
          </div>
        </div>
      </Button>
    )
  }

  handleClick = () => {
    return this.props.selectTransaction(this.props.transactionId);
  }

  render() {
    const { transaction, isSelected } = this.props;

    return (
      <Fragment>
        <ListItem onClick={ this.handleClick } className={ classnames('transactions-item h-12', {
          'selected': isSelected,
        }) } role="transaction-row">
          <div className="w-full flex flex-row">
            <p
              className="flex-shrink w-2/5 transaction-item-name overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap font-semibold place-self-center"
            >
              { transaction.getTitle() }
            </p>

            <p
              className="flex-auto transaction-expense-name overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap"
            >
              { this.getSpentFromString() }
            </p>
            <div className="flex-none w-1/5 flex items-center">
              { transaction.isPending && <Chip label="Pending" className="align-middle self-center"/> }
              <div className="w-full flex justify-end">
                <Typography className={ classnames('amount align-middle self-center place-self-center', {
                  'addition': transaction.getIsAddition(),
                }) }>
                  <b>{ transaction.getAmountString() }</b>
                </Typography>
              </div>
            </div>
          </div>
        </ListItem>
        <Divider/>
      </Fragment>
    )
  }
}

export default connect(
  (state, props: PropTypes) => {
    const transaction = getTransactionById(props.transactionId)(state);
    const isSelected = getTransactionIsSelected(props.transactionId)(state);

    return {
      transaction,
      isSelected,
      spending: getSpendingById(transaction.spendingId)(state),
    }
  },
  {
    selectTransaction,
  }
)(TransactionItem)
