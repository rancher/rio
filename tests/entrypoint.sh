#!/bin/sh

curl -sfL https://get.rio.io | sh - > /dev/null 2>&1

# Install rio if it isn't already installed
if [ "$(rio info | grep "rio install")" ] ; then rio install ; fi

exec "$@"
