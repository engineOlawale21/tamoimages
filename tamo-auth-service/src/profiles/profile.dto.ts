import { ApiPropertyOptional } from '@nestjs/swagger';
import { Transform } from 'class-transformer';
import { IsOptional, IsString, Matches, MaxLength } from 'class-validator';
export class UpdateProfileDto{
  @ApiPropertyOptional({maxLength:100}) @IsOptional() @Transform(({value})=>typeof value==='string'?value.trim():value) @IsString() @MaxLength(100) displayName?:string;
  @ApiPropertyOptional({maxLength:160}) @IsOptional() @Transform(({value})=>typeof value==='string'?value.trim():value) @IsString() @MaxLength(160) organisation?:string;
  @ApiPropertyOptional({maxLength:100}) @IsOptional() @Transform(({value})=>typeof value==='string'?value.trim():value) @IsString() @MaxLength(100) profession?:string;
  @ApiPropertyOptional({maxLength:1000}) @IsOptional() @Transform(({value})=>typeof value==='string'?value.trim():value) @IsString() @MaxLength(1000) biography?:string;
  @ApiPropertyOptional({pattern:'^[A-Z]{2}$'}) @IsOptional() @Transform(({value})=>typeof value==='string'?value.trim().toUpperCase():value) @Matches(/^[A-Z]{2}$/) countryCode?:string;
  @ApiPropertyOptional({maxLength:100}) @IsOptional() @Transform(({value})=>typeof value==='string'?value.trim():value) @IsString() @MaxLength(100) city?:string;
}
