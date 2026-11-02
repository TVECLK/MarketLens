import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',

  outputFileTracingRoot: import.meta.dirname,
  turbopack: {
    root: import.meta.dirname,
  },
};

export default nextConfig;
