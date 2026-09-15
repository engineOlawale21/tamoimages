import { Inject, Injectable } from '@nestjs/common';
import { Pool, PoolClient } from 'pg';
import { DATABASE_POOL } from '../database/database.constants';
import { AccountRole } from './dto';

export type AccountStatus = 'pending_verification' | 'active' | 'suspended' | 'deletion_requested';
export type UserRecord = { id: string; email: string; firstName: string; lastName: string; role: AccountRole; passwordHash: string; status: AccountStatus; emailVerifiedAt?: Date; registrationNoticeVersion?:string };
export type SessionRecord = { id: string; userId: string; familyId: string; expiresAt: Date };
export type RotationResult = { state: 'rotated'; session: SessionRecord } | { state: 'invalid' } | { state: 'reused' };

export abstract class UsersRepository {
  abstract create(user: Omit<UserRecord, 'id' | 'status'>): Promise<UserRecord>;
  abstract findActiveByEmail(email: string): Promise<UserRecord | undefined>;
  abstract findById(id: string): Promise<UserRecord | undefined>;
  abstract createVerificationToken(userId: string, tokenHash: Buffer, expiresAt: Date): Promise<void>;
  abstract consumeVerificationToken(tokenHash: Buffer): Promise<boolean>;
  abstract createSession(input: { userId: string; familyId: string; tokenHash: Buffer; expiresAt: Date; userAgentHash?: Buffer; ipPrefixHash?: Buffer }): Promise<SessionRecord>;
  abstract rotateSession(input: { tokenHash: Buffer; nextTokenHash: Buffer; nextExpiresAt: Date; idleAfter: Date }): Promise<RotationResult>;
  abstract revokeSession(tokenHash: Buffer, reason: string): Promise<void>;
  abstract revokeAllSessions(userId: string, reason: string): Promise<void>;
  abstract createRecoveryToken(userId: string, tokenHash: Buffer, expiresAt: Date): Promise<void>;
  abstract consumeRecoveryToken(tokenHash: Buffer, passwordHash: string): Promise<string | undefined>;
  abstract requestDeletion(userId: string): Promise<void>;
}

const userColumns = `id,email,first_name AS "firstName",last_name AS "lastName",role,password_hash AS "passwordHash",status,email_verified_at AS "emailVerifiedAt"`;

