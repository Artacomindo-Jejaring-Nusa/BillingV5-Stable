#!/bin/bash
set -e

echo "Waiting for MySQL to be ready..."
until docker exec mysql-master mysqladmin ping -h localhost --silent; do
    sleep 1
done
echo "MySQL is ready!"

echo "Importing production data with foreign key checks disabled..."
docker exec -i mysql-master mysql -uroot -proot billing_revaktor <<EOF
SET FOREIGN_KEY_CHECKS=0;
SOURCE /data/01-production-data.sql;
SOURCE /data/02-production-data-cleaned.sql;
SET FOREIGN_KEY_CHECKS=1;
EOF

echo "Data import completed successfully!"
