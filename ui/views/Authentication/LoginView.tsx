import TextWithLine from 'components/TextWithLine';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { Link as RouterLink, RouteComponentProps, withRouter } from 'react-router-dom';
import User from 'models/User';
import bootstrapLogin from 'shared/authentication/actions/bootstrapLogin';
import request from 'shared/util/request';
import { getReCAPTCHAKey, getShouldVerifyLogin, getSignUpAllowed } from 'shared/bootstrap/selectors';
import classnames from 'classnames';
import {
  Alert,
  AlertTitle,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Snackbar,
  TextField
} from '@mui/material';
import { Formik, FormikHelpers, FormikValues } from 'formik';
import { AppState } from 'store';
import verifyEmailAddress from 'util/verifyEmailAddress';
import CaptchaMaybe from 'views/Captcha/CaptchaMaybe';

import Logo from 'assets';

interface LoginValues {
  email: string | null;
  password: string | null;
}

interface State {
  error: string | null;
  loading: boolean;
  verification: string | null;
  resendEmailAddress: string | null;
}

interface WithConnectionPropTypes extends RouteComponentProps {
  ReCAPTCHAKey: string | null;
  bootstrapLogin: (token: string, user: User) => Promise<void>;
  verifyLogin: boolean;
  allowSignUp: boolean;
}

class LoginView extends Component<WithConnectionPropTypes, State> {

  state = {
    error: null,
    loading: false,
    verification: null,
    resendEmailAddress: null,
  };

  renderErrorMaybe = (): React.ReactNode | null => {
    const { error } = this.state;
    if (!error) {
      return null;
    }

    return (
      <Snackbar open={ !!error } autoHideDuration={ 10000 } onClose={ () => this.setState({ error: null }) }>
        <Alert variant="filled" severity="error">
          <AlertTitle>Error</AlertTitle>
          { this.state.error }
        </Alert>
      </Snackbar>
    );
  };

  validateInput = (values: LoginValues): Partial<LoginValues> => {
    let errors: Partial<LoginValues> = {};

    if (values.email) {
      if (!verifyEmailAddress(values.email)) {
        errors['email'] = 'Please provide a valid email address.';
      }
    }

    if (values.password) {
      if (values.password.length < 8) {
        errors['password'] = 'Password must be at least 8 characters long.'
      }
    }

    return errors;
  };

  submit = (values: LoginValues, helpers: FormikHelpers<LoginValues>): Promise<void> => {
    helpers.setSubmitting(true);
    this.setState({
      error: null,
      loading: true,
    });

    return request()
      .post('/authentication/login', {
        captcha: this.state.verification,
        email: values.email,
        password: values.password,
      })
      .then(result => {
        return this.props.bootstrapLogin(result.data.token, result.data.user)
          .then(() => {
            if (result.data.nextUrl) {
              this.props.history.push(result.data.nextUrl);
              return
            }

            this.props.history.push('/');
          });
      })
      .catch(error => {
        const errorMessage = error?.response?.data?.error || 'Failed to authenticate.'

        switch (error?.response?.status) {
          case 428: // Email not verified.
            return this.props.history.push('/verify/email/resend', {
              'emailAddress': values.email,
            });
          case 403: // Invalid login.
            return this.setState({
              error: errorMessage,
              loading: false,
            });
        }

        return this.setState({
          error: errorMessage,
          loading: false,
        });
      })
      .finally(() => helpers.setSubmitting(false));
  };

  renderResendVerificationDialogMaybe = (): React.ReactNode => {
    const { resendEmailAddress } = this.state;

    if (!resendEmailAddress) {
      return null;
    }

    const closeDialog = () => this.setState({
      resendEmailAddress: null,
    });

    return (
      <Dialog
        open
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">Resend Email Verification Link</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            It looks like your email address has not been verified. Do you want to resend the email verification link?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={ closeDialog } color="secondary">
            Cancel
          </Button>
          <Button onClick={ closeDialog } color="primary" autoFocus>
            Resend
          </Button>
        </DialogActions>
      </Dialog>
    )
  };

