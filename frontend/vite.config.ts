import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	server: {
		port: 5174,
		host: '0.0.0.0'
	},
	plugins: [tailwindcss(), sveltekit()]
});
