import { Inject, Injectable } from '@nestjs/common';
import { Pool } from 'pg';
import { DATABASE_POOL } from './database/database.constants';
import { KafkaService } from './integrations/kafka/kafka.service';
import { RedisService } from './integrations/redis/redis.service';

type Check = { name: string; status: 'up' | 'down' };

@Injectable()
export class HealthService {
  constructor(
    @Inject(DATABASE_POOL) private readonly database: Pool,
    private readonly redis: RedisService,
    private readonly kafka: KafkaService,
  ) {}

  async readiness() {
    const checks = await Promise.all([
      this.check('postgres', () => this.database.query('SELECT 1').then(() => undefined)),
      this.check('redis', () => this.redis.ping()),
      this.check('kafka', () => this.kafka.ready()),
    ]);
    return { status: checks.every((check) => check.status === 'up') ? 'ready' : 'not-ready', checks };
  }

  private async check(name: string, operation: () => Promise<void>): Promise<Check> {
    try {
      await operation();
      return { name, status: 'up' };
    } catch {
      return { name, status: 'down' };
    }
  }
}
