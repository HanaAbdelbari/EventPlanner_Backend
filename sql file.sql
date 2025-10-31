-- Create database
CREATE DATABASE eventplanner;

-- Create user
CREATE USER 'appuser'@'localhost' IDENTIFIED BY 'apppass';

-- Grant permissions
GRANT ALL PRIVILEGES ON eventplanner.* TO 'appuser'@'localhost';
FLUSH PRIVILEGES;

-- Use database
USE eventplanner;

-- Create users table (GORM will auto-migrate later)