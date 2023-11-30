import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { LocalizationProvider } from '@mui/lab';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { renderHook, RenderHookResult, WrapperComponent } from '@testing-library/react-hooks';
import { createMemoryHistory, MemoryHistory } from 'history';
import { IconVariant, SnackbarProvider } from 'notistack';
import React from 'react';
import { Provider } from 'react-redux';
import { unstable_HistoryRouter as HistoryRouter } from 'react-router-dom';
import { AppStore, configureStore } from 'store';
import theme from 'theme';

export interface HooksOptions {
  store?: AppStore;
  history?: MemoryHistory;
}

function testRenderHook<TProps, TResult>(callback: (props: TProps) => TResult, options?: HooksOptions): RenderHookResult<TProps, TResult> {
  const store = options?.store || configureStore();
  const history = options?.history || createMemoryHistory();

  const Wrapper: WrapperComponent<TProps> = (props: React.PropsWithChildren<any>) => {
    const snackbarIcons: Partial<IconVariant> = {
      error: <ErrorIcon className="mr-2.5"/>,
      success: <DoneIcon className="mr-2.5"/>,
      warning: <WarningIcon className="mr-2.5"/>,
      info: <InfoIcon className="mr-2.5"/>,
    };

    return (
      <Provider store={ store }>
        <HistoryRouter history={ history }>
          <ThemeProvider theme={ theme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                <CssBaseline/>
                { props.children }
              </SnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </HistoryRouter>
      </Provider>
    );
  }

  return renderHook<TProps, TResult>(callback, {
    wrapper: Wrapper,
  })
}

export default testRenderHook;
