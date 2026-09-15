import { writeStructuredLog } from '../src/common/logging/structured-logger';

describe('structured logger privacy boundary',()=>{
  it('keeps only approved operational fields',()=>{const output=jest.spyOn(process.stdout,'write').mockImplementation(()=>true);writeStructuredLog({event:'login.failed',severity:'warn',correlationId:'corr-1',password:'secret',email:'person@example.com',authorization:'Bearer token',cookie:'session=value'});const line=String(output.mock.calls[0][0]);expect(JSON.parse(line)).toMatchObject({service:'tamo-auth-service',event:'login.failed',severity:'warn',correlationId:'corr-1'});expect(line).not.toContain('secret');expect(line).not.toContain('person@example.com');expect(line).not.toContain('Bearer');expect(line).not.toContain('session=value');output.mockRestore();});
});
