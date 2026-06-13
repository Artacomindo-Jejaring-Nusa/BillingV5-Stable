#!/bin/bash
set -e

if [ "${SKIP_DATA_IMPORT}" = "true" ] || [ "${SKIP_DATA_IMPORT}" = "1" ]; then
    echo "SKIP_DATA_IMPORT is set — skipping production data import."
else
    echo "Starting production data import..."
    # Import production data with foreign key checks disabled
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" billing_revaktor --init-command='SET FOREIGN_KEY_CHECKS=0;' < /data/production-data.sql
    echo "Production data import completed successfully!"
fi
