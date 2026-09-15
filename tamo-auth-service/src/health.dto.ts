import { ApiProperty } from '@nestjs/swagger';

export class LivenessResponseDto {
  @ApiProperty({ example: 'alive' }) status!: string;
  @ApiProperty({ example: 'tamo-auth-service' }) service!: string;
}

export class DependencyCheckDto {
  @ApiProperty() name!: string;
  @ApiProperty({ enum: ['up', 'down'] }) status!: 'up' | 'down';
}

export class ReadinessResponseDto {
  @ApiProperty({ enum: ['ready', 'not-ready'] }) status!: 'ready' | 'not-ready';
  @ApiProperty({ type: [DependencyCheckDto] }) checks!: DependencyCheckDto[];
}
