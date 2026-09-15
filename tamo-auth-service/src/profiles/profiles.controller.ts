import { Body, Controller, Delete, Get, Patch, Req, UseGuards } from '@nestjs/common';
import { ApiBearerAuth, ApiOkResponse, ApiTags } from '@nestjs/swagger';
import { AuthenticatedRequest } from '../auth/authenticated-request';
import { BearerAuthGuard } from '../auth/bearer-auth.guard';
import { UpdateProfileDto } from './profile.dto';
import { ProfilesRepository } from './profiles.repository';
import { AuthService } from '../auth/auth.service';
@ApiTags('profiles') @ApiBearerAuth() @UseGuards(BearerAuthGuard) @Controller('profiles')
export class ProfilesController{constructor(private readonly profiles:ProfilesRepository,private readonly auth:AuthService){}@Get('me') @ApiOkResponse({description:'Current role-specific profile'})get(@Req()request:AuthenticatedRequest){return this.profiles.get(request.user.sub,request.user.role);}@Patch('me') @ApiOkResponse({description:'Updated role-specific profile'})update(@Req()request:AuthenticatedRequest,@Body()body:UpdateProfileDto){return this.profiles.update(request.user.sub,request.user.role,body);}@Delete('me') @ApiOkResponse({description:'Account deletion requested and sessions revoked'})remove(@Req()request:AuthenticatedRequest){return this.auth.requestAccountDeletion(request.user.sub);}}
