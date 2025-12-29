/** @type {import('next').NextConfig} */
const nextConfig = {
  // API backend URL
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },
  // Enable React strict mode for better error handling
  reactStrictMode: true,
}

module.exports = nextConfig
