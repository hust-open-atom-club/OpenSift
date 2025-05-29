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
};

export default nextConfig;
