export type DataSubjectRequestKind = 'access' | 'correction' | 'portability' | 'deletion' | 'restriction' | 'objection' | 'consent_withdrawal' | 'complaint';

export type DataSubjectRequest = {
  requestId: string;
  subjectUserId: string;
  kind: DataSubjectRequestKind;
  receivedAt: Date;
  correlationId: string;
};

export type DataSubjectRequestReceipt = {
  requestId: string;
  status: 'accepted' | 'requires-verification';
  acceptedAt: Date;
};

/**
 * Compliance boundary for NDPR data-subject workflows. Implementations must
 * authenticate the subject, preserve an audit trail, and avoid logging request data.
 */
export abstract class DataSubjectRequestPort {
  abstract submit(request: DataSubjectRequest): Promise<DataSubjectRequestReceipt>;
  abstract listForSubject(subjectUserId:string):Promise<{requests:unknown[]}>;
}
