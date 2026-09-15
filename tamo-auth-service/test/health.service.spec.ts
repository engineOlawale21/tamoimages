import { Pool } from 'pg';
import { HealthService } from '../src/health.service';
import { KafkaService } from '../src/integrations/kafka/kafka.service';
import { RedisService } from '../src/integrations/redis/redis.service';

describe('HealthService', () => {
  it('reports ready only when every dependency responds', async () => {
    const database = { query: jest.fn().mockResolvedValue({ rows: [{ '?column?': 1 }] }) } as unknown as Pool;
    const redis = { ping: jest.fn().mockResolvedValue(undefined) } as unknown as RedisService;
    const kafka = { ready: jest.fn().mockResolvedValue(undefined) } as unknown as KafkaService;
    await expect(new HealthService(database, redis, kafka).readiness()).resolves.toMatchObject({ status: 'ready' });
  });

  it('reports only dependency names and states when a dependency fails', async () => {
    const database = { query: jest.fn().mockRejectedValue(new Error('contains private connection details')) } as unknown as Pool;
    const redis = { ping: jest.fn().mockResolvedValue(undefined) } as unknown as RedisService;
    const kafka = { ready: jest.fn().mockResolvedValue(undefined) } as unknown as KafkaService;
    const result = await new HealthService(database, redis, kafka).readiness();
    expect(result).toEqual({ status: 'not-ready', checks: [
      { name: 'postgres', status: 'down' }, { name: 'redis', status: 'up' }, { name: 'kafka', status: 'up' },
    ] });
    expect(JSON.stringify(result)).not.toContain('private connection details');
  });
});
