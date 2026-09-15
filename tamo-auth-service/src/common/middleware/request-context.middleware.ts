import { Injectable, NestMiddleware } from '@nestjs/common';
import { NextFunction, Request, Response } from 'express';
import { randomUUID } from 'node:crypto';
import { writeStructuredLog } from '../logging/structured-logger';

type ContextRequest = Request & { correlationId?: string };

@Injectable()
export class RequestContextMiddleware implements NestMiddleware {
  use(request: ContextRequest, response: Response, next: NextFunction) {
    const supplied = request.header('x-correlation-id');
    const correlationId = supplied && /^[A-Za-z0-9._:-]{1,128}$/.test(supplied) ? supplied : randomUUID();
    const startedAt = performance.now();
    request.correlationId = correlationId;
    response.setHeader('x-correlation-id', correlationId);
    response.setHeader('x-content-type-options', 'nosniff');
    response.setHeader('x-frame-options', 'DENY');
    response.setHeader('referrer-policy', 'no-referrer');
    response.setHeader('permissions-policy', 'camera=(), microphone=(), geolocation=()');
    response.setHeader('cache-control', 'no-store');
    response.removeHeader('x-powered-by');
    response.on('finish', () => writeStructuredLog({
      event: 'http.request.completed', correlationId, method: request.method,
      route: request.route?.path ?? request.path, status: response.statusCode,
      durationMs: Math.round((performance.now() - startedAt) * 100) / 100,
    }));
    next();
  }
}
