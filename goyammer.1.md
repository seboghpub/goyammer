% GOYAMMER(1)
% Sebastian Bogan
% April 2020

<!-- http://jeromebelleman.gitlab.io/posts/publishing/manpages/ -->

# NAME

goyammer - notify about new Yammer messages.

# SYNOPSIS

**goyammer** \<command\> [\<args\>]

# DESCRIPTION

Goyammer is a simple cli tool to poll for new Yammer messages (private ones as well as messages in subscribed groups). New messages will be logged on the console and send to a notification daemon to display desktop notifications. Polling is done using the Yammer API. Authentication is implemented via OAuth 2.0 Implicit Grant.

# COMMANDS

**goyammer-login(1)** Login to Yammer and get an access token.

**goyammer-poll(1)** Poll for new messages and notify.


<!--
# Local Variables:
# mode: markdown
# ispell-local-dictionary: "english"
# eval: (flyspell-mode 1)
# coding: utf-8
# End:
-->
