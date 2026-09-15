import { Injectable, NestMiddleware } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { NextFunction, Request, Response } from 'express';
import { createHmac } from 'node:crypto';
import { RedisService } from '../../integrations/redis/redis.service';

type ContextRequest = Request & { correlationId?: string };

@Injectable()
export class RateLimitMiddleware implements NestMiddleware {
  private readonly limit: number;
  private readonly windowMilliseconds: number;
  private readonly hmacSecret: string;

  constructor(config: ConfigService, private readonly redis: RedisService) {
    this.limit = config.getOrThrow<number>('AUTH_RATE_LIMIT_MAX');
    this.windowMilliseconds = config.getOrThrow<number>('AUTH_RATE_LIMIT_WINDOW_MS');
    this.hmacSecret = config.getOrThrow<string>('RATE_LIMIT_HMAC_SECRET');
  }

  async use(request: ContextRequest, response: Response, next: NextFunction) {
    const address = request.socket.remoteAddress ?? 'unknown';
    const subject = createHmac('sha256', this.hmacSecret).update(address).digest('hex');
    const email=typeof request.body?.email==='string'?request.body.email.trim().toLowerCase():undefined;
    const keys=[this.redis.key('rate-limit',`auth:ip:${subject}`)];
    if(email)keys.push(this.redis.key('rate-limit',`auth:account:${createHmac('sha256',this.hmacSecret).update(email).digest('hex')}`));
    try {
      const results=await Promise.all(keys.map((key)=>this.redis.consumeLimit(key,this.windowMilliseconds)));
      const result=results.reduce((highest,current)=>current.count>highest.count?current:highest);
      response.setHeader('RateLimit-Limit', this.limit);
      response.setHeader('RateLimit-Remaining', Math.max(this.limit - result.count, 0));
      response.setHeader('RateLimit-Reset', Math.ceil(result.ttlMilliseconds / 1000));
      if (result.count <= this.limit) return next();
      response.setHeader('Retry-After', Math.ceil(result.ttlMilliseconds / 1000));
      return this.problem(response, request, 429, 'Too Many Requests', 'Too many authentication attempts. Try again later.');
    } catch {
      return this.problem(response, request, 503, 'Service Unavailable', 'Authentication protection is temporarily unavailable.');
    }
  }

  private problem(response: Response, request: ContextRequest, status: number, title: string, detail: string) {
    response.status(status).type('application/problem+json').send({
      type: `https://httpstatuses.io/${status}`, title, status, detail,
      instance: request.originalUrl, correlationId: request.correlationId,
    });
  }
}
