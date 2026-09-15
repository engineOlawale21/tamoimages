import { Module } from '@nestjs/common';
import { JwtModule } from '@nestjs/jwt';
import { ConfigService } from '@nestjs/config';
import { AuditModule } from '../audit/audit.module';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';
import { BearerAuthGuard } from './bearer-auth.guard';
import { RolesGuard } from './roles.guard';
import { PostgresUsersRepository, UsersRepository } from './users.repository';
@Module({ imports: [AuditModule, JwtModule.registerAsync({ inject: [ConfigService], useFactory: (c: ConfigService) => ({ secret: c.getOrThrow('JWT_SECRET'), signOptions: { expiresIn: c.getOrThrow('JWT_TTL'), issuer: c.getOrThrow('JWT_ISSUER'), audience: c.getOrThrow('JWT_AUDIENCE') } }) })], controllers: [AuthController], providers: [AuthService,BearerAuthGuard,RolesGuard,{ provide: UsersRepository, useClass: PostgresUsersRepository }],exports:[JwtModule,BearerAuthGuard,RolesGuard,AuthService] })
export class AuthModule {}
