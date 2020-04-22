# goyammer

[![Build Status](https://travis-ci.com/seboghpub/goyammer.svg?branch=master)](https://travis-ci.com/seboghpub/goyammer)

Notify about new Yammer messages (private ones as well as messages in subscribed
groups).

## Register App

Follow [this] guide to register a Yammer app and to optain a client ID.

## Login:

After an app has been registered, one needs to get a Yammer access token. Using:

~~~ {.bash}
goyammer login --client <xyz>
~~~

where `xyz` must be replaced with the client ID.

If successful, the accquired Yammer access token will be stored in
`~/.goyammer`.

Note, this only needs to be done once.

## Poll:

Using:

~~~ {.bash}
goyammer poll [--interval <seconds>]
~~~

one starts the polling and notification.

## Screenshot

![goyammer]

  [this]: https://developer.yammer.com/docs/app-registration
  [goyammer]: screenshot.png
