#!/bin/bash

plugin_pid=$(ps -o pid= -C "$1")
while [ -z ${plugin_pid}];
do plugin_pid=$(ps -o pid= -C "$1"); done

dlv --listen=:2345 --headless=true --api-version=2 attach  ${plugin_pid}
