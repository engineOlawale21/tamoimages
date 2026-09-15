import { Inject, Injectable } from '@nestjs/common';
import { createHash } from 'node:crypto';
import { Pool } from 'pg';
import { DATABASE_POOL } from '../database/database.constants';

export type SecurityEvent = {
  type: 'registration_requested' | 'email_verified' | 'login' | 'session_refreshed' | 'session_reuse' | 'logout' | 'logout_all' | 'password_recovery_requested' | 'password_reset' | 'account_deletion_requested';
  outcome: 'success' | 'failure';
  correlationId?: string;
  userId?: string;
  subject?: string;
};

@Injectable()
export class SecurityAuditService {
  constructor(@Inject(DATABASE_POOL) private readonly pool: Pool) {}

  async record(event: SecurityEvent): Promise<void> {
    const subjectHash = event.subject ? createHash('sha256').update(event.subject.trim().toLowerCase()).digest() : null;
    await this.pool.query(
      `INSERT INTO security_audit_events(user_id,event_type,outcome,correlation_id,subject_hash)
       VALUES($1,$2,$3,$4,$5)`,
      [event.userId ?? null, event.type, event.outcome, event.correlationId ?? null, subjectHash],
    );
  }
}
