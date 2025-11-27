-- Create database
CREATE DATABASE eventplanner_db;

-- Create user
CREATE USER 'appuser'@'localhost' IDENTIFIED BY 'apppass';

-- Grant permissions
GRANT ALL PRIVILEGES ON eventplanner.* TO 'appuser'@'localhost';
FLUSH PRIVILEGES;

-- Use database
USE eventplanner_db;

-- Create users table (GORM will auto-migrate later)