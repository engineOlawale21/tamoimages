import { BadRequestException, Inject, Injectable, NotFoundException } from '@nestjs/common';
import { Pool } from 'pg';
import { DATABASE_POOL } from '../database/database.constants';
import { ConsentPurposeDto } from './privacy.dto';

@Injectable()
export class PrivacyService {
  constructor(@Inject(DATABASE_POOL) private readonly pool:Pool){}

  async currentNotice(){
    const result=await this.pool.query(`SELECT version,title,content,effective_at AS "effectiveAt" FROM privacy_notices WHERE effective_at<=now() AND retired_at IS NULL ORDER BY effective_at DESC LIMIT 1`);
    if(!result.rowCount)throw new NotFoundException('No active privacy notice is available');
    return result.rows[0];
  }

  async recordConsent(userId:string,purpose:ConsentPurposeDto,granted:boolean,noticeVersion:string,correlationId:string){
    const notice=await this.pool.query('SELECT 1 FROM privacy_notices WHERE version=$1 AND effective_at<=now()',[noticeVersion]);
    if(!notice.rowCount)throw new BadRequestException('Privacy notice version is invalid');
    const result=await this.pool.query(
      `INSERT INTO consent_records(user_id,notice_version,purpose,granted,channel,correlation_id)
       VALUES($1,$2,$3,$4,'account_api',$5)
       RETURNING purpose,granted,notice_version AS "noticeVersion",recorded_at AS "recordedAt"`,
      [userId,noticeVersion,purpose,granted,correlationId],
    );
    return result.rows[0];
  }

  async currentChoices(userId:string){
    const result=await this.pool.query(
      `SELECT DISTINCT ON (purpose) purpose,granted,notice_version AS "noticeVersion",recorded_at AS "recordedAt"
       FROM consent_records WHERE user_id=$1 ORDER BY purpose,recorded_at DESC`,[userId]);
    return {choices:result.rows};
  }
}