  renderBottomButtons = (
    isSubmitting: boolean,
    disableForVerification: boolean,
    values: LoginValues,
    submitForm: () => Promise<any>,
  ): React.ReactNode => {
    return (
      <div>
        <div className="w-full pt-2.5 pb-2.5">
          <Button
            className="w-full"
            color="primary"
            disabled={ isSubmitting || (!values.password || !values.email || !disableForVerification) }
            onClick={ submitForm }
            type="submit"
            variant="contained"
          >
            { isSubmitting && <CircularProgress
              className={ classnames('mr-2', {
                'opacity-50': isSubmitting,
              }) }
              size="1em"
              thickness={ 5 }
            /> }
            { isSubmitting ? 'Signing In...' : 'Sign In' }
          </Button>
        </div>
      </div>
    )
  };

  render() {
    const initialValues: LoginValues = {
      email: '',
      password: '',
    }

    const disableForVerification = !this.props.verifyLogin || (this.props.ReCAPTCHAKey && this.state.verification);

    return (
      <Fragment>
        { this.renderResendVerificationDialogMaybe() }
        { this.renderErrorMaybe() }
        <Formik
          initialValues={ initialValues }
          validate={ this.validateInput }
          onSubmit={ this.submit }
        >
          { ({
               values,
               errors,
               touched,
               handleChange,
               handleBlur,
               handleSubmit,
               isSubmitting,
               submitForm,
             }) => (
            <form onSubmit={ handleSubmit } className="h-full overflow-y-auto">
              <div className="flex items-center justify-center w-full h-full max-h-full">
                <div className="w-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
                  <div className="flex justify-center w-full mb-5">
                    <img src={ Logo } className="w-1/3"/>
                  </div>
                  { this.props.allowSignUp && (
                    <div>
                      <div className="w-full pb-2.5">
                        <Button
                          className="w-full"
                          color="secondary"
                          component={ RouterLink }
                          disabled={ isSubmitting }
                          to="/register"
                          variant="contained"
                        >
                          Sign Up For monetr
                        </Button>
                      </div>
                      <div className="w-full opacity-50 pb-2.5">
                        <TextWithLine>
                          or sign in with your email
                        </TextWithLine>
                      </div>
                    </div>
                  ) }
                  <div className="w-full">
                    <div className="w-full pb-2.5">
                      <TextField
                        autoComplete="username"
                        autoFocus
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.email && !!errors.email }
                        helperText={ (touched.email && errors.email) ? errors.email : null }
                        id="login-email"
                        label="Email"
                        name="email"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        value={ values.email }
                        variant="outlined"
                      />
                    </div>
                    <div className="w-full pt-2.5 pb-2.5">
                      <TextField
                        autoComplete="current-password"
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.password && !!errors.password }
                        helperText={ (touched.password && errors.password) ? errors.password : null }
                        id="login-password"
                        label="Password"
                        name="password"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        type="password"
                        value={ values.password }
                        variant="outlined"
                      />
                    </div>
                  </div>
                  <CaptchaMaybe
                    loading={ isSubmitting }
                    show={ this.props.verifyLogin }
                    onVerify={ (value) => this.setState({
                      verification: value,
                    }) }
                  />
                  { this.renderBottomButtons(isSubmitting, disableForVerification, values, submitForm) }
                </div>
              </div>
            </form>
          ) }
        </Formik>
      </Fragment>
    )
  }
}

export default connect(
  (state: AppState) => ({
    ReCAPTCHAKey: getReCAPTCHAKey(state),
    allowSignUp: getSignUpAllowed(state),
    verifyLogin: getShouldVerifyLogin(state),
  }),
  {
    bootstrapLogin,
  }
)(withRouter(LoginView));
