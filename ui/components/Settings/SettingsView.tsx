import React, { useState } from 'react';
import { Info, Security, Settings } from '@mui/icons-material';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';

import AboutView from 'components/About/AboutView/AboutView';
import ChangePassword from 'components/Settings/ChangePassword';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
  className?: string;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={ value !== index }
      id={ `simple-tabpanel-${ index }` }
      aria-labelledby={ `simple-tab-${ index }` }
      { ...other }
    >
      { value === index && (
        <div className="p-2.5">
          { children }
        </div>
      ) }
    </div>
  );
}

export default function SettingsView(): JSX.Element {
  const [value, setValue] = useState(0);

  const handleChange = (event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };

  return (
    <div className="w-full h-full flex flex-col">
      <div>
        <Tabs
          value={ value }
          onChange={ handleChange }
          aria-label="scrollable force tabs example"
          className=""
        >
          <Tab className="h-12 min-h-0" label="Security" iconPosition="start" icon={ <Security /> } />
          <Tab className="h-12 min-h-0" label="About" iconPosition="start" icon={ <Info /> } />
        </Tabs>
      </div>
      <TabPanel value={ value } index={ 0 } className="w-full h-full">
        <div className="w-full 2xl:w-1/2">
          <div className="grid gap-16">
            <ChangePassword />
          </div>
        </div>
      </TabPanel>
      <TabPanel value={ value } index={ 1 }>
        <AboutView />
      </TabPanel>
    </div>
  );
}
