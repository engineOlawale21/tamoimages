import { ArgumentsHost, BadRequestException } from '@nestjs/common';
import { ProblemDetailsFilter } from '../src/common/filters/problem-details.filter';

describe('ProblemDetailsFilter', () => {
  it('returns a correlated validation problem without leaking internals', () => {
    const response = {
      status: jest.fn().mockReturnThis(), type: jest.fn().mockReturnThis(), send: jest.fn(),
    };
    const request = { originalUrl: '/api/v1/auth/register', correlationId: 'request-123' };
    const host = {
      switchToHttp: () => ({ getResponse: () => response, getRequest: () => request }),
    } as unknown as ArgumentsHost;
    new ProblemDetailsFilter().catch(new BadRequestException({ message: ['email must be an email', 'password is too short'] }), host);
    expect(response.status).toHaveBeenCalledWith(400);
    expect(response.type).toHaveBeenCalledWith('application/problem+json');
    expect(response.send).toHaveBeenCalledWith(expect.objectContaining({
      status: 400, correlationId: 'request-123', instance: '/api/v1/auth/register',
      detail: 'email must be an email; password is too short',
    }));
  });
});
