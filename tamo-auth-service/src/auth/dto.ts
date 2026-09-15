import { ApiProperty } from '@nestjs/swagger';
import { Transform } from 'class-transformer';
import { Equals, IsEmail, IsEnum, IsString, MaxLength, MinLength } from 'class-validator';
export enum AccountRole { BUYER='buyer', CONTRIBUTOR='contributor' }
export class RegisterDto {
  @ApiProperty() @Transform(({ value }) => typeof value === 'string' ? value.trim().toLowerCase() : value) @IsEmail() @MaxLength(254) email!: string;
  @ApiProperty() @Transform(({ value }) => typeof value === 'string' ? value.trim() : value) @IsString() @MinLength(1) @MaxLength(100) firstName!: string;
  @ApiProperty() @Transform(({ value }) => typeof value === 'string' ? value.trim() : value) @IsString() @MinLength(1) @MaxLength(100) lastName!: string;
  @ApiProperty({ minLength: 12, maxLength: 128 }) @IsString() @MinLength(12) @MaxLength(128) password!: string;
  @ApiProperty({ enum: AccountRole }) @IsEnum(AccountRole) role!: AccountRole;
  @ApiProperty({example:'2026-09-09'}) @IsString() @MaxLength(64) noticeVersion!:string;
  @ApiProperty({description:'Confirms the current privacy notice was presented; this is not consent to optional processing'}) @Equals(true) privacyNoticeAcknowledged!:boolean;
}
export class LoginDto {
  @ApiProperty() @Transform(({ value }) => typeof value === 'string' ? value.trim().toLowerCase() : value) @IsEmail() @MaxLength(254) email!: string;
  @ApiProperty() @IsString() @MaxLength(128) password!: string;
}
export class TokenDto { @ApiProperty({ minLength: 32, maxLength: 512 }) @IsString() @MinLength(32) @MaxLength(512) token!: string; }
export class ForgotPasswordDto { @ApiProperty() @Transform(({value})=>typeof value==='string'?value.trim().toLowerCase():value) @IsEmail() @MaxLength(254) email!: string; }
export class ResetPasswordDto extends TokenDto { @ApiProperty({minLength:12,maxLength:128}) @IsString() @MinLength(12) @MaxLength(128) password!: string; }
