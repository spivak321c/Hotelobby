<script lang="ts">
	import Skeleton from './Skeleton.svelte';

	let { count = 3 }: { count?: number } = $props();
</script>

<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
	{#each { length: count } as _}
		<div class="overflow-hidden rounded-xl border border-border">
			<Skeleton variant="rectangular" height="12rem" />
			<div class="space-y-3 p-4">
				<Skeleton variant="text" width="60%" />
				<Skeleton variant="text" width="40%" />
				<div class="flex gap-2 pt-2">
					<Skeleton variant="text" width="5rem" height="2rem" />
					<Skeleton variant="text" width="5rem" height="2rem" />
				</div>
			</div>
		</div>
	{/each}
</div>
