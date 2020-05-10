% GOYAMMER-POLL(1)
% Sebastian Bogan
% April 2020

<!-- http://jeromebelleman.gitlab.io/posts/publishing/manpages/ -->

# NAME

goyammer-poll - poll Yammer for new messages and notify.

# SYNOPSIS

**goyammer** **poll** [--foreground] [--interval] [--output]

# DESCRIPTION

Login to Yammer and get an access token.

# OPTIONS

**--foreground**
:   Do not detach but run in foreground.

**--interval** \<seconds>
:   The number of seconds to wait between requests.

**--output** \<path>
:   Where to send output to (ignored if **--foregorund** is set). If not specified, output will be discarded.

<!--
# Local Variables:
# mode: markdown
# ispell-local-dictionary: "english"
# eval: (flyspell-mode 1)
# coding: utf-8
# End:
-->
