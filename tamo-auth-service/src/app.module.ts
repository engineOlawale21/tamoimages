import { MiddlewareConsumer, Module, NestModule, RequestMethod } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { AuthModule } from './auth/auth.module';
import { AuthController } from './auth/auth.controller';
import { RateLimitMiddleware } from './common/middleware/rate-limit.middleware';
import { RequestContextMiddleware } from './common/middleware/request-context.middleware';
import { SameOriginMiddleware } from './common/middleware/same-origin.middleware';
import { validateEnvironment } from './config/environment';
import { DatabaseModule } from './database/database.module';
import { HealthController } from './health.controller';
import { HealthService } from './health.service';
import { IntegrationsModule } from './integrations/integrations.module';
import { ProfilesModule } from './profiles/profiles.module';
import { ComplianceModule } from './compliance/compliance.module';

@Module({
  imports: [ConfigModule.forRoot({ isGlobal: true, cache: true, validate: validateEnvironment }), DatabaseModule, IntegrationsModule, AuthModule, ProfilesModule, ComplianceModule],
  controllers: [HealthController],
  providers: [HealthService],
})
export class AppModule implements NestModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(RequestContextMiddleware).forRoutes({path:'{*path}',method:RequestMethod.ALL});
    consumer.apply(RateLimitMiddleware).forRoutes(AuthController);
    consumer.apply(SameOriginMiddleware).forRoutes('auth/refresh','auth/logout');
  }
}
