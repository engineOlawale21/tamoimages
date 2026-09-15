import { ConfigService } from '@nestjs/config';
import { Request, Response } from 'express';
import { RateLimitMiddleware } from '../src/common/middleware/rate-limit.middleware';
import { HealthController } from '../src/health.controller';
import { HealthService } from '../src/health.service';
import { RedisService } from '../src/integrations/redis/redis.service';

function config(): ConfigService {
  const values: Record<string, unknown> = {
    AUTH_RATE_LIMIT_MAX: 2,
    AUTH_RATE_LIMIT_WINDOW_MS: 60000,
    RATE_LIMIT_HMAC_SECRET: 'a-test-rate-limit-secret-of-32-characters',
  };
  return { getOrThrow: (key: string) => values[key] } as ConfigService;
}

function response() {
  const result = {
    setHeader: jest.fn(), status: jest.fn(), type: jest.fn(), send: jest.fn(),
  };
  result.status.mockReturnValue(result);
  result.type.mockReturnValue(result);
  return result as unknown as Response;
}

describe('HTTP foundation', () => {
  it('serves liveness without calling dependencies', () => {
    const health = { readiness: jest.fn() } as unknown as HealthService;
    expect(new HealthController(health).liveness()).toEqual({ status: 'alive', service: 'tamo-auth-service' });
    expect(health.readiness).not.toHaveBeenCalled();
  });

  it('permits authentication requests below the Redis-backed limit', async () => {
    const redis = { key: jest.fn().mockReturnValue('test:key'), consumeLimit: jest.fn().mockResolvedValue({ count: 1, ttlMilliseconds: 50000 }) } as unknown as RedisService;
    const middleware = new RateLimitMiddleware(config(), redis);
    const next = jest.fn();
    await middleware.use({ socket: { remoteAddress: '127.0.0.1' } } as Request, response(), next);
    expect(next).toHaveBeenCalledTimes(1);
    expect(redis.key).toHaveBeenCalledWith('rate-limit', expect.stringMatching(/^auth:ip:[a-f0-9]{64}$/));
  });

  it('returns problem details when the limit is exceeded', async () => {
    const redis = { key: jest.fn().mockReturnValue('test:key'), consumeLimit: jest.fn().mockResolvedValue({ count: 3, ttlMilliseconds: 50000 }) } as unknown as RedisService;
    const result = response();
    await new RateLimitMiddleware(config(), redis).use({ socket: { remoteAddress: '127.0.0.1' }, originalUrl: '/api/v1/auth/login', correlationId: 'request-1' } as Request & { correlationId: string }, result, jest.fn());
    expect(result.status).toHaveBeenCalledWith(429);
    expect(result.send).toHaveBeenCalledWith(expect.objectContaining({ status: 429, correlationId: 'request-1' }));
  });
});
