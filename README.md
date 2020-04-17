# goyammer

Notify about new Yammer messages (private ones as well as messages in subscribed groups).

## Login:

First one needs to get a Yammer access token. Using:

~~~~ {.bash}
goyammer login --client <xyz>
~~~~

where `xyz` must be replaced with the appropriate value, acquires a Yammer access token and stores it locally in `~/.goyammer`.

Note, this only needs to be done once.

## Poll:

Using: 

~~~~ {.bash}
goyammer poll [--interval <seconds>]
~~~~
 
one starts the polling and notification.

## Screenshot

![goyammer](screenshot.png)


