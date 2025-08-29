# custom-reports-collection

This repo is a collection of custom reports that can be embedded inside Genesys Cloud as a Client App. Previously I was creating these as individual repos but for scale it made more sense to create these in the one repo. There are existing examples for WorkAutomation [here](https://github.com/mcphee11/workautomation-custom-uis) as well as custom dashboard metrics [here](https://github.com/mcphee11/genesys-dashboard-metric) longer term i may roll these into this same repo but it will depend on if i can be bothered or not 😃

Most of these example are and will be based on the `User Snippet` I have created for _Genesys Reports_. These are also designed to be embedded into Genesys Cloud as a [Client App](https://developer.genesys.cloud/platform/integrations/client-apps/). These are also using the newer PKCE Auth Flow. I have put details below in a generic way for most of the examples. If an examples report uses a different OAuth type then I will say so in the folders specific readMe.md file.

Each example will be in its own folder to allow me to create specific notes in a readMe for that example as well as put a description of what it does.

## OAuth

The OAuth for these examples follows the `user snippet` OAuth method which is to use the `Code Authorization / PKCE` type. You will also need to se your redirect for example when workgin locally

```
http://localhost:8080/index.html
```

In a PROD environment that would be your hosting location eg google bucket or amazon S3. As the code is executed on the client side and the Ids are parsed in a URL Parameters it is safe to host this publicly in my view. You will want to check your own organizations view on this.

![](/docs/images/oauth.png?raw=true)

You will now need to set the required `scopes` based on the example you are using. Once created you will need to copy the `ClientID`

```
clientId
```

## Installing the Client App

In the `Integrations` in your Genesys Cloud ORG install the `Client Apps` integration.

![](/docs/images/clientApps.png?raw=true)

Now configure the app with the name you want it to be displayed as in the "Apps" menu in Genesys Cloud.

Under the Properties set the `Application URL` to the hosting location with the additional parameters of your environment.

```
http://localhost:8080/index.html?gc_region=YOUR_REGION&gc_clientId=YOUR_CLIENTID&gc_redirectUrl=http://localhost:8080/index.html
```

Replacing the items for example

- YOUR_REGION = mypurecloud.com.au
- YOUR_CLIENTID = The Id you created in the step above.
- http://localhost:8080/index.html = your hosting location and file name that is accessible.

Set teh `Application Type` as

```
standalone
```

Set the `Iframe Sandbox Options` as

```
allow-scripts,allow-same-origin,allow-forms,allow-modals,allow-downloads
```

#### NOTE: Most of the above will be there by default but without `allow-downloads` the CSV download option will not work.

Set `Group Filtering` to the Group of users you want to be able to see the new report.

Save the settings then ensure that you have set the Integration to `Active`

![](/docs/images/activeToggle.png?raw=true)

Now refresh your browser and if your in that Group you will see the new option under the `Apps` Menu

![](/docs/images/appsMenu.png?raw=true)

## Testing

Personally I like to have an Integration setup called `localhost:8080` and have it pointing to my local environment so then i can do code changes on teh fly while testing before moving the code to a hosted environment to speed up the development process. You can also load the Application URL directly in the browser. For this the notifications will then use the window.alert instead of the GC toaster as your not inside the iFrame of Genesys.
