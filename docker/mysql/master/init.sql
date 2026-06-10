CREATE USER 'repl_user'@'%' IDENTIFIED WITH mysql_native_password BY 'repl_password';
GRANT REPLICATION SLAVE ON *.* TO 'repl_user'@'%';
FLUSH PRIVILEGES;

CREATE DATABASE IF NOT EXISTS billing_revaktor;

USE billing_revaktor;
SET FOREIGN_KEY_CHECKS=0;
