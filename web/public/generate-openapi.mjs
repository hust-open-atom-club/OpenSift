import { createClient } from '@hey-api/openapi-ts';
import { writeFile } from "node:fs/promises";
import path from 'node:path';

const dir = path.dirname(new URL(import.meta.url).pathname);

const r = await fetch('http://localhost:5000/swagger/doc.json')
const content = await r.text();
await writeFile(`${dir}/src/service/openapi.json`, content);

createClient({
  input: 'src/service/openapi.json',
  output: 'src/service/client',
  plugins: [
    {
      name: '@hey-api/client-fetch',
      runtimeConfigPath: "./src/service/hey-api.ts"
    }],
})