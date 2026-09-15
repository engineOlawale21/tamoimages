import { randomBytes, randomUUID } from 'node:crypto';
import { Pool } from 'pg';
import { AccountRole } from '../src/auth/dto';
import { PostgresUsersRepository } from '../src/auth/users.repository';

const connectionString=process.env.DATABASE_INTEGRATION_URL;
const describeIntegration=connectionString?describe:describe.skip;

describeIntegration('authentication lifecycle integration',()=>{
  let pool:Pool; let repository:PostgresUsersRepository; let userId:string;
  beforeAll(async()=>{pool=new Pool({connectionString,max:4});repository=new PostgresUsersRepository(pool);const result=await pool.query<{id:string}>(`INSERT INTO users(email,first_name,last_name,role,password_hash,status,email_verified_at) VALUES($1,'Integration','Test','buyer','not-used','active',now()) RETURNING id`,[`integration-${randomUUID()}@example.test`]);userId=result.rows[0].id;});
  afterAll(async()=>{if(pool){await pool.query('DELETE FROM users WHERE id=$1',[userId]);await pool.end();}});
  it('consumes an email-verification token only once',async()=>{const pending=await repository.create({email:`pending-${randomUUID()}@example.test`,firstName:'Pending',lastName:'Test',role:AccountRole.BUYER,passwordHash:'not-used'});const hash=randomBytes(32);await repository.createVerificationToken(pending.id,hash,new Date(Date.now()+60_000));await expect(repository.consumeVerificationToken(hash)).resolves.toBe(true);await expect(repository.consumeVerificationToken(hash)).resolves.toBe(false);await pool.query('DELETE FROM users WHERE id=$1',[pending.id]);});
  it('allows one concurrent rotation and treats the other as replay',async()=>{const tokenHash=randomBytes(32);await repository.createSession({userId,familyId:randomUUID(),tokenHash,expiresAt:new Date(Date.now()+60_000)});const input=()=>({tokenHash,nextTokenHash:randomBytes(32),nextExpiresAt:new Date(Date.now()+60_000),idleAfter:new Date(Date.now()-60_000)});const results=await Promise.all([repository.rotateSession(input()),repository.rotateSession(input())]);expect(results.map(result=>result.state).sort()).toEqual(['reused','rotated']);const active=await pool.query(`SELECT 1 FROM refresh_sessions WHERE user_id=$1 AND revoked_at IS NULL`,[userId]);expect(active.rowCount).toBe(0);});
  it('rejects refresh after the idle deadline',async()=>{const tokenHash=randomBytes(32);await repository.createSession({userId,familyId:randomUUID(),tokenHash,expiresAt:new Date(Date.now()+60_000)});await expect(repository.rotateSession({tokenHash,nextTokenHash:randomBytes(32),nextExpiresAt:new Date(Date.now()+60_000),idleAfter:new Date(Date.now()+1_000)})).resolves.toEqual({state:'invalid'});});
});
