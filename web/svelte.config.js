import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
      pages: process.env.BUILD_DIR || "build",
      assets: process.env.BUILD_DIR || "build",
      fallback: "fallback.html",
      precompress: false,
      strict: true,
    }),
	}
};

export default config;
