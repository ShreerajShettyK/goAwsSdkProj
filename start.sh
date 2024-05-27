#!/bin/sh
# Source environment variables from the .env file

. /root/.env
set +a

# Run the services
./main &
./db/createMongodb
