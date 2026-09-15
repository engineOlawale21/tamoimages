import { Injectable, OnApplicationShutdown } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Admin, Kafka, logLevel, Producer } from 'kafkajs';
import { randomUUID } from 'node:crypto';

export type DomainEvent<T> = {
  eventId: string;
  eventType: string;
  occurredAt: string;
  correlationId: string;
  producer: 'tamo-auth-service';
  data: T;
};

@Injectable()
export class KafkaService implements OnApplicationShutdown {
  private readonly admin: Admin;
  private readonly producer: Producer;
  private adminConnected = false;
  private producerConnected = false;

  constructor(config: ConfigService) {
    const kafka = new Kafka({
      clientId: 'tamo-auth-service',
      brokers: config.getOrThrow<string>('KAFKA_BROKERS').split(','),
      connectionTimeout: config.getOrThrow<number>('KAFKA_CONNECT_TIMEOUT_MS'),
      requestTimeout: config.getOrThrow<number>('KAFKA_REQUEST_TIMEOUT_MS'),
      retry: { retries: 3, initialRetryTime: 200 },
      logLevel: logLevel.NOTHING,
    });
    this.admin = kafka.admin();
    this.producer = kafka.producer({ allowAutoTopicCreation: false, idempotent: true, maxInFlightRequests: 1 });
  }

  async ready(): Promise<void> {
    if (!this.adminConnected) { await this.admin.connect(); this.adminConnected = true; }
    await this.admin.listTopics();
  }

  async publish<T>(topic: string, eventType: string, correlationId: string, data: T): Promise<DomainEvent<T>> {
    if (!this.producerConnected) { await this.producer.connect(); this.producerConnected = true; }
    const event: DomainEvent<T> = {
      eventId: randomUUID(), eventType, occurredAt: new Date().toISOString(), correlationId,
      producer: 'tamo-auth-service', data,
    };
    await this.producer.send({ topic, messages: [{ key: event.eventId, value: JSON.stringify(event), headers: { correlationId, eventType } }] });
    return event;
  }

  async onApplicationShutdown() {
    const operations: Promise<unknown>[] = [];
    if (this.producerConnected) operations.push(this.producer.disconnect());
    if (this.adminConnected) operations.push(this.admin.disconnect());
    await Promise.allSettled(operations);
  }
}
