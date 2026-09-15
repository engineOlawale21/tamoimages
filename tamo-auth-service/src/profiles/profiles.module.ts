import { Module } from '@nestjs/common';
import { AuthModule } from '../auth/auth.module';
import { ProfilesController } from './profiles.controller';
import { ProfilesRepository } from './profiles.repository';
@Module({imports:[AuthModule],controllers:[ProfilesController],providers:[ProfilesRepository]})
export class ProfilesModule{}
