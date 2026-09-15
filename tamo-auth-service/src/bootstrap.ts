import { INestApplication, ValidationPipe } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { DocumentBuilder, OpenAPIObject, SwaggerModule } from '@nestjs/swagger';
import { json, urlencoded } from 'express';
import { AppModule } from './app.module';
import { ProblemDetailsFilter } from './common/filters/problem-details.filter';

export async function createApplication(): Promise<INestApplication> {
  const app = await NestFactory.create(AppModule, { bufferLogs: true, bodyParser: false });
  const environment = app.get(ConfigService);
  const bodyLimit = environment.getOrThrow<string>('JSON_BODY_LIMIT');
  app.use(json({ limit: bodyLimit }));
  app.use(urlencoded({ extended: false, limit: bodyLimit }));
  app.enableShutdownHooks();
  app.setGlobalPrefix(environment.getOrThrow<string>('API_PREFIX'));
  app.enableCors({ origin: [environment.getOrThrow<string>('WEB_ORIGIN')], credentials: true });
  app.useGlobalPipes(new ValidationPipe({ whitelist: true, forbidNonWhitelisted: true, transform: true }));
  app.useGlobalFilters(new ProblemDetailsFilter());
  return app;
}

export function createOpenApiDocument(app: INestApplication): OpenAPIObject {
  const options = new DocumentBuilder()
    .setTitle('Tamo Identity API')
    .setDescription('Authentication, accounts and role management')
    .setVersion('1.0')
    .addBearerAuth()
    .addServer('http://localhost:4000', 'Local development')
    .build();
  return SwaggerModule.createDocument(app, options);
}

export function exposeSwagger(app: INestApplication) {
  SwaggerModule.setup('docs', app, createOpenApiDocument(app));
}
