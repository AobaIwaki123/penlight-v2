import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  outputFileTracingRoot: __dirname,
  images: {
    unoptimized: true,
  },
  devIndicators: false,
  reactStrictMode: true,
};

export default nextConfig;
