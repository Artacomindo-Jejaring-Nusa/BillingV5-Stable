#!/bin/bash
set -e

# Skip data import jika SKIP_DATA_IMPORT=true
if [ "${SKIP_DATA_IMPORT}" = "true" ] || [ "${SKIP_DATA_IMPORT}" = "1" ]; then
    echo "SKIP_DATA_IMPORT is set — skipping production data import."
    exec mysqld "$@"
fi

# Run the original MySQL entrypoint
/docker-entrypoint.sh mysqld "$@" &

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
until mysqladmin ping -h localhost --silent; do
    sleep 1
done
echo "MySQL is ready!"

# Import production data with foreign key checks disabled
if [ -f /data/01-production-data.sql ]; then
    echo "Importing production data with foreign key checks disabled..."
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" billing_revaktor -e "SET FOREIGN_KEY_CHECKS=0;"
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" billing_revaktor < /data/01-production-data.sql
    echo "Imported DataTerbaru2026.sql"
fi

if [ -f /data/02-production-data-cleaned.sql ]; then
    mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" billing_revaktor < /data/02-production-data-cleaned.sql
    echo "Imported DataTerbaru2026_Cleaned.sql"
fi

# Re-enable foreign key checks
mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" billing_revaktor -e "SET FOREIGN_KEY_CHECKS=1;"
echo "Foreign key checks re-enabled"

# Wait for the background MySQL process
wait
