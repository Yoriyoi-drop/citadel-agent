-- Initial database setup for Citadel Agent
-- This file is used by docker-compose to initialize the database

-- Create extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the citadel database and user (in case they don't exist)
-- NOTE: These are typically handled by docker-compose environment variables,
-- but the init.sql file is executed inside the postgres container as the postgres user

-- Insert initial data if needed
-- (We'll keep this minimal for now, more seed data can be added as needed)