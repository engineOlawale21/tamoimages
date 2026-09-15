import { validateEnvironment } from '../src/config/environment';

const validEnvironment = {
  NODE_ENV: 'test', PORT: '4000', WEB_ORIGIN: 'http://localhost:3000',
  DATABASE_URL: 'postgresql://tamo:password@localhost:5432/tamo_identity',
  REDIS_URL: 'redis://localhost:6379', KAFKA_BROKERS: 'localhost:9092',
  JWT_SECRET: 'test', JWT_TTL: '15m',
  RATE_LIMIT_HMAC_SECRET: 'a-test-rate-limit-secret-of-32-characters',
};

describe('environment validation', () => {
  it('normalizes validated values', () => {
    expect(validateEnvironment(validEnvironment)).toMatchObject({ PORT: 4000, API_PREFIX: 'api/v1' });
  });

  it('rejects weak production JWT secrets', () => {
    expect(() => validateEnvironment({ ...validEnvironment, NODE_ENV: 'production', JWT_SECRET: 'weak' }))
      .toThrow('JWT_SECRET must contain at least 32 characters');
  });

  it('rejects malformed dependency URLs', () => {
    expect(() => validateEnvironment({ ...validEnvironment, DATABASE_URL: 'not-a-url' }))
      .toThrow('DATABASE_URL must be a valid URL');
  });
});
