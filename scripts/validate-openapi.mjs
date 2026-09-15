import SwaggerParser from '@apidevtools/swagger-parser';

const contracts = [
  'tamo-auth-service/generated/openapi.json',
  'tamo-media-service/api/openapi.yaml',
];

for (const contract of contracts) {
  await SwaggerParser.validate(contract);
  process.stdout.write(`validated ${contract}\n`);
}
