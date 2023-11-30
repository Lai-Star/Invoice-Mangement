import React, { Fragment, useState } from 'react';
import { useSelector, useStore } from 'react-redux';
import { Navigate, Route, Routes } from 'react-router-dom';
import { Backdrop, CircularProgress } from '@mui/material';

import NavigationBar from 'components/Layout/NavigationBar/NavigationBar';
import Sidebar from 'components/Layout/Sidebar/Sidebar';
import useMountEffect from 'hooks/useMountEffect';
import AccountsPage from 'pages/accounts';
import ExpensesPage from 'pages/expenses';
import GoalsPage from 'pages/goals';
import LogoutPage from 'pages/logout';
import SettingsPage from 'pages/settings';
import SubscriptionPage from 'pages/subscription';
import TransactionsPage from 'pages/transactions';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import fetchBankAccounts from 'shared/bankAccounts/actions/fetchBankAccounts';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import useFetchLinksIfNeeded from 'shared/links/hooks/useFetchLinksIfNeeded';
import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';
import fetchSpending from 'shared/spending/actions/fetchSpending';
import useFetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';
import InitialPlaidSetup from 'views/Setup/InitialPlaidSetup';

const AuthenticatedApp = (): JSX.Element => {
  const [loading, setLoading] = useState(true);
  const { dispatch, getState } = useStore();

  const [sidebarClosed, setSidebarClosed] = useState(true);

  const fetchInitialTransactionsIfNeeded = useFetchInitialTransactionsIfNeeded();
  const fetchLinksIfNeeded = useFetchLinksIfNeeded();

  useMountEffect(() => {
    Promise.all([
      fetchLinksIfNeeded(),
      fetchBankAccounts()(dispatch).then(() => Promise.all([
        fetchInitialTransactionsIfNeeded(),
        fetchFundingSchedulesIfNeeded()(dispatch, getState),
        fetchSpending()(dispatch, getState),
        fetchBalances()(dispatch, getState),
      ])),
    ])
      .finally(() => setLoading(false));
  });

  const hasAnyLinks = useSelector(getHasAnyLinks);

  if (loading) {
    return (
      <Backdrop open={ true }>
        <CircularProgress color="inherit" />
      </Backdrop>
    );
  }

  // If the user has no links setup then we want to only give them a handful of routes to get things setup.
  if (!hasAnyLinks) {
    return (
      <Routes>
        <Route path="/logout" element={ <LogoutPage /> } />
        <Route path="/setup" element={ <InitialPlaidSetup /> } />
        <Route path="/plaid/oauth-return" element={ <OAuthRedirect /> } />
        <Route path="*" element={ <Navigate replace to="/setup" /> } />
      </Routes>
    );
  }

  function toggleSidebar() {
    setSidebarClosed(!sidebarClosed);
  }

  return (
    <Fragment>
      <div className="flex h-full min-w-0 min-h-full">
        <Sidebar
          closed={ sidebarClosed }
          onToggleSidebar={ toggleSidebar }
          closeSidebar={ () => setSidebarClosed(true) }
        />
        <div className="relative flex flex-col flex-1 w-0 min-w-0 mb-8 lg:ml-64">
          <NavigationBar onToggleSidebar={ toggleSidebar } />
          <Routes>
            <Route path="/register" element={ <Navigate replace to="/" /> } />
            <Route path="/login" element={ <Navigate replace to="/" /> } />
            <Route path="/logout" element={ <LogoutPage /> } />
            <Route path="/transactions" element={ <TransactionsPage /> } />
            <Route path="/expenses" element={ <ExpensesPage /> } />
            <Route path="/goals" element={ <GoalsPage /> } />
            <Route path="/accounts" element={ <AccountsPage /> } />
            <Route path="/settings" element={ <SettingsPage /> } />
            <Route path="/subscription" element={ <SubscriptionPage /> } />
            <Route path="*" element={ <Navigate replace to="/transactions" /> } />
          </Routes>
        </div>
      </div>
    </Fragment>
  );
};

export default AuthenticatedApp;
