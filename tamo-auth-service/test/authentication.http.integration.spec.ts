import * as bcrypt from 'bcrypt';
import type { INestApplication } from '@nestjs/common';
import { randomUUID } from 'node:crypto';
import { Pool } from 'pg';

const enabled=Boolean(process.env.DATABASE_INTEGRATION_URL&&process.env.REDIS_INTEGRATION_URL);
const describeIntegration=enabled?describe:describe.skip;

describeIntegration('authentication HTTP integration',()=>{
  let pool:Pool; let app:INestApplication; let base:string; let userId:string;
  beforeAll(async()=>{const {createApplication}=await import('../src/bootstrap');pool=new Pool({connectionString:process.env.DATABASE_INTEGRATION_URL});const email=`http-${randomUUID()}@example.test`;const result=await pool.query<{id:string}>(`INSERT INTO users(email,first_name,last_name,role,password_hash,status,email_verified_at) VALUES($1,'HTTP','Test','buyer',$2,'active',now()) RETURNING id`,[email,await bcrypt.hash('correct horse battery staple',12)]);userId=result.rows[0].id;process.env.HTTP_TEST_EMAIL=email;app=await createApplication();await app.listen(0,'127.0.0.1');const address=app.getHttpServer().address() as {port:number};base=`http://127.0.0.1:${address.port}/api/v1`;},30_000);
  afterAll(async()=>{if(app)await app.close();if(pool){await pool.query('DELETE FROM users WHERE id=$1',[userId]);await pool.end();}});
  it('rejects invalid bodies as problem details',async()=>{const response=await fetch(`${base}/auth/login`,{method:'POST',headers:{'content-type':'application/json'},body:'{}'});expect(response.status).toBe(400);expect(response.headers.get('content-type')).toContain('application/problem+json');});
  it('logs in, rotates the cookie, and rejects refresh replay',async()=>{const login=await fetch(`${base}/auth/login`,{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({email:process.env.HTTP_TEST_EMAIL,password:'correct horse battery staple'})});expect(login.status).toBe(200);const cookie=login.headers.get('set-cookie');expect(cookie).toContain('__Host-tamo-refresh=');expect(cookie).toContain('HttpOnly');const refresh=await fetch(`${base}/auth/refresh`,{method:'POST',headers:{cookie:cookie!.split(';')[0],origin:'http://localhost:3000'}});expect(refresh.status).toBe(200);const replay=await fetch(`${base}/auth/refresh`,{method:'POST',headers:{cookie:cookie!.split(';')[0],origin:'http://localhost:3000'}});expect(replay.status).toBe(401);});
  it('enforces exact origin on cookie mutations',async()=>{const response=await fetch(`${base}/auth/logout`,{method:'POST',headers:{origin:'https://attacker.example'}});expect(response.status).toBe(403);});
});