@Injectable()
export class PostgresUsersRepository implements UsersRepository {
  constructor(@Inject(DATABASE_POOL) private readonly pool: Pool) {}
  async create(user: Omit<UserRecord, 'id' | 'status'>): Promise<UserRecord> { const result=await this.pool.query<UserRecord>(`INSERT INTO users(email,first_name,last_name,role,password_hash,registration_notice_version,registration_notice_acknowledged_at) SELECT $1,$2,$3,$4,$5,$6,now() WHERE EXISTS (SELECT 1 FROM privacy_notices WHERE version=$6 AND effective_at<=now() AND retired_at IS NULL) RETURNING ${userColumns}`,[user.email,user.firstName,user.lastName,user.role,user.passwordHash,user.registrationNoticeVersion]);if(!result.rowCount)throw new Error('privacy_notice_version_invalid');return result.rows[0]; }
  async findActiveByEmail(email:string):Promise<UserRecord|undefined>{const result=await this.pool.query<UserRecord>(`SELECT ${userColumns} FROM users WHERE email=$1 AND deleted_at IS NULL LIMIT 1`,[email]);return result.rows[0];}
  async findById(id:string):Promise<UserRecord|undefined>{const result=await this.pool.query<UserRecord>(`SELECT ${userColumns} FROM users WHERE id=$1 AND deleted_at IS NULL LIMIT 1`,[id]);return result.rows[0];}
  async createVerificationToken(userId:string,tokenHash:Buffer,expiresAt:Date):Promise<void>{await this.transaction(async(client)=>{await client.query(`UPDATE email_verification_tokens SET consumed_at=now() WHERE user_id=$1 AND consumed_at IS NULL`,[userId]);await client.query(`INSERT INTO email_verification_tokens(user_id,token_hash,expires_at) VALUES($1,$2,$3)`,[userId,tokenHash,expiresAt]);});}
  async consumeVerificationToken(tokenHash:Buffer):Promise<boolean>{return this.transaction(async(client)=>{const token=await client.query<{userId:string}>(`UPDATE email_verification_tokens SET consumed_at=now() WHERE token_hash=$1 AND consumed_at IS NULL AND expires_at>now() RETURNING user_id AS "userId"`,[tokenHash]);if(!token.rowCount)return false;await client.query(`UPDATE users SET status='active',email_verified_at=COALESCE(email_verified_at,now()),updated_at=now() WHERE id=$1 AND status='pending_verification'`,[token.rows[0].userId]);return true;});}
  async createSession(input:{userId:string;familyId:string;tokenHash:Buffer;expiresAt:Date;userAgentHash?:Buffer;ipPrefixHash?:Buffer}):Promise<SessionRecord>{const result=await this.pool.query<SessionRecord>(`INSERT INTO refresh_sessions(user_id,family_id,token_hash,expires_at,user_agent_hash,ip_prefix_hash) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,user_id AS "userId",family_id AS "familyId",expires_at AS "expiresAt"`,[input.userId,input.familyId,input.tokenHash,input.expiresAt,input.userAgentHash,input.ipPrefixHash]);return result.rows[0];}
  async rotateSession(input:{tokenHash:Buffer;nextTokenHash:Buffer;nextExpiresAt:Date;idleAfter:Date}):Promise<RotationResult>{return this.transaction(async(client)=>{const current=await client.query<{id:string;userId:string;familyId:string;expiresAt:Date;lastUsedAt:Date;rotatedAt?:Date;revokedAt?:Date}>(`SELECT id,user_id AS "userId",family_id AS "familyId",expires_at AS "expiresAt",last_used_at AS "lastUsedAt",rotated_at AS "rotatedAt",revoked_at AS "revokedAt" FROM refresh_sessions WHERE token_hash=$1 FOR UPDATE`,[input.tokenHash]);if(!current.rowCount)return {state:'invalid'};const session=current.rows[0];if(session.rotatedAt||session.revokedAt){await client.query(`UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,now()),revocation_reason='refresh_reuse' WHERE family_id=$1`,[session.familyId]);return {state:'reused'};}if(new Date(session.expiresAt)<=new Date()||new Date(session.lastUsedAt)<input.idleAfter)return {state:'invalid'};const next=await client.query<SessionRecord>(`INSERT INTO refresh_sessions(user_id,family_id,token_hash,expires_at) VALUES($1,$2,$3,$4) RETURNING id,user_id AS "userId",family_id AS "familyId",expires_at AS "expiresAt"`,[session.userId,session.familyId,input.nextTokenHash,input.nextExpiresAt]);await client.query(`UPDATE refresh_sessions SET rotated_at=now(),replaced_by_session_id=$2,last_used_at=now() WHERE id=$1`,[session.id,next.rows[0].id]);return {state:'rotated',session:next.rows[0]};});}
  async revokeSession(tokenHash:Buffer,reason:string):Promise<void>{await this.pool.query(`UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,now()),revocation_reason=$2 WHERE token_hash=$1`,[tokenHash,reason]);}
  async revokeAllSessions(userId:string,reason:string):Promise<void>{await this.pool.query(`UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,now()),revocation_reason=$2 WHERE user_id=$1 AND revoked_at IS NULL`,[userId,reason]);}
  async createRecoveryToken(userId:string,tokenHash:Buffer,expiresAt:Date):Promise<void>{await this.transaction(async(client)=>{await client.query(`UPDATE password_recovery_tokens SET consumed_at=now() WHERE user_id=$1 AND consumed_at IS NULL`,[userId]);await client.query(`INSERT INTO password_recovery_tokens(user_id,token_hash,expires_at) VALUES($1,$2,$3)`,[userId,tokenHash,expiresAt]);});}
  async consumeRecoveryToken(tokenHash:Buffer,passwordHash:string):Promise<string|undefined>{return this.transaction(async(client)=>{const token=await client.query<{userId:string}>(`UPDATE password_recovery_tokens SET consumed_at=now() WHERE token_hash=$1 AND consumed_at IS NULL AND expires_at>now() RETURNING user_id AS "userId"`,[tokenHash]);if(!token.rowCount)return undefined;const userId=token.rows[0].userId;await client.query(`UPDATE users SET password_hash=$2,password_changed_at=now(),updated_at=now() WHERE id=$1`,[userId,passwordHash]);await client.query(`UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,now()),revocation_reason='password_reset' WHERE user_id=$1 AND revoked_at IS NULL`,[userId]);return userId;});}
  async requestDeletion(userId:string):Promise<void>{await this.transaction(async(client)=>{await client.query(`UPDATE users SET status='deletion_requested',updated_at=now() WHERE id=$1 AND deleted_at IS NULL`,[userId]);await client.query(`UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,now()),revocation_reason='account_deletion_requested' WHERE user_id=$1 AND revoked_at IS NULL`,[userId]);});}
  private async transaction<T>(operation:(client:PoolClient)=>Promise<T>):Promise<T>{const client=await this.pool.connect();try{await client.query('BEGIN');const result=await operation(client);await client.query('COMMIT');return result;}catch(error){await client.query('ROLLBACK');throw error;}finally{client.release();}}
}
