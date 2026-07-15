<script lang="ts">
	import Skeleton from './Skeleton.svelte';

	let { rows = 5, cols = 4 }: { rows?: number; cols?: number } = $props();
</script>

<div class="overflow-hidden rounded-xl border border-border">
	<table class="w-full">
		<thead>
			<tr class="border-b border-border bg-surface-alt">
				{#each { length: cols } as _}
					<th class="px-4 py-3">
						<Skeleton variant="text" width="60%" height="1rem" />
					</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each { length: rows } as _}
				<tr class="border-b border-border last:border-0">
					{#each { length: cols } as _}
						<td class="px-4 py-3">
							<Skeleton variant="text" width="80%" />
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>
