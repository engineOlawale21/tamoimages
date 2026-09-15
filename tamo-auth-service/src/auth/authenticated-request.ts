import { Request } from 'express';
import { AccountRole } from './dto';
export type AuthenticatedPrincipal={sub:string;role:AccountRole;jti:string;iat:number;exp:number};
export type AuthenticatedRequest=Request&{user:AuthenticatedPrincipal;correlationId?:string};
