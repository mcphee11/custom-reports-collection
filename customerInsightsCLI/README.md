# customer-insights-cli

## THIS IS CURRENTLY IN BETA TESTING

This is designed as a cli version of the web based customer insights report that can be used by Partners and Customers that require to run reports over larger date ranges and gather more data then the client side browser limits you to of 2Gb of memory.

As this runs as a server side cli app this uses `client credentials` to OAuth as well as it generates a HTML file output with the data inside it so that it can be opened and viewed at any time without the need to call the API to view the data again.

## Running the CLI
While if you have GO installed you can run this from the source, I have also compiled version for 

- Linux
- Windows x86
- Windows ARM64
- MacOS

Each file is in a folder in this repo. In case you want to know how to build these from a local source you have after editing the source maybe in your own fork have a look at the `build.sh` file I have that I use to build these files.

## CLI Options

- clientId: This is the `clientId` of the client credentials of the OAuth token created in Genesys Cloud
- secret: This is the `secret` of the client credentials of the OAuth token created in Genesys Cloud
- startDate: ISO formatted dateTime eg 2026-01-01T00:00:00
- endDate: ISO formatted dateTime eg 2026-01-31T23:59:00
- region: of the Genesys Cloud ORG eg mypurecloud.com.au
- timeZone: in IANA format eg Australia/Melbourne

If you want to check the version of the CLI you can pass in the `-v` flag

For example if your running Linux you would use the linux built file in the linux folder and enter the below to run the CLI from where ever the customer-insights file is.

```
./customer-insights -clientId YOUR_CLIENT_ID -secret YOUR_SECRET -startDate 2026-01-01T00:00:00 -endDate 2026-01-31T23:59:00 -region mypurecloud.com.au -timeZone Australia/Melbourne
```

### NOTE: you may need to allow this to be an executable eg `chmod 777` in linux
