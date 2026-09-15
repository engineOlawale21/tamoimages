import { Module } from '@nestjs/common';
import { AuthModule } from '../auth/auth.module';
import { DatabaseModule } from '../database/database.module';
import { ComplianceController, PrivacyNoticeController } from './compliance.controller';
import { DataSubjectRequestPort } from './data-subject-request.port';
import { PostgresDataSubjectRequestService } from './postgres-data-subject-request.service';
import { PrivacyService } from './privacy.service';

@Module({imports:[AuthModule,DatabaseModule],controllers:[ComplianceController,PrivacyNoticeController],providers:[PrivacyService,{provide:DataSubjectRequestPort,useClass:PostgresDataSubjectRequestService}]})
export class ComplianceModule{}
