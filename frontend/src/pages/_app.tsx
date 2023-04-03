import { AppProps } from 'next/app';
import Head from 'next/head';
import { Global, Header, MantineProvider } from '@mantine/core';
import { createStyles } from "@mantine/styles"
import { Notifications } from '@mantine/notifications';

const useStyles = createStyles((theme) => ({
  header: {
    background: "#7289da"
  }
}))

export default function App(props: AppProps) {
  const { Component, pageProps } = props;
  const {classes} = useStyles()

  return (
    <>
      <Head>
        <title>Users</title>
        <meta name="viewport" content="minimum-scale=1, initial-scale=1, width=device-width" />
      </Head>

      <MantineProvider
        withGlobalStyles
        withNormalizeCSS
        theme={{
          /** Put your mantine theme override here */
          colorScheme: 'light',
        }}
      >
        <Global
          styles={(theme) => ({
            body: {
              background: "#282b30"
            }
          })}
        />
        <Notifications/>
        <Header
          height={70}
          withBorder={false}
          className={classes.header}
        >
          <div>

          </div>
        </Header>
        <Component {...pageProps} />
      </MantineProvider>
    </>
  );
}