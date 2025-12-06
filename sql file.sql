-- Create database
CREATE DATABASE eventplanner;

-- Update status values to match new requirements
UPDATE event_attendees
SET status = CASE
    WHEN status = 'accepted' THEN 'going'
    WHEN status = 'declined' THEN 'not_going'
    WHEN status = 'pending' THEN 'maybe'
    ELSE status
END;

-- Create user
CREATE USER 'appuser'@'localhost' IDENTIFIED BY 'apppass';

-- Grant permissions
GRANT ALL PRIVILEGES ON eventplanner.* TO 'appuser'@'localhost';
FLUSH PRIVILEGES;

-- Use database
USE eventplanner;

-- Create users table (GORM will auto-migrate later)