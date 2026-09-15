import { ConfigService } from '@nestjs/config';
import { randomUUID } from 'node:crypto';
import { RedisService } from '../src/integrations/redis/redis.service';

const enabled=Boolean(process.env.REDIS_INTEGRATION_URL);
const describeIntegration=enabled?describe:describe.skip;

describeIntegration('Redis abuse-prevention integration',()=>{
  let redis:RedisService;
  beforeAll(()=>{const values:Record<string,unknown>={NODE_ENV:'test',REDIS_URL:process.env.REDIS_INTEGRATION_URL,REDIS_CONNECT_TIMEOUT_MS:2000};redis=new RedisService({getOrThrow:(key:string)=>values[key]} as ConfigService);});
  afterAll(async()=>redis?.onApplicationShutdown());
  it('increments a pseudonymous key and applies a bounded expiry',async()=>{const key=redis.key('integration',randomUUID());const first=await redis.consumeLimit(key,5000);const second=await redis.consumeLimit(key,5000);expect(first.count).toBe(1);expect(second.count).toBe(2);expect(second.ttlMilliseconds).toBeGreaterThan(0);expect(second.ttlMilliseconds).toBeLessThanOrEqual(5000);});
});
