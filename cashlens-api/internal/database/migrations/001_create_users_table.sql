-- Migration: 001_create_users_table
-- Description: Create users table for Clerk authentication integration
-- Author: Asmit Singh
-- Date: 2024-11-04

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clerk_user_id TEXT UNIQUE NOT NULL,
    email TEXT NOT NULL,
    full_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on clerk_user_id for fast lookups
CREATE INDEX IF NOT EXISTS idx_users_clerk_user_id ON users(clerk_user_id);

-- Create index on email for queries
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add comment to table
COMMENT ON TABLE users IS 'User accounts synchronized from Clerk authentication';
COMMENT ON COLUMN users.clerk_user_id IS 'Clerk user ID (used for JWT validation)';
COMMENT ON COLUMN users.email IS 'User email address';
COMMENT ON COLUMN users.full_name IS 'User full name (first + last)';
