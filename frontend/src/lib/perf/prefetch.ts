import { onMount } from 'svelte';

export function prefetchOnHover(path: string) {
	let link: HTMLAnchorElement | undefined;

	function handleMouseEnter() {
		if (!link) return;
		const url = new URL(link.href, window.location.origin);
		// Trigger SvelteKit's preload_data
		const data = link.dataset;
		if (data.sveltekitPreloadData === undefined) {
			// Use the built-in preload mechanism
			link.dispatchEvent(new MouseEvent('mouseenter', { bubbles: false }));
		}
	}

	onMount(() => {
		if (link) {
			link.addEventListener('mouseenter', handleMouseEnter, { once: true, passive: true });
		}
		return () => {
			link?.removeEventListener('mouseenter', handleMouseEnter);
		};
	});

	return {
		get element() {
			return link;
		},
		set element(el: HTMLAnchorElement | undefined) {
			link = el;
		}
	};
}
