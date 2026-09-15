import { Injectable, NestMiddleware } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { NextFunction, Request, Response } from 'express';
@Injectable()
export class SameOriginMiddleware implements NestMiddleware{
  private readonly expected:string;
  constructor(config:ConfigService){this.expected=new URL(config.getOrThrow<string>('WEB_ORIGIN')).origin;}
  use(request:Request,response:Response,next:NextFunction){const origin=request.header('origin');if(origin===this.expected)return next();return response.status(403).type('application/problem+json').send({type:'https://httpstatuses.io/403',title:'Forbidden',status:403,detail:'A valid same-origin request is required.',instance:request.originalUrl});}
}
