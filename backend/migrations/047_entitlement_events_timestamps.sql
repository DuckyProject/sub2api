-- Fix entitlement_events schema to match Ent mixins (TimeMixin + SoftDeleteMixin).
-- Migration 046 created entitlement_events without updated_at/deleted_at, which breaks Ent ORM writes.
-- This migration is idempotent and safe to re-run.

ALTER TABLE IF EXISTS entitlement_events
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

