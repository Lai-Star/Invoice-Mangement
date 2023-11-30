import React, {Component} from "react";
import {Box, Button, CircularProgress, Container, Grid, Grow, Paper, Typography} from "@material-ui/core";
import {bindActionCreators} from "redux";
import {withRouter} from "react-router-dom";
import {connect} from "react-redux";
import request from "../../shared/util/request";
import {PlaidLink} from "react-plaid-link";
import PropTypes from "prop-types";
import logout from "../../shared/authentication/actions/logout";
import {PlaidConnectButton} from "./hookyBoi";


export class FirstTimeSetup extends Component {
  state = {
    loading: true,
    linkToken: '',
  };

  static propTypes = {
    logout: PropTypes.func.isRequired,
  }

  componentDidMount() {
    request()
      .get('/api/plaid/link/token/new')
      .then(result => {
        this.setState({
          loading: false,
          linkToken: result.data.linkToken,
        });
      })
      .catch(error => {
        alert(error);
      });
  }

  doCancel = () => {
    this.props.logout();
  }

  plaidLinkSuccess = result => {

  };

  renderPlaidLink = () => {
    const {loading, linkToken} = this.state;
    if (loading) {
      return <CircularProgress style={{float: 'right'}}/>;
    }

    if (linkToken.length > 0) {
      return (
        <PlaidConnectButton token={linkToken} onSuccess={this.plaidLinkSuccess}/>
      )
    }

    return <Typography>Something went wrong...</Typography>;
  };


  render() {

    return (
      <Box m={12}>
        <Container maxWidth="sm">
          <Grow in>
            <Paper elevation={3}>
              <Box m={4}>
                <Grid container spacing={4}>
                  <Grid item xs={12}>
                    <Typography variant="h5">Welcome to Harder Than It Needs To Be</Typography>
                    <Typography>To continue, you'll need to link your bank account.</Typography>
                    <Typography>If you would not like to do this, click cancel and you will be logged out.</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Button variant="outlined" onClick={this.doCancel}>Cancel</Button>
                  </Grid>
                  <Grid item xs={6}>
                    {this.renderPlaidLink()}
                  </Grid>
                </Grid>
              </Box>
            </Paper>
          </Grow>
        </Container>
      </Box>
    );
  }
}

export default connect(
  state => ({}),
  dispatch => bindActionCreators({
    logout,
  }, dispatch),
)(withRouter(FirstTimeSetup));
