<script lang="ts">
	import { onMount, tick } from 'svelte';

	let {
		items,
		height = 400,
		itemHeight = 60,
		overscan = 5,
		getKey,
		children
	}: {
		items: unknown[];
		height?: number;
		itemHeight?: number;
		overscan?: number;
		getKey?: (item: unknown, index: number) => string | number;
		children: any;
	} = $props();

	let container: HTMLDivElement | undefined = $state(undefined);
	let scrollTop = $state(0);

	let startIndex = $derived(Math.max(0, Math.floor(scrollTop / itemHeight) - overscan));
	let endIndex = $derived(Math.min(items.length, Math.ceil((scrollTop + height) / itemHeight) + overscan));
	let visibleItems = $derived(items.slice(startIndex, endIndex));
	let totalHeight = $derived(items.length * itemHeight);
	let offsetY = $derived(startIndex * itemHeight);

	function onScroll() {
		if (container) {
			scrollTop = container.scrollTop;
		}
	}

	async function scrollToIndex(index: number) {
		if (!container) return;
		await tick();
		container.scrollTop = index * itemHeight;
	}

	onMount(() => {
		if (container) {
			container.addEventListener('scroll', onScroll, { passive: true });
		}
		return () => {
			container?.removeEventListener('scroll', onScroll);
		};
	});
</script>

<div
	bind:this={container}
	class="overflow-auto"
	style:height="{height}px"
	role="list"
>
	<div style:height="{totalHeight}px" class="relative">
		<div style:transform="translateY({offsetY}px)">
			{#each visibleItems as item, i (getKey ? getKey(item, startIndex + i) : startIndex + i)}
				<div style:height="{itemHeight}px" role="listitem">
					{@render children(item, startIndex + i)}
				</div>
			{/each}
		</div>
	</div>
</div>
