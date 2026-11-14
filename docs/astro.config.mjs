import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';
import tailwind from '@astrojs/tailwind';

export default defineConfig({
  site: 'https://kube-triage.dev',
  integrations: [mdx(), tailwind()],
  markdown: {
    shikiConfig: {
      themes: {
        light: 'github-light',
        dark: 'github-dark'
      },
      wrap: true
    }
  },
  compressHTML: true,
  prefetch: true,
  image: {
    service: {
      entrypoint: 'astro/assets/services/sharp'
    }
  }
});