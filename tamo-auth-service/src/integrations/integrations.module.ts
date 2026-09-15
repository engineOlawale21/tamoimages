import { Global, Module } from '@nestjs/common';
import { KafkaService } from './kafka/kafka.service';
import { RedisService } from './redis/redis.service';

@Global()
@Module({ providers: [RedisService, KafkaService], exports: [RedisService, KafkaService] })
export class IntegrationsModule {}
