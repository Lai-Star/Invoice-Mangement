import { Checkbox, ListItem, ListItemIcon, Typography } from "@material-ui/core";
import classnames from 'classnames';
import Spending from "data/Spending";
import Transaction from "data/Transaction";
import React, { Component } from 'react';
import { connect } from "react-redux";
import { getSpendingById } from "shared/spending/selectors/getSpendingById";
import selectTransaction from "shared/transactions/actions/selectTransaction";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";

import './styles/TransactionItem.scss';
import { getTransactionIsSelected } from "shared/transactions/selectors/getTransactionIsSelected";

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

  getSpentFromString(): string {
    const { spending, transaction } = this.props;

    if (transaction.getIsAddition()) {
      return 'Deposited Into Safe-To-Spend';
    }

    if (!spending) {
      return 'Spent From Safe-To-Spend';
    }

    return `Spent From ${ spending.name }`;
  }

  handleClick = () => {
    return this.props.selectTransaction(this.props.transactionId);
  }

  render() {
    const { transaction, isSelected } = this.props;

    return (
      <ListItem button onClick={ this.handleClick } className="transactions-item" role="transaction-row">
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ isSelected }
            tabIndex={ -1 }
            color="primary"
          />
        </ListItemIcon>
        <div className="grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full">
          <div className="col-span-3">
            <Typography className="transaction-item-name">{ transaction.getName() }</Typography>
          </div>
          <div className="col-span-3 opacity-75">
            <Typography className="transaction-expense-name">{ this.getSpentFromString() }</Typography>
          </div>
          <div className="row-span-2 col-span-1 flex justify-end">
            <Typography className={ classnames('amount align-middle self-center', {
              'addition': transaction.getIsAddition(),
            }) }>
              { transaction.getAmountString() }
            </Typography>
          </div>
        </div>
      </ListItem>
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
