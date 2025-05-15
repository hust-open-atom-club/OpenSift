import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  /* config options here */
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: (process.env.BACKEND || 'http://localhost:5000') + '/api/:path*',
      }
    ]
  },
  env: {
    NEXT_PUBLIC_BACKEND: (process.env.BACKEND || 'http://localhost:5000')
  }
};

export default nextConfig;
