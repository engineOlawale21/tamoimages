import { publicEnvironment } from '../env/public';
import { serverEnvironment } from '../env/server';
import { createIdentityApi } from './identity';
import { createMediaApi } from './media';

export function serverApis(fetchImplementation?: typeof fetch) {
  const publicConfig = publicEnvironment();
  const serverConfig = serverEnvironment();
  return {
    identity: createIdentityApi(publicConfig.NEXT_PUBLIC_IDENTITY_API, fetchImplementation, serverConfig.API_TIMEOUT_MS),
    media: createMediaApi(publicConfig.NEXT_PUBLIC_MEDIA_API, fetchImplementation, serverConfig.API_TIMEOUT_MS),
  };
}
