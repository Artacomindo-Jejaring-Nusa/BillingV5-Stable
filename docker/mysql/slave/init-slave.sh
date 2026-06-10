#!/bin/bash
set -e

# Wait for master to be ready
echo "Waiting for master database to be ready..."
until mysql -h mysql-master -u root -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1" &> /dev/null; do
    echo "Master is unavailable - sleeping"
    sleep 3
done
echo "Master is up!"

# Get Master status
echo "Getting Master Status..."
MASTER_STATUS=$(mysql -h mysql-master -u root -p"$MYSQL_ROOT_PASSWORD" -e "SHOW MASTER STATUS\G")

# Extract File and Position (using basic bash text processing)
MASTER_LOG_FILE=$(echo "$MASTER_STATUS" | grep File | awk '{print $2}')
MASTER_LOG_POS=$(echo "$MASTER_STATUS" | grep Position | awk '{print $2}')

if [ -z "$MASTER_LOG_FILE" ] || [ -z "$MASTER_LOG_POS" ]; then
    echo "Failed to get master status. Replication might not start correctly."
    exit 1
fi

echo "Master Log File: $MASTER_LOG_FILE"
echo "Master Log Pos: $MASTER_LOG_POS"

# Configure Slave
echo "Configuring Slave..."
mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "
    STOP SLAVE;
    CHANGE MASTER TO
    MASTER_HOST='mysql-master',
    MASTER_USER='repl_user',
    MASTER_PASSWORD='repl_password',
    MASTER_LOG_FILE='$MASTER_LOG_FILE',
    MASTER_LOG_POS=$MASTER_LOG_POS,
    GET_MASTER_PUBLIC_KEY=1;
    START SLAVE;
    SHOW SLAVE STATUS\G
"
echo "Slave configured successfully!"