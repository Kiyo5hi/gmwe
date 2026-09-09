import { fileURLToPath } from 'node:url'
import sharp from 'sharp'

for (const size of [180, 192, 512]) {
  await sharp(fileURLToPath(new URL('../public/gmwe.webp', import.meta.url)))
    .resize(size, size, { fit: 'contain', background: '#f5f7f6' })
    .flatten({ background: '#f5f7f6' })
    .png()
    .toFile(fileURLToPath(new URL(`../public/pwa-${size}.png`, import.meta.url)))
}
