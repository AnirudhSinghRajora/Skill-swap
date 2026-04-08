import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
	output: 'standalone',
	images: {
		remotePatterns: [
			{
				protocol: 'https',
				hostname: 'media.gettyimages.com',
				pathname: '/id/**'
			}
		]
	}
};

export default nextConfig;
