import adapter from '@sveltejs/adapter-static';
import preprocess from 'svelte-preprocess';

const dev = process.env.NODE_ENV === 'development';
/** @type {import('@sveltejs/kit').Config} */
//const config = {
//  // Consult https://github.com/sveltejs/svelte-preprocess
//  // for more information about preprocessors
//  preprocess: preprocess(),
//
//  kit: {
//    adapter: adapter({
//      // default options are shown
//      pages: 'build',
//      assets: 'build',
//      fallback: null,
//      precompress: false
//    }),
//
//    prerender: {
//      // This can be false if you're using a fallback (i.e. SPA mode)
//      default: true
//    }
//  }
//};

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: preprocess(),
  kit: {
    adapter: adapter({
      // default options are shown
      pages: 'build',
      assets: 'build',
      fallback: null,
      precompress: false
    }),

    prerender: {
      // This can be false if you're using a fallback (i.e. SPA mode)
      default: true
    },
    paths: {
      base: dev ? '' : '/ZePedroResende.github.io'
    },
    // If you are not using a .nojekyll file, change your appDir to something not starting with an underscore.
    // For example, instead of '_app', use 'app_', 'internal', etc.
    appDir: 'internal'
  }
};

export default config;
