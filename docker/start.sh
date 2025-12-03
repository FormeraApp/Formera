#!/bin/sh

# Start nginx
nginx

# Start backend
./server &

# Start nuxt
node .output/server/index.mjs &

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $?
