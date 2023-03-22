/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  redirects: async () => {
    return [
      {
        source: '/fe-api/:path*',
        destination: '/api/:path*',
        permanent: true,
      }
    ]
  }
}

module.exports = nextConfig
