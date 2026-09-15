import { ApiProperty } from '@nestjs/swagger';
import { IsBoolean, IsEnum, IsString, MaxLength } from 'class-validator';

export enum ConsentPurposeDto {
  PRODUCT_UPDATES='product_updates',
  USAGE_ANALYTICS='usage_analytics',
}

export class RecordConsentDto {
  @ApiProperty({enum:ConsentPurposeDto}) @IsEnum(ConsentPurposeDto)
  purpose!:ConsentPurposeDto;
  @ApiProperty() @IsBoolean()
  granted!:boolean;
  @ApiProperty({maxLength:64}) @IsString() @MaxLength(64)
  noticeVersion!:string;
}
