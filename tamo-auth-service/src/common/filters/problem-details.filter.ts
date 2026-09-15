import { ArgumentsHost, Catch, ExceptionFilter, HttpException, HttpStatus } from '@nestjs/common';
import { Request, Response } from 'express';

@Catch()
export class ProblemDetailsFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost) {
    const response = host.switchToHttp().getResponse<Response>();
    const request = host.switchToHttp().getRequest<Request & { correlationId?: string }>();
    const status = exception instanceof HttpException ? exception.getStatus() : HttpStatus.INTERNAL_SERVER_ERROR;
    const raw = exception instanceof HttpException ? exception.getResponse() : undefined;
    const detail = status >= 500 ? 'The service could not complete the request.' : this.messageFrom(raw);
    response.status(status).type('application/problem+json').send({
      type: `https://httpstatuses.io/${status}`, title: HttpStatus[status] ?? 'Error', status, detail,
      instance: request.originalUrl, correlationId: request.correlationId,
    });
  }

  private messageFrom(value: unknown): string {
    if (typeof value === 'string') return value;
    if (!value || typeof value !== 'object') return 'The request could not be completed.';
    const message = (value as { message?: unknown }).message;
    return Array.isArray(message) ? message.join('; ') : String(message ?? 'The request could not be completed.');
  }
}
