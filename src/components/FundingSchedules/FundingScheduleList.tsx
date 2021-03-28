import { Button, List, ListItem, ListItemText } from "@material-ui/core";
import NewFundingScheduleDialog from "components/FundingSchedules/NewFundingScheduleDialog";
import FundingSchedule from "data/FundingSchedule";
import { Map } from 'immutable';
import React, { Component } from "react";
import { connect } from "react-redux";

export interface PropTypes {
  onHide: { (): void }
}

interface WithConnectionPropTypes extends PropTypes {
  fundingSchedules: Map<number, FundingSchedule>;
}


interface State {
  newFundingScheduleDialogOpen: boolean;
}

export class FundingScheduleList extends Component<WithConnectionPropTypes, State> {

  state = {
    newFundingScheduleDialogOpen: false,
  };

  openNewFundingScheduleDialog = () => {
    return this.setState({
      newFundingScheduleDialogOpen: true,
    });
  };

  closeFundingScheduleDialog = () => {
    return this.setState({
      newFundingScheduleDialogOpen: false,
    });
  };

  render() {
    const { fundingSchedules, onHide } = this.props;
    return (
      <div className="w-full funding-schedule-list">
        <NewFundingScheduleDialog onClose={ this.closeFundingScheduleDialog }
                                  isOpen={ this.state.newFundingScheduleDialogOpen }/>
        <div className="w-full p-5">
          <Button onClick={ onHide }>
            Back
          </Button>
          <Button onClick={ this.openNewFundingScheduleDialog }>
            New Funding Schedule
          </Button>
        </div>
        <List className="w-full">
          {
            fundingSchedules.map(schedule => (
              <ListItem key={ schedule.fundingScheduleId } button>
                <ListItemText>
                  { schedule.name }
                </ListItemText>
              </ListItem>
            )).toArray()
          }
        </List>
      </div>
    )
  }
}

export default connect(
  state => ({
    fundingSchedules: Map<number, FundingSchedule>(),
  }),
  {}
)(FundingScheduleList);
