import { ApiProperty } from '@nestjs/swagger';
import { IsEnum } from 'class-validator';

export enum DataSubjectRequestKindDto {
  ACCESS='access', CORRECTION='correction', PORTABILITY='portability', DELETION='deletion',
  RESTRICTION='restriction', OBJECTION='objection', CONSENT_WITHDRAWAL='consent_withdrawal', COMPLAINT='complaint',
}

export class CreateDataSubjectRequestDto {
  @ApiProperty({enum:DataSubjectRequestKindDto}) @IsEnum(DataSubjectRequestKindDto)
  kind!: DataSubjectRequestKindDto;
}
