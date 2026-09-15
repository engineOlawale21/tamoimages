import { ApiProperty } from '@nestjs/swagger';
import { AccountRole } from './dto';

export class AuthenticatedUserDto {
  @ApiProperty({ format: 'uuid' }) id!: string;
  @ApiProperty({ format: 'email' }) email!: string;
  @ApiProperty() firstName!: string;
  @ApiProperty() lastName!: string;
  @ApiProperty({ enum: AccountRole }) role!: AccountRole;
}

export class AuthResponseDto {
  @ApiProperty({ description: 'Short-lived bearer token' }) accessToken!: string;
  @ApiProperty({ type: AuthenticatedUserDto }) user!: AuthenticatedUserDto;
}
export class MessageResponseDto { @ApiProperty() message!: string; }
