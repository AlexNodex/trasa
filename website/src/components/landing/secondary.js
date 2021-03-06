import { Typography } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import { makeStyles } from '@material-ui/core/styles';
import React from 'react';
import ThemeBase from '../muiTheme';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';

const useStyles = makeStyles(() => ({
  ctaPad: {
    marginTop: 150,
    textAlign: 'center',
  },
  contained: {
    color: '#000080',
    backgroundColor: 'white',
    borderColorL: '#000080',
    fontWeight: 600,
    //  fontSize: '14px',
    boxShadow: 'none',
  },
}));

export default function Enterprise() {
  const classes = useStyles();
  return (
    <ThemeBase>
      <Grid container spacing={2} direction="column" justify="center" alignItems="center">
        <Grid item xs={8}>
          <div className={classes.ctaPad}>
            <Typography variant="h1">Why you need TRASA ?</Typography>
            <Typography variant="body1" component="span" style={{ textAlign: 'center' }}>
              Data center or dynamic cloud infrastructure, dedicated servers or ephemeral
              applications and <br /> services, access by internal team or managed service provider;
            </Typography>
            <Typography variant="subtitle1" style={{ textAlign: 'center' }}>
              <b>TRASA</b> provides modern security features and enables best practice security{' '}
              <br /> to protect Web, SSH, RDP and Database services from unauthorized or malicious
              access.
            </Typography>
          </div>
        </Grid>
        {/* <Grid item xs={12}>
          <Link className={clsx('button  button--lg', classes.contained)} to={useBaseUrl('docs/')}>
            Learn more about features
          </Link>
        </Grid> */}
      </Grid>
    </ThemeBase>
  );
}
