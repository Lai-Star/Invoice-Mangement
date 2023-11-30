import { Map, OrderedMap } from 'immutable';
import Transaction from 'models/Transaction';

export default class TransactionState {
  items: Map<number, OrderedMap<number, Transaction>>;
  loaded: boolean;
  loading: boolean;
  selectedTransactionId?: number;

  constructor() {
    this.items = Map<number, OrderedMap<number, Transaction>>();
  }
}
