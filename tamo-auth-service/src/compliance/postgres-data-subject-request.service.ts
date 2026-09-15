import { Inject, Injectable } from '@nestjs/common';
import { Pool } from 'pg';
import { DATABASE_POOL } from '../database/database.constants';
import { DataSubjectRequest, DataSubjectRequestPort, DataSubjectRequestReceipt } from './data-subject-request.port';

@Injectable()
export class PostgresDataSubjectRequestService implements DataSubjectRequestPort {
  constructor(@Inject(DATABASE_POOL) private readonly pool:Pool) {}
  async submit(request:DataSubjectRequest):Promise<DataSubjectRequestReceipt>{
    const result=await this.pool.query<{requestId:string;acceptedAt:Date}>(
      `INSERT INTO data_subject_requests(id,user_id,kind,correlation_id) VALUES($1,$2,$3,$4)
       RETURNING id AS "requestId",received_at AS "acceptedAt"`,
      [request.requestId,request.subjectUserId,request.kind,request.correlationId],
    );
    return {...result.rows[0],status:'accepted'};
  }
  async listForSubject(subjectUserId:string){
    const result=await this.pool.query(
      `SELECT id AS "requestId",kind,status,received_at AS "receivedAt",verified_at AS "verifiedAt",
              deadline_at AS "deadlineAt",decision,decided_at AS "decidedAt",fulfilled_at AS "fulfilledAt",exemption_code AS "exemptionCode"
       FROM data_subject_requests WHERE user_id=$1 ORDER BY received_at DESC`,[subjectUserId]);
    return {requests:result.rows};
  }
}
