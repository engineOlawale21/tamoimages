import { Body, Controller, Get, HttpCode, Post, Req, UseGuards } from '@nestjs/common';
import { ApiAcceptedResponse, ApiBearerAuth, ApiOkResponse, ApiTags } from '@nestjs/swagger';
import { randomUUID } from 'node:crypto';
import { AuthenticatedRequest } from '../auth/authenticated-request';
import { BearerAuthGuard } from '../auth/bearer-auth.guard';
import { CreateDataSubjectRequestDto } from './data-subject-request.dto';
import { DataSubjectRequestPort } from './data-subject-request.port';
import { PrivacyService } from './privacy.service';
import { RecordConsentDto } from './privacy.dto';

type ContextRequest=AuthenticatedRequest&{correlationId?:string};
@ApiTags('privacy') @ApiBearerAuth() @UseGuards(BearerAuthGuard) @Controller('privacy/requests')
export class ComplianceController {
  constructor(private readonly requests:DataSubjectRequestPort,private readonly privacy:PrivacyService){}
  @Post() @HttpCode(202) @ApiAcceptedResponse({description:'Privacy request accepted for authenticated processing'})
  submit(@Body() body:CreateDataSubjectRequestDto,@Req() request:ContextRequest){
    const now=new Date();
    return this.requests.submit({requestId:randomUUID(),subjectUserId:request.user.sub,kind:body.kind,receivedAt:now,correlationId:request.correlationId??randomUUID()});
  }

  @Get() @ApiOkResponse({description:'Authenticated subject request history without request payloads'})
  list(@Req() request:ContextRequest){return this.requests.listForSubject(request.user.sub);}

  @Get('consents') @ApiOkResponse({description:'Latest granular consent choice for each optional purpose'})
  consents(@Req() request:ContextRequest){return this.privacy.currentChoices(request.user.sub);}

  @Post('consents') @HttpCode(200) @ApiOkResponse({description:'Consent choice or withdrawal recorded'})
  consent(@Body() body:RecordConsentDto,@Req() request:ContextRequest){return this.privacy.recordConsent(request.user.sub,body.purpose,body.granted,body.noticeVersion,request.correlationId??randomUUID());}
}

@ApiTags('privacy') @Controller('privacy/notices')
export class PrivacyNoticeController {
  constructor(private readonly privacy:PrivacyService){}
  @Get('current') @ApiOkResponse({description:'Current versioned privacy notice'})
  current(){return this.privacy.currentNotice();}
}
