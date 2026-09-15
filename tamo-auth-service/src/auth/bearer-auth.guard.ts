import { CanActivate, ExecutionContext, Injectable, UnauthorizedException } from '@nestjs/common';
import { JwtService } from '@nestjs/jwt';
import { AuthenticatedPrincipal, AuthenticatedRequest } from './authenticated-request';
@Injectable()
export class BearerAuthGuard implements CanActivate{
  constructor(private readonly jwt:JwtService){}
  async canActivate(context:ExecutionContext){const request=context.switchToHttp().getRequest<AuthenticatedRequest>();const header=request.header('authorization');if(!header?.startsWith('Bearer '))throw new UnauthorizedException('Authentication required');try{request.user=await this.jwt.verifyAsync<AuthenticatedPrincipal>(header.slice(7));return true;}catch{throw new UnauthorizedException('Authentication required');}}
}
