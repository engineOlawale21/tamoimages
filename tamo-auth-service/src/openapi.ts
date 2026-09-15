import SwaggerParser from '@apidevtools/swagger-parser';
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { createApplication, createOpenApiDocument } from './bootstrap';

async function exportOpenApi() {
  const app = await createApplication();
  try {
    await app.init();
    const document = createOpenApiDocument(app);
    const portableDocument: unknown = JSON.parse(JSON.stringify(document));
    await SwaggerParser.validate(portableDocument as Parameters<typeof SwaggerParser.validate>[0]);
    const outputDirectory = resolve(process.cwd(), 'generated');
    await mkdir(outputDirectory, { recursive: true });
    await writeFile(resolve(outputDirectory, 'openapi.json'), `${JSON.stringify(document, null, 2)}\n`, 'utf8');
    process.stdout.write('OpenAPI contract validated and exported to generated/openapi.json\n');
  } finally {
    await app.close();
  }
}

void exportOpenApi().catch((error: unknown) => {
  process.stderr.write(`OpenAPI export failed: ${error instanceof Error ? error.message : 'unknown error'}\n`);
  process.exitCode = 1;
});
