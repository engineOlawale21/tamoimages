import { readFile, readdir } from 'node:fs/promises';
import { resolve } from 'node:path';
import { Pool, PoolClient } from 'pg';
import { validateEnvironment } from '../config/environment';

async function migrate() {
  const environment = validateEnvironment(process.env);
  const pool = new Pool({ connectionString: String(environment.DATABASE_URL), max: 1 });
  const client = await pool.connect();
  try {
    await client.query('CREATE TABLE IF NOT EXISTS schema_migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())');
    if (process.argv[2] === 'down') await rollback(client);
    else await apply(client);
  } finally {
    client.release();
    await pool.end();
  }
}

async function apply(client: PoolClient) {
  const directory = resolve(process.cwd(), 'migrations');
  const files = (await readdir(directory)).filter((name) => name.endsWith('.up.sql')).sort();
  for (const file of files) {
    const applied = await client.query('SELECT 1 FROM schema_migrations WHERE name = $1', [file]);
    if (applied.rowCount) continue;
    await client.query('BEGIN');
    try {
      await client.query(await readFile(resolve(directory, file), 'utf8'));
      await client.query('INSERT INTO schema_migrations(name) VALUES ($1)', [file]);
      await client.query('COMMIT');
      process.stdout.write(`applied ${file}\n`);
    } catch (error) {
      await client.query('ROLLBACK');
      throw error;
    }
  }
}

async function rollback(client: PoolClient) {
  const latest = await client.query<{ name: string }>('SELECT name FROM schema_migrations ORDER BY applied_at DESC, name DESC LIMIT 1');
  if (!latest.rowCount) return;
  const upFile = latest.rows[0].name;
  const downFile = upFile.replace(/\.up\.sql$/, '.down.sql');
  await client.query('BEGIN');
  try {
    await client.query(await readFile(resolve(process.cwd(), 'migrations', downFile), 'utf8'));
    await client.query('DELETE FROM schema_migrations WHERE name = $1', [upFile]);
    await client.query('COMMIT');
    process.stdout.write(`rolled back ${upFile}\n`);
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  }
}

void migrate().catch((error: unknown) => {
  process.stderr.write(`migration failed: ${error instanceof Error ? error.message : 'unknown error'}\n`);
  process.exitCode = 1;
});
