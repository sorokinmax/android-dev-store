# websspi

[![GoDoc](https://godoc.org/github.com/quasoft/websspi?status.svg)](https://godoc.org/github.com/quasoft/websspi) [![Build Status](https://travis-ci.org/quasoft/websspi.png?branch=master)](https://travis-ci.org/quasoft/websspi) [![Coverage Status](https://coveralls.io/repos/github/quasoft/websspi/badge.svg?branch=master)](https://coveralls.io/github/quasoft/websspi?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/quasoft/websspi)](https://goreportcard.com/report/github.com/quasoft/websspi)

`websspi` is an HTTP middleware for Golang that uses Kerberos/NTLM for single sign-on (SSO) authentication of browser based clients in a Windows environment.

It performs authentication of HTTP requests without the need to create or use keytab files.

The middleware implements the scheme defined by RFC4559 (SPNEGO-based HTTP Authentication in Microsoft Windows) to exchange security tokens via HTTP headers and uses SSPI (Security Support Provider Interface) to authenticate HTTP requests.

## How to use

The [examples directory](https://github.com/quasoft/websspi/tree/master/examples) contains a [simple web server](https://github.com/quasoft/websspi/blob/master/examples/server_windows.go) that demonstrates how to use the package.
Before trying it, you need to prepare your environment:

1. Create a separate user account in active directory, under which the web server process will be running (eg. `user` under the `domain.local` domain)

2. Create a service principal name for the host with class HTTP:
   - Start Command prompt or PowerShell as domain administrator
   - Run the command below, replacing `host.domain.local` with the fully qualified domain name of the server where the web application will be running, and `domain\user` with the name of the account created in step 1.:

         setspn -A HTTP/host.domain.local domain\user

3. Start the web server app under the account created in step 1.

4. If you are using Chrome, Edge or Internet Explorer, add the URL of the web app to the Local intranet sites (`Internet Options -> Security -> Local intranet -> Sites`)

5. Start Chrome, Edge or Internet Explorer and navigate to the URL of the web app (eg. `http://host.domain.local:9000`)

6. The web app should greet you with the name of your AD account without asking you to login. In case it doesn't, make sure that:

   - You are not running the web browser on the same server where the web app is running. You should be running the web browser on a domain joined computer (client) that is different from the server. If you do run the web browser at the same server SSPI package will fallback to NTLM protocol and Kerberos will not be used.
   - There is only one HTTP/... SPN for the host
   - The SPN contains only the hostname, without the port
   - You have added the URL of the web app to the `Local intranet` zone
   - The clocks of the server and client should not differ with more than 5 minutes
   - `Integrated Windows Authentication` should be enabled in Internet Explorer (under `Advanced settings`)

## Security requirements

- SPNEGO over HTTP provides no facilities for protection of the authroization data contained in HTTP headers (the `Authorization` and `WWW-Authenticate` headers), which means that the web server **MUST** enforce use of HTTPS to provide confidentiality for the data in those headers!
