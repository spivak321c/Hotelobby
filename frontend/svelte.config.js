import adapter from '@sveltejs/adapter-static';

const config = {
	kit: {
		adapter: adapter({
			fallback: '200.html'
		}),
		prerender: {
			handleUnseenRoutes: 'ignore'
		}
	},
	compilerOptions: {
		runes: true
	}
};

export default config;
