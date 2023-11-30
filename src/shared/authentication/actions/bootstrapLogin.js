import User from "data/User";
import request from "shared/util/request";
import { BOOTSTRAP_LOGIN } from "shared/authentication/actions";
import { getAPIUrl } from "shared/bootstrap/selectors";
import axios from "axios";
import Cookies from 'js-cookie'

export default function bootstrapLogin(token = null, user = null) {
  return (dispatch, getState) => {
    if (token) {
      // Trying to switch over to using cookies, but I don't want to break anything at the moment.
      try {
        Cookies.set('M-Token', token, {
          // TODO Make the cookie domain a configuration variable.
          domain: '.staging.monetr.dev',
          secure: true,
          path: '/',
        });
      } catch (e) {
        console.error(e);
      }

      window.localStorage.setItem('H-Token', token);
    } else {
      token = window.localStorage.getItem('H-Token');
    }

    // If the token is not present at this point then the user is not authenticated. We want to dispatch accordingly and
    // store in redux that the user is not authenticated.
    if (!token) {
      dispatch({
        type: BOOTSTRAP_LOGIN,
        payload: {
          isAuthenticated: false,
          token: null,
          user: null,
        }
      });
      return Promise.resolve();
    }

    const apiUrl = getAPIUrl(getState());
    window.API = axios.create({
      baseURL: apiUrl,
      headers: {
        'H-Token': token,
      },
    });

    if (!user) {
      // If we do have the token but we don't have the user info then we need to retrieve it using an API call to get
      // our user data from the API.
      return request().get('/users/me')
        .then(result => {
          dispatch({
            type: BOOTSTRAP_LOGIN,
            payload: {
              isAuthenticated: true,
              token: token,
              user: new User(result.data.user),
            }
          })
        })
        .catch(error => {
          window.localStorage.removeItem('H-Token');
          console.error(error);
        });
    }

    dispatch({
      type: BOOTSTRAP_LOGIN,
      payload: {
        isAuthenticated: true,
        token: token,
        user: new User(user),
      }
    });
    return Promise.resolve();
  };
}
