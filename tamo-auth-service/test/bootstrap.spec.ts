import type { INestApplication } from '@nestjs/common';
import type { OpenAPIObject } from '@nestjs/swagger';

describe('application bootstrap', () => {
  let application: INestApplication | undefined;
  let createApplication: () => Promise<INestApplication>;
  let createOpenApiDocument: (app: INestApplication) => OpenAPIObject;

  beforeAll(async () => {
    Object.assign(process.env, {
      NODE_ENV: 'test', PORT: '4000', API_PREFIX: 'api/v1', WEB_ORIGIN: 'http://localhost:3000',
      JWT_SECRET: 'test-jwt-secret-with-at-least-32-characters',
      RATE_LIMIT_HMAC_SECRET: 'test-rate-secret-with-at-least-32-characters',
      DATABASE_URL: 'postgresql://tamo:tamo_local@localhost:5433/tamo_identity',
      REDIS_URL: 'redis://localhost:6379', KAFKA_BROKERS: 'localhost:9092',
    });
    ({ createApplication, createOpenApiDocument } = await import('../src/bootstrap'));
  });

  afterEach(async () => {
    await application?.close();
    application = undefined;
  });

  it('initializes the complete module graph and generates its API contract', async () => {
    application = await createApplication();
    await application.init();
    const document = createOpenApiDocument(application);
    expect(document.paths['/api/v1/health/live']).toBeDefined();
    expect(document.paths['/api/v1/auth/register']).toBeDefined();
  });
});
