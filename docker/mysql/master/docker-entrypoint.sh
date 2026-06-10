#!/bin/bash
set -e

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
