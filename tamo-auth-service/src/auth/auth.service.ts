import { Injectable, Optional, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { JwtService } from '@nestjs/jwt';
import * as bcrypt from 'bcrypt';
import { createHash, randomBytes, randomUUID } from 'node:crypto';
import { KafkaService } from '../integrations/kafka/kafka.service';
import { ForgotPasswordDto, LoginDto, RegisterDto, ResetPasswordDto } from './dto';
import { UserRecord, UsersRepository } from './users.repository';
import { SecurityAuditService } from '../audit/security-audit.service';
import { RedisService } from '../integrations/redis/redis.service';

type ClientContext = { correlationId: string; userAgent?: string; address?: string };
type SessionResult = { accessToken: string; refreshToken: string; refreshExpiresAt: Date; user: ReturnType<AuthService['publicUser']> };

@Injectable()
export class AuthService {
  private readonly dummyPasswordHash = bcrypt.hash('constant-time-invalid-account-password', 12);
  constructor(private readonly jwt:JwtService,private readonly users:UsersRepository,@Optional() private readonly config?:ConfigService,@Optional() private readonly kafka?:KafkaService,@Optional() private readonly audit?:SecurityAuditService,@Optional() private readonly redis?:RedisService){}

  async register(input:RegisterDto,correlationId='system'){
    const email=input.email.toLowerCase();
    try{
      const user=await this.users.create({email,firstName:input.firstName,lastName:input.lastName,role:input.role,passwordHash:await bcrypt.hash(input.password,12),registrationNoticeVersion:input.noticeVersion});
      const token=this.randomToken();
      await this.users.createVerificationToken(user.id,this.hash(token),this.after(this.setting('EMAIL_VERIFICATION_TTL','30m')));
      await this.kafka?.publish('identity.email-verification-requested.v1','identity.email-verification-requested.v1',correlationId,{userId:user.id,email,verificationToken:token});
    }catch(error){if(!this.isUniqueViolation(error))throw error;}
    await this.audit?.record({type:'registration_requested',outcome:'success',correlationId,subject:email});
    return {message:'If registration can be completed, verification instructions will be sent.'};
  }

  async resendVerification(email:string,correlationId='system'){
    const user=await this.users.findActiveByEmail(email.toLowerCase());
    if(user?.status==='pending_verification'){
      const token=this.randomToken();
      await this.users.createVerificationToken(user.id,this.hash(token),this.after(this.setting('EMAIL_VERIFICATION_TTL','30m')));
      await this.kafka?.publish('identity.email-verification-requested.v1','identity.email-verification-requested.v1',correlationId,{userId:user.id,email:user.email,verificationToken:token});
    }
    return {message:'If the account is eligible, verification instructions will be sent.'};
  }

  async verifyEmail(token:string){if(!await this.users.consumeVerificationToken(this.hash(token)))throw new UnauthorizedException('Verification token is invalid or expired');await this.audit?.record({type:'email_verified',outcome:'success'});return {message:'Email verified. You can now sign in.'};}

  async login(input:LoginDto,context:ClientContext={correlationId:'system'}):Promise<SessionResult>{
    const email=input.email.toLowerCase();
    const user=await this.users.findActiveByEmail(email);
    const matches=await bcrypt.compare(input.password,user?.passwordHash??await this.dummyPasswordHash);
    if(!user||!matches||user.status!=='active'){
      await this.audit?.record({type:'login',outcome:'failure',correlationId:context.correlationId,subject:email});
      throw new UnauthorizedException('Invalid credentials');
    }
    const session=await this.createSession(user,context);
    await this.audit?.record({type:'login',outcome:'success',correlationId:context.correlationId,userId:user.id});
    return session;
  }

  async refresh(refreshToken:string):Promise<SessionResult>{
    const tokenHash=this.hash(refreshToken);
    if(await this.redis?.isRevoked(tokenHash.toString('hex')))throw new UnauthorizedException('Session is invalid or expired');
    const nextToken=this.randomToken();
    const expiry=this.after(this.setting('REFRESH_TOKEN_TTL','30d'));
    const result=await this.users.rotateSession({tokenHash,nextTokenHash:this.hash(nextToken),nextExpiresAt:expiry,idleAfter:new Date(Date.now()-this.durationMs(this.setting('REFRESH_IDLE_TTL','7d')))});
    if(result.state==='reused'){await this.audit?.record({type:'session_reuse',outcome:'failure'});throw new UnauthorizedException('Session is invalid or expired');}
    if(result.state==='invalid')throw new UnauthorizedException('Session is invalid or expired');
    const user=await this.users.findById(result.session.userId);
    if(!user||user.status!=='active'){await this.users.revokeAllSessions(result.session.userId,'account_inactive');throw new UnauthorizedException('Session is invalid or expired');}
    await this.audit?.record({type:'session_refreshed',outcome:'success',userId:user.id});
    return {accessToken:await this.accessToken(user),refreshToken:nextToken,refreshExpiresAt:expiry,user:this.publicUser(user)};
  }

  async logout(refreshToken:string){const tokenHash=this.hash(refreshToken);await this.users.revokeSession(tokenHash,'logout');await this.redis?.cacheRevocation(tokenHash.toString('hex'),this.durationMs(this.setting('REFRESH_TOKEN_TTL','30d')));await this.audit?.record({type:'logout',outcome:'success'});return {message:'Signed out.'};}
  async logoutAll(userId:string){await this.users.revokeAllSessions(userId,'logout_all');await this.audit?.record({type:'logout_all',outcome:'success',userId});return {message:'All sessions have been signed out.'};}

  async forgotPassword(input:ForgotPasswordDto,correlationId='system'){
    const user=await this.users.findActiveByEmail(input.email.toLowerCase());
    if(user&&user.status!=='suspended'){
      const token=this.randomToken();
      await this.users.createRecoveryToken(user.id,this.hash(token),this.after(this.setting('PASSWORD_RECOVERY_TTL','15m')));
      await this.kafka?.publish('identity.password-recovery-requested.v1','identity.password-recovery-requested.v1',correlationId,{userId:user.id,email:user.email,recoveryToken:token});
    }
    await this.audit?.record({type:'password_recovery_requested',outcome:'success',correlationId,subject:input.email});
    return {message:'If the account is eligible, password-reset instructions will be sent.'};
  }

  async resetPassword(input:ResetPasswordDto){const userId=await this.users.consumeRecoveryToken(this.hash(input.token),await bcrypt.hash(input.password,12));if(!userId)throw new UnauthorizedException('Recovery token is invalid or expired');await this.audit?.record({type:'password_reset',outcome:'success',userId});return {message:'Password reset. Sign in again on every device.'};}
  async requestAccountDeletion(userId:string){await this.users.requestDeletion(userId);await this.audit?.record({type:'account_deletion_requested',outcome:'success',userId});return {message:'Account deletion request accepted and all sessions revoked.'};}

  private async createSession(user:UserRecord,context:ClientContext):Promise<SessionResult>{const refreshToken=this.randomToken();const expiry=this.after(this.setting('REFRESH_TOKEN_TTL','30d'));await this.users.createSession({userId:user.id,familyId:randomUUID(),tokenHash:this.hash(refreshToken),expiresAt:expiry,userAgentHash:context.userAgent?this.hash(context.userAgent):undefined,ipPrefixHash:context.address?this.hash(context.address):undefined});return {accessToken:await this.accessToken(user),refreshToken,refreshExpiresAt:expiry,user:this.publicUser(user)};}
  private async accessToken(user:UserRecord){return this.jwt.signAsync({sub:user.id,role:user.role,jti:randomUUID()});}
  private publicUser(user:UserRecord){return {id:user.id,email:user.email,firstName:user.firstName,lastName:user.lastName,role:user.role};}
  private randomToken(){return randomBytes(32).toString('base64url');}
  private hash(value:string){return createHash('sha256').update(value).digest();}
  private setting(key:string,fallback:string){return this.config?.get<string>(key)??fallback;}
  private after(duration:string){return new Date(Date.now()+this.durationMs(duration));}
  private durationMs(duration:string){const match=/^(\d+)([mhd])$/.exec(duration);if(!match)throw new Error('Invalid authentication duration');const units:{[key:string]:number}={m:60000,h:3600000,d:86400000};return Number(match[1])*units[match[2]];}
  private isUniqueViolation(error:unknown):boolean{return typeof error==='object'&&error!==null&&'code' in error&&error.code==='23505';}
}
