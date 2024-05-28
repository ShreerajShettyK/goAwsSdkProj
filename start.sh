#!/bin/sh

# Debugging step to list the contents of the /root directory
ls -la /root/

# Source environment variables from the .env file
set -a
. /root/.env
set +a

# Print environment variables for debugging
env

# Run the services
./main &
./db/createMongodb
