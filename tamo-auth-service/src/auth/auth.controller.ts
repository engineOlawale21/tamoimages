import { Body, Controller, HttpCode, Post, Req, Res, UnauthorizedException, UseGuards } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ApiAcceptedResponse, ApiCreatedResponse, ApiOkResponse, ApiTags } from '@nestjs/swagger';
import { Request, Response } from 'express';
import { AuthService } from './auth.service';
import { ForgotPasswordDto, LoginDto, RegisterDto, ResetPasswordDto, TokenDto } from './dto';
import { AuthResponseDto, MessageResponseDto } from './responses.dto';
import { AuthenticatedRequest } from './authenticated-request';
import { BearerAuthGuard } from './bearer-auth.guard';

type ContextRequest=Request&{correlationId?:string};
const refreshCookie='__Host-tamo-refresh';

@ApiTags('auth')
@Controller('auth')
export class AuthController {
  constructor(private readonly auth:AuthService,private readonly config:ConfigService){}

  @Post('register') @HttpCode(202) @ApiAcceptedResponse({type:MessageResponseDto})
  register(@Body() body:RegisterDto,@Req() request:ContextRequest){return this.auth.register(body,request.correlationId);}

  @Post('verify-email') @HttpCode(200) @ApiOkResponse({type:MessageResponseDto})
  verifyEmail(@Body() body:TokenDto){return this.auth.verifyEmail(body.token);}

  @Post('resend-verification') @HttpCode(202) @ApiAcceptedResponse({type:MessageResponseDto})
  resend(@Body() body:ForgotPasswordDto,@Req() request:ContextRequest){return this.auth.resendVerification(body.email,request.correlationId);}

  @Post('login') @HttpCode(200) @ApiOkResponse({type:AuthResponseDto})
  async login(@Body() body:LoginDto,@Req() request:ContextRequest,@Res({passthrough:true}) response:Response){const result=await this.auth.login(body,{correlationId:request.correlationId??'system',userAgent:request.header('user-agent'),address:request.socket.remoteAddress});return this.respondWithSession(response,result);}

  @Post('refresh') @HttpCode(200) @ApiOkResponse({type:AuthResponseDto})
  async refresh(@Req() request:Request,@Res({passthrough:true}) response:Response){const token=this.cookie(request,refreshCookie);if(!token)throw new UnauthorizedException('Session is invalid or expired');return this.respondWithSession(response,await this.auth.refresh(token));}

  @Post('logout') @HttpCode(200) @ApiOkResponse({type:MessageResponseDto})
  async logout(@Req() request:Request,@Res({passthrough:true}) response:Response){const token=this.cookie(request,refreshCookie);if(token)await this.auth.logout(token);response.clearCookie(refreshCookie,this.cookieOptions());return {message:'Signed out.'};}

  @Post('logout-all') @UseGuards(BearerAuthGuard) @HttpCode(200) @ApiOkResponse({type:MessageResponseDto})
  logoutAll(@Req() request:AuthenticatedRequest){return this.auth.logoutAll(request.user.sub);}

  @Post('forgot-password') @HttpCode(202) @ApiAcceptedResponse({type:MessageResponseDto})
  forgot(@Body() body:ForgotPasswordDto,@Req() request:ContextRequest){return this.auth.forgotPassword(body,request.correlationId);}

  @Post('reset-password') @HttpCode(200) @ApiOkResponse({type:MessageResponseDto})
  reset(@Body() body:ResetPasswordDto){return this.auth.resetPassword(body);}

  private respondWithSession(response:Response,result:Awaited<ReturnType<AuthService['login']>>){response.cookie(refreshCookie,result.refreshToken,{...this.cookieOptions(),expires:result.refreshExpiresAt});return {accessToken:result.accessToken,user:result.user};}
  private cookie(request:Request,name:string){const header=request.header('cookie')??'';for(const part of header.split(';')){const [key,...value]=part.trim().split('=');if(key===name)return decodeURIComponent(value.join('='));}return undefined;}
  private cookieOptions(){return {httpOnly:true,secure:this.config.get<string>('NODE_ENV')!=='development'&&this.config.get<string>('NODE_ENV')!=='test',sameSite:'lax' as const,path:'/'};}
}
