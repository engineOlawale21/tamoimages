import { Injectable, OnApplicationShutdown } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import Redis from 'ioredis';

@Injectable()
export class RedisService implements OnApplicationShutdown {
  private readonly client: Redis;
  private readonly prefix: string;

  constructor(config: ConfigService) {
    this.prefix = `${config.getOrThrow<string>('NODE_ENV')}:tamo-auth-service`;
    this.client = new Redis(config.getOrThrow<string>('REDIS_URL'), {
      lazyConnect: true,
      connectTimeout: config.getOrThrow<number>('REDIS_CONNECT_TIMEOUT_MS'),
      maxRetriesPerRequest: 1,
      retryStrategy: () => null,
      enableOfflineQueue: false,
    });
    this.client.on('error', () => undefined);
  }

  key(scope: string, identifier: string): string {
    return `${this.prefix}:${scope}:${identifier}`;
  }

  async ping(): Promise<void> {
    await this.connect();
    const response = await this.client.ping();
    if (response !== 'PONG') throw new Error('Redis ping failed');
  }

  async consumeLimit(key: string, windowMilliseconds: number): Promise<{ count: number; ttlMilliseconds: number }> {
    await this.connect();
    const result = await this.client.eval(`
      local count = redis.call('INCR', KEYS[1])
      if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
      local ttl = redis.call('PTTL', KEYS[1])
      return {count, ttl}
    `, 1, key, windowMilliseconds) as [number, number];
    return { count: Number(result[0]), ttlMilliseconds: Math.max(Number(result[1]), 0) };
  }

  async cacheRevocation(identifier:string,ttlMilliseconds:number):Promise<void>{await this.connect();await this.client.set(this.key('revoked-session',identifier),'1','PX',ttlMilliseconds);}
  async isRevoked(identifier:string):Promise<boolean>{await this.connect();return await this.client.exists(this.key('revoked-session',identifier))===1;}

  private async connect() {
    if (this.client.status === 'wait' || this.client.status === 'end') await this.client.connect();
  }

  async onApplicationShutdown() {
    if (this.client.status === 'ready') await this.client.quit();
    else this.client.disconnect();
  }
}
