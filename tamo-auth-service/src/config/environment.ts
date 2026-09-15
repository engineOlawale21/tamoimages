const environments = new Set(['development', 'test', 'staging', 'production']);

function requiredUrl(environment: Record<string, unknown>, key: string): string {
  const value = String(environment[key] ?? '');
  try {
    return new URL(value).toString();
  } catch {
    throw new Error(`${key} must be a valid URL`);
  }
}

export function validateEnvironment(environment: Record<string, unknown>) {
  const nodeEnvironment = String(environment.NODE_ENV ?? 'development');
  if (!environments.has(nodeEnvironment)) throw new Error('NODE_ENV is invalid');
  const port = Number(environment.PORT ?? 4000);
  if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error('PORT must be between 1 and 65535');
  const jwtSecret = String(environment.JWT_SECRET ?? '');
  if (nodeEnvironment !== 'test' && jwtSecret.length < 32) throw new Error('JWT_SECRET must contain at least 32 characters');
  const rateLimitSecret = String(environment.RATE_LIMIT_HMAC_SECRET ?? (nodeEnvironment === 'test' ? jwtSecret : ''));
  if (rateLimitSecret.length < 32) throw new Error('RATE_LIMIT_HMAC_SECRET must contain at least 32 characters');
  const jwtTtl = String(environment.JWT_TTL ?? '15m');
  if (!/^\d+[smhd]$/.test(jwtTtl)) throw new Error('JWT_TTL must use s, m, h, or d units');
  const emailVerificationTtl = duration(environment.EMAIL_VERIFICATION_TTL, '30m', 'EMAIL_VERIFICATION_TTL');
  const passwordRecoveryTtl = duration(environment.PASSWORD_RECOVERY_TTL, '15m', 'PASSWORD_RECOVERY_TTL');
  const refreshTokenTtl = duration(environment.REFRESH_TOKEN_TTL, '30d', 'REFRESH_TOKEN_TTL');
  const refreshIdleTtl = duration(environment.REFRESH_IDLE_TTL, '7d', 'REFRESH_IDLE_TTL');
  const brokers = String(environment.KAFKA_BROKERS ?? 'localhost:9092').split(',').map((value) => value.trim()).filter(Boolean);
  if (brokers.length === 0 || brokers.some((broker) => !broker.includes(':'))) throw new Error('KAFKA_BROKERS must contain host:port values');

  return {
    ...environment,
    NODE_ENV: nodeEnvironment,
    PORT: port,
    API_PREFIX: String(environment.API_PREFIX ?? 'api/v1').replace(/^\/+|\/+$/g, ''),
    WEB_ORIGIN: requiredUrl({ WEB_ORIGIN: environment.WEB_ORIGIN ?? 'http://localhost:3000' }, 'WEB_ORIGIN'),
    DATABASE_URL: requiredUrl(environment, 'DATABASE_URL'),
    REDIS_URL: requiredUrl(environment, 'REDIS_URL'),
    KAFKA_BROKERS: brokers.join(','),
    JWT_SECRET: jwtSecret,
    JWT_TTL: jwtTtl,
    JWT_ISSUER: boundedIdentifier(environment.JWT_ISSUER, 'tamo-identity', 'JWT_ISSUER'),
    JWT_AUDIENCE: boundedIdentifier(environment.JWT_AUDIENCE, 'tamo-platform', 'JWT_AUDIENCE'),
    EMAIL_VERIFICATION_TTL: emailVerificationTtl,
    PASSWORD_RECOVERY_TTL: passwordRecoveryTtl,
    REFRESH_TOKEN_TTL: refreshTokenTtl,
    REFRESH_IDLE_TTL: refreshIdleTtl,
    RATE_LIMIT_HMAC_SECRET: rateLimitSecret,
    AUTH_RATE_LIMIT_MAX: boundedInteger(environment.AUTH_RATE_LIMIT_MAX, 10, 1, 1000, 'AUTH_RATE_LIMIT_MAX'),
    AUTH_RATE_LIMIT_WINDOW_MS: boundedInteger(environment.AUTH_RATE_LIMIT_WINDOW_MS, 60000, 1000, 3600000, 'AUTH_RATE_LIMIT_WINDOW_MS'),
    JSON_BODY_LIMIT: String(environment.JSON_BODY_LIMIT ?? '100kb'),
    DATABASE_POOL_MAX: boundedInteger(environment.DATABASE_POOL_MAX, 10, 1, 50, 'DATABASE_POOL_MAX'),
    DATABASE_CONNECT_TIMEOUT_MS: boundedInteger(environment.DATABASE_CONNECT_TIMEOUT_MS, 3000, 100, 30000, 'DATABASE_CONNECT_TIMEOUT_MS'),
    DATABASE_IDLE_TIMEOUT_MS: boundedInteger(environment.DATABASE_IDLE_TIMEOUT_MS, 30000, 1000, 300000, 'DATABASE_IDLE_TIMEOUT_MS'),
    REDIS_CONNECT_TIMEOUT_MS: boundedInteger(environment.REDIS_CONNECT_TIMEOUT_MS, 2000, 100, 30000, 'REDIS_CONNECT_TIMEOUT_MS'),
    KAFKA_CONNECT_TIMEOUT_MS: boundedInteger(environment.KAFKA_CONNECT_TIMEOUT_MS, 3000, 100, 30000, 'KAFKA_CONNECT_TIMEOUT_MS'),
    KAFKA_REQUEST_TIMEOUT_MS: boundedInteger(environment.KAFKA_REQUEST_TIMEOUT_MS, 5000, 500, 60000, 'KAFKA_REQUEST_TIMEOUT_MS'),
  };
}

function duration(value:unknown,fallback:string,key:string):string{const parsed=String(value??fallback);if(!/^\d+[mhd]$/.test(parsed))throw new Error(`${key} must use m, h, or d units`);return parsed;}

function boundedIdentifier(value: unknown, fallback: string, key: string): string {
  const parsed = String(value ?? fallback).trim();
  if (!/^[A-Za-z0-9._:-]{1,128}$/.test(parsed)) throw new Error(`${key} must be a 1-128 character identifier`);
  return parsed;
}

function boundedInteger(value: unknown, fallback: number, minimum: number, maximum: number, key: string): number {
  const parsed = Number(value ?? fallback);
  if (!Number.isInteger(parsed) || parsed < minimum || parsed > maximum) throw new Error(`${key} must be between ${minimum} and ${maximum}`);
  return parsed;
}
