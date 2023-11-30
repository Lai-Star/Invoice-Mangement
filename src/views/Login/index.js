import React, {Component} from 'react';
import {connect} from 'react-redux';
import {bindActionCreators} from "redux";
import {withRouter} from 'react-router-dom';
import PropTypes from "prop-types";
import {
  Box,
  Button,
  Card, CardActions,
  CardContent,
  CardHeader,
  Container,
  Grid, Grow,
  Paper,
  TextField,
  Typography
} from "@material-ui/core";
import {Formik, Form, Field, ErrorMessage} from 'formik';

import './styles/login.scss';
import {Link as RouterLink} from 'react-router-dom'
import ReCAPTCHA from "react-google-recaptcha";
import {getReCAPTCHAKey, getShouldVerifyLogin, getSignUpAllowed} from "../../shared/bootstrap/selectors";
import request from "../../shared/util/request";
import bootstrapLogin from "../../shared/authentication/actions/bootstrapLogin";

export class LoginView extends Component {
  static propTypes = {
    allowSignUp: PropTypes.bool.isRequired,
    verifyLogin: PropTypes.bool.isRequired,
    ReCAPTCHAKey: PropTypes.string,
    bootstrapLogin: PropTypes.func.isRequired,
    history: PropTypes.instanceOf(History).isRequired,
  };

  submitLogin = values => {
    console.log(values);
    return request().post('/api/authentication/login', {
      email: values.email,
      password: values.password,
    })
      .then(result => {
        return this.props.bootstrapLogin(result.data.token, result.data.user);
      })
      .then(() => {
        this.props.history.push('/');
      })
      .catch(error => {
        alert(error);
      });
  };

  renderReCAPTCHAMaybe = () => {
    const {verifyLogin, ReCAPTCHAKey} = this.props;
    if (!verifyLogin) {
      return null;
    }

    return (
      <Grid item xs={12}>
        <ReCAPTCHA
          sitekey={ReCAPTCHAKey}
          onChange={value => this.setState({verification: value})}
        />
      </Grid>
    )
  };

  render() {
    return (
      <div className="login-view">
        <Formik
          initialValues={{
            email: '',
            password: '',
          }}
          onSubmit={(values, {setSubmitting}) => {
            this.submitLogin(values)
              .finally(() => setSubmitting(false));
          }}
        >
          {({
              values,
              errors,
              touched,
              handleChange,
              handleBlur,
              handleSubmit,
              isSubmitting,
              /* and other goodies */
            }) => (
            <Box m={6}>
              <Container maxWidth="xs" className={"login-root"}>
                <Grow in>
                  <Card>
                    <CardHeader title="Harder Than It Needs To Be" subheader="Login"/>
                    <CardContent>
                      <Grid container spacing={1}>
                        <Grid item xs={12}>
                          <TextField
                            fullWidth
                            id="email"
                            label="Email"
                            name="email"
                            value={values.email}
                            onChange={handleChange}
                            error={touched.email && !!errors.email}
                            helperText={touched.email && errors.email}
                            disabled={isSubmitting}
                          />
                        </Grid>
                        <Grid item xs={12}>
                          <TextField
                            fullWidth
                            id="password"
                            label="Password"
                            name="password"
                            type="password"
                            value={values.password}
                            onChange={handleChange}
                            error={touched.password && !!errors.password}
                            helperText={touched.password && errors.password}
                            disabled={isSubmitting}
                          />
                        </Grid>
                        {this.renderReCAPTCHAMaybe()}
                      </Grid>
                    </CardContent>
                    <CardActions>
                      <div style={{marginLeft: 'auto'}}/>
                      {this.props.allowSignUp &&
                      <Button
                        to="/register"
                        component={RouterLink}
                        variant="outlined"
                        color="secondary"
                      >
                        Sign Up
                      </Button>
                      }
                      <Button onClick={handleSubmit} variant="outlined" color="primary">Login</Button>
                    </CardActions>
                  </Card>
                </Grow>
              </Container>
            </Box>
          )}
        </Formik>
      </div>
    )
  }
}

export default connect(
  state => ({
    allowSignUp: getSignUpAllowed(state),
    verifyLogin: getShouldVerifyLogin(state),
    ReCAPTCHAKey: getReCAPTCHAKey(state),
  }),
  dispatch => bindActionCreators({
    bootstrapLogin,
  }, dispatch),
)(withRouter(LoginView));
