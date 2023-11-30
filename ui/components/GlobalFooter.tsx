import { Typography } from '@mui/material';
import React, { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { getRelease } from 'shared/bootstrap/selectors';

const GlobalFooter = (): JSX.Element => {
  const release = useSelector(getRelease);

  let versionLink: JSX.Element = null;
  if (release && !release.startsWith('LOCAL')) {
    versionLink = (
      <Fragment>
        <span>- </span>
        <a
          target="_blank"
          rel="noreferrer"
          href={ `https://github.com/monetr/monetr/releases/tag/${ release }` }
        >
          { release }
        </a>
      </Fragment>
    )
  } else {
    versionLink = (
      <Fragment>
        <span>- LOCAL DEVELOPMENT</span>
      </Fragment>
    )
  }

  return (
    <Typography
      className="absolute inline w-full text-center bottom-1 opacity-30"
    >
      © { new Date().getFullYear() } monetr LLC { versionLink }
    </Typography>
  );
};

export default GlobalFooter;