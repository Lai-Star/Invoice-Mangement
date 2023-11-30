import { Chip, Divider, ListItem, Paper, Popover, Typography } from "@material-ui/core";
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
import SelectButton from "components/SelectyBoi/SelectButton";
import SpendingSelectionList from "components/Spending/SpendingSelectionList";
import updateTransaction from "shared/transactions/actions/updateTransaction";

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  spending?: Spending;
  isSelected: boolean;
  selectTransaction: { (transactionId: number): void }
  updateTransaction: (transaction: Transaction) => Promise<void>;
}

interface State {
  anchorEl: Element | null;
  width: number | null;
}

export class TransactionItem extends Component<WithConnectionPropTypes, State> {

  state = {
    anchorEl: null,
    width: 0,
  };

  getSpentFromString() {
    const { spending, transaction, updateTransaction } = this.props;
    const { anchorEl } = this.state;

    if (transaction.getIsAddition()) {
      return null;
    }

    const updateSpentFrom = (selection: Spending | null) => {
      const spendingId = selection ? selection.spendingId : null;

      if (spendingId === transaction.spendingId) {
        return Promise.resolve();
      }

      const updatedTransaction = new Transaction({
        ...transaction,
        spendingId: spendingId,
      });

      return updateTransaction(updatedTransaction)
        .catch(error => alert(error));
    };

    const openPopover = (event: { currentTarget: Element }) => {
      this.setState({
        anchorEl: event.currentTarget,
        width: event.currentTarget.clientWidth,
      });
    };

    return (
      <Fragment>
        <SelectButton
          open={ Boolean(anchorEl) }
          onClick={ openPopover }
        >
          <span className="opacity-50 mr-1">
            Spent From
          </span>
          <span className={ classnames('overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap', {
            'opacity-50': !spending,
          }) }>
            { spending ? spending.name : 'Safe-To-Spend' }
          </span>
        </SelectButton>
        <Popover
          id={ `transaction-spent-from-popover-${ transaction.transactionId }` }
          open={ Boolean(anchorEl) }
          anchorEl={ anchorEl }
          onClose={ () => this.setState({ anchorEl: null }) }
          anchorOrigin={ {
            vertical: 'bottom',
            horizontal: 'left',
          } }
          transformOrigin={ {
            vertical: 'top',
            horizontal: 'left',
          } }
        >
          <Paper style={ { width: `${ this.state.width }px` } } className="min-w-96 max-h-96 p-0 overflow-auto">
            <SpendingSelectionList value={ transaction.spendingId } onChange={ updateSpentFrom }/>
          </Paper>
        </Popover>
      </Fragment>
    )
  }



  handleClick = () => {
    return this.props.selectTransaction(this.props.transactionId);
  }

  render() {
    const { transaction, isSelected } = this.props;

    return (
      <Fragment>
        <ListItem className={ classnames('transactions-item h-12', {
          'selected': false,
        }) } role="transaction-row">
          <div className="w-full flex flex-row">
            <p
              className="flex-shrink w-2/5 transaction-item-name overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap font-semibold place-self-center pr-1"
            >
              <SelectButton>
                <span className={ classnames('overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap') }>
                { transaction.getTitle() }
                </span>
              </SelectButton>
            </p>

            <p
              className="flex-auto transaction-expense-name overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap pr-1"
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
    updateTransaction,
  }
)(TransactionItem)
