import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { routes } from '../routes.ts'

const html = await readFile(new URL('../dist/index.html', import.meta.url), 'utf8')
for (const route of routes) {
  if (route.path === '/') continue
  const directory = new URL(`../dist${route.path}/`, import.meta.url)
  await mkdir(directory, { recursive: true })
  await writeFile(new URL('index.html', directory), html.replace('<title>GMWE</title>', `<title>${route.meta.title}</title>`))
}
