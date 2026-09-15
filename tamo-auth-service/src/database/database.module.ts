import { Global, Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Pool } from 'pg';
import { DATABASE_POOL } from './database.constants';
import { DatabaseLifecycle } from './database-lifecycle.service';

@Global()
@Module({
  providers: [{
    provide: DATABASE_POOL,
    inject: [ConfigService],
    useFactory: (config: ConfigService) => new Pool({
      connectionString: config.getOrThrow<string>('DATABASE_URL'),
      max: config.getOrThrow<number>('DATABASE_POOL_MAX'),
      connectionTimeoutMillis: config.getOrThrow<number>('DATABASE_CONNECT_TIMEOUT_MS'),
      idleTimeoutMillis: config.getOrThrow<number>('DATABASE_IDLE_TIMEOUT_MS'),
      application_name: 'tamo-auth-service',
    }),
  }, DatabaseLifecycle],
  exports: [DATABASE_POOL],
})
export class DatabaseModule {}
