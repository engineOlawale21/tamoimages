import { spawnSync } from 'node:child_process';

const windows = process.platform === 'win32';
const npmCommand = windows ? process.env.ComSpec ?? 'cmd.exe' : 'npm';
const environment = {
  ...process.env,
  NEXT_PUBLIC_IDENTITY_API:
    process.env.NEXT_PUBLIC_IDENTITY_API ?? 'http://localhost:4000/api/v1',
  NEXT_PUBLIC_MEDIA_API:
    process.env.NEXT_PUBLIC_MEDIA_API ?? 'http://localhost:5000/api/v1',
  NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN:
    process.env.NEXT_PUBLIC_MEDIA_STORAGE_ORIGIN ?? 'http://localhost:9000',
  API_TIMEOUT_MS: process.env.API_TIMEOUT_MS ?? '5000',
};

for (const args of [
  ['run', 'lint', '--workspaces', '--if-present'],
  ['test', '--workspaces', '--if-present'],
  ['run', 'build', '--workspaces', '--if-present'],
]) {
  const commandArgs = windows
    ? ['/d', '/s', '/c', `npm.cmd ${args.join(' ')}`]
    : args;
  const result = spawnSync(npmCommand, commandArgs, {
    env: environment,
    stdio: 'inherit',
  });

  if (result.error) throw result.error;
  if (result.status !== 0) process.exit(result.status ?? 1);
}
