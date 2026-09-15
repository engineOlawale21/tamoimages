import { publicEnvironment } from '../env/public';
import { createIdentityApi } from './identity';
import { createMediaApi } from './media';

export function browserApis() {
  const environment = publicEnvironment();
  return {
    identity: createIdentityApi(environment.NEXT_PUBLIC_IDENTITY_API),
    media: createMediaApi(environment.NEXT_PUBLIC_MEDIA_API),
  };
}
