import { ConfigService } from '@nestjs/config';
import { createApplication, exposeSwagger } from './bootstrap';

async function bootstrap() {
  const app = await createApplication();
  const environment = app.get(ConfigService);
  if (environment.get<string>('NODE_ENV') !== 'production') exposeSwagger(app);
  await app.listen(environment.getOrThrow<number>('PORT'));
}

void bootstrap();
