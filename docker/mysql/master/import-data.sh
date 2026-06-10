#!/bin/bash
set -e

echo "Starting production data import..."

# Import production data with foreign key checks disabled
mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" billing_revaktor --init-command='SET FOREIGN_KEY_CHECKS=0;' < /data/production-data.sql

echo "Production data import completed successfully!"
