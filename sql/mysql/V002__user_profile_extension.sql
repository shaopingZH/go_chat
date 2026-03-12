-- =============================================
-- Feature: user profile extension
-- Version: V002
-- Date: 2026-02-12
-- Depends: V001
-- =============================================

ALTER TABLE users
  ADD COLUMN display_name VARCHAR(50) NOT NULL DEFAULT '' AFTER username,
  ADD COLUMN bio VARCHAR(200) NOT NULL DEFAULT '' AFTER avatar;
