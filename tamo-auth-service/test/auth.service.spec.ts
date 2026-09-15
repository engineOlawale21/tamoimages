import { JwtService } from '@nestjs/jwt';
import { AuthService } from '../src/auth/auth.service';
import { AccountRole } from '../src/auth/dto';
import { RotationResult, SessionRecord, UserRecord, UsersRepository } from '../src/auth/users.repository';

class FakeUsersRepository implements UsersRepository {
  created?: Omit<UserRecord, 'id' | 'status'>;
  record?: UserRecord;
  verificationHash?: Buffer;
  sessionInput?: { userId: string; familyId: string; tokenHash: Buffer; expiresAt: Date };
  rotation: RotationResult = { state: 'invalid' };
  async create(user: Omit<UserRecord, 'id' | 'status'>): Promise<UserRecord> {
    this.created = user;
    this.record = { id: '01900000-0000-7000-8000-000000000001', status: 'pending_verification', ...user };
    return this.record;
  }
  async findActiveByEmail(): Promise<UserRecord | undefined> { return this.record; }
  async findById(): Promise<UserRecord | undefined> { return this.record; }
  async createVerificationToken(_userId: string, tokenHash: Buffer): Promise<void> { this.verificationHash = tokenHash; }
  async consumeVerificationToken(): Promise<boolean> { return true; }
  async createSession(input: { userId: string; familyId: string; tokenHash: Buffer; expiresAt: Date }): Promise<SessionRecord> { this.sessionInput=input; return { id: 'session-1', userId: input.userId, familyId: input.familyId, expiresAt: input.expiresAt }; }
  async rotateSession(): Promise<RotationResult> { return this.rotation; }
  async revokeSession(): Promise<void> {}
  async revokeAllSessions(): Promise<void> {}
  async createRecoveryToken(): Promise<void> {}
  async consumeRecoveryToken(): Promise<string | undefined> { return undefined; }
  async requestDeletion(): Promise<void> { if(this.record)this.record.status='deletion_requested'; }
}

describe('AuthService', () => {
  it('stores a password hash without retaining the plaintext password', async () => {
    const users = new FakeUsersRepository();
    const service = new AuthService(new JwtService({ secret: 'a-test-secret-that-is-long-enough' }), users);
    await service.register({ email: 'person@example.com', firstName: 'Ada', lastName: 'Okafor', role: AccountRole.CONTRIBUTOR, password: 'correct horse battery staple', noticeVersion:'2026-09-09', privacyNoticeAcknowledged:true });
    expect(users.created).toBeDefined();
    expect(users.created).not.toHaveProperty('password');
    expect(users.created?.passwordHash).not.toContain('correct horse battery staple');
    expect(users.verificationHash).toHaveLength(32);
  });

  it('does not issue credentials during registration', async () => {
    const service=new AuthService(new JwtService({secret:'a-test-secret-that-is-long-enough'}),new FakeUsersRepository());
    await expect(service.register({email:'person@example.com',firstName:'Ada',lastName:'Okafor',role:AccountRole.BUYER,password:'correct horse battery staple',noticeVersion:'2026-09-09',privacyNoticeAcknowledged:true})).resolves.toEqual({message:expect.any(String)});
  });

  it('rejects login until a valid account is active', async () => {
    const users=new FakeUsersRepository();
    const service=new AuthService(new JwtService({secret:'a-test-secret-that-is-long-enough'}),users);
    await service.register({email:'person@example.com',firstName:'Ada',lastName:'Okafor',role:AccountRole.CONTRIBUTOR,password:'correct horse battery staple',noticeVersion:'2026-09-09',privacyNoticeAcknowledged:true});
    await expect(service.login({email:'person@example.com',password:'correct horse battery staple'})).rejects.toMatchObject({message:'Invalid credentials'});
    users.record!.status='active';
    const result=await service.login({email:'person@example.com',password:'correct horse battery staple'});
    expect(result.accessToken).toEqual(expect.any(String));
    expect(result.refreshToken).toEqual(expect.any(String));
    expect(users.sessionInput?.tokenHash).toHaveLength(32);
    expect(users.sessionInput?.tokenHash.toString()).not.toContain(result.refreshToken);
  });

  it('rejects a reused refresh-token family', async () => {
    const users=new FakeUsersRepository();users.rotation={state:'reused'};
    const service=new AuthService(new JwtService({secret:'a-test-secret-that-is-long-enough'}),users);
    await expect(service.refresh('a'.repeat(43))).rejects.toMatchObject({message:'Session is invalid or expired'});
  });

  it('returns the same generic error for an unknown account', async () => {
    const service = new AuthService(new JwtService({ secret: 'a-test-secret-that-is-long-enough' }), new FakeUsersRepository());
    await expect(service.login({ email: 'missing@example.com', password: 'not-the-password' })).rejects.toMatchObject({ message: 'Invalid credentials' });
  });
});
