<script lang="ts">
	import { onMount } from 'svelte';

	let {
		src,
		alt,
		width,
		height,
		class: className = '',
		loading = 'lazy',
		sizes,
		srcset,
		priority = false
	}: {
		src: string;
		alt: string;
		width?: number;
		height?: number;
		class?: string;
		loading?: 'lazy' | 'eager';
		sizes?: string;
		srcset?: string;
		priority?: boolean;
	} = $props();

	let containerEl: HTMLDivElement | undefined = $state(undefined);
	let loaded = $state(false);
	let error = $state(false);
	let intersects = $state(false);

	onMount(() => {
		if (loading === 'eager' || priority) {
			intersects = true;
			return;
		}

		if (!containerEl) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) {
					intersects = true;
					observer.disconnect();
				}
			},
			{ rootMargin: '200px' }
		);

		observer.observe(containerEl);

		return () => observer.disconnect();
	});

	function onLoad() {
		loaded = true;
	}

	function onError() {
		error = true;
		loaded = true;
	}
</script>

<div
	bind:this={containerEl}
	class="relative overflow-hidden {className}"
	style:width={width ? `${width}px` : undefined}
	style:height={height ? `${height}px` : undefined}
>
	{#if error}
		<div class="flex h-full w-full items-center justify-center bg-surface-alt text-text-muted">
			<svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="1.5"
					d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
				/>
			</svg>
		</div>
	{:else if !loaded}
		<div class="absolute inset-0 animate-pulse bg-surface-alt"></div>
	{/if}

	{#if intersects}
		<img
			{src}
			{alt}
			{width}
			{height}
			{sizes}
			{srcset}
			loading={priority ? 'eager' : 'lazy'}
			decoding="async"
			class="h-full w-full object-cover transition-opacity duration-300 {loaded ? 'opacity-100' : 'opacity-0'}"
			fetchpriority={priority ? 'high' : undefined}
			onload={onLoad}
			onerror={onError}
		/>
	{/if}
</div>
