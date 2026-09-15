import { PrivacyService } from '../src/compliance/privacy.service';
import { ConsentPurposeDto } from '../src/compliance/privacy.dto';

describe('PrivacyService',()=>{
  it('records a granular consent withdrawal against a notice version',async()=>{
    const query=jest.fn()
      .mockResolvedValueOnce({rowCount:1,rows:[{}]})
      .mockResolvedValueOnce({rowCount:1,rows:[{purpose:'usage_analytics',granted:false,noticeVersion:'2026-09-09'}]});
    const service=new PrivacyService({query} as never);
    await expect(service.recordConsent('user-1',ConsentPurposeDto.USAGE_ANALYTICS,false,'2026-09-09','request-1')).resolves.toMatchObject({granted:false});
    expect(query.mock.calls[1][1]).toEqual(['user-1','2026-09-09','usage_analytics',false,'request-1']);
  });

  it('rejects an unknown notice version',async()=>{
    const service=new PrivacyService({query:jest.fn().mockResolvedValue({rowCount:0,rows:[]})} as never);
    await expect(service.recordConsent('user-1',ConsentPurposeDto.PRODUCT_UPDATES,true,'unknown','request-1')).rejects.toMatchObject({message:'Privacy notice version is invalid'});
  });
});
