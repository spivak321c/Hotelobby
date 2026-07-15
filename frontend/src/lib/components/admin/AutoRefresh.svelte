<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	let { onRefresh, storageKey, loading = false }: {
		onRefresh: () => void;
		storageKey: string;
		loading?: boolean;
	} = $props();

	let polling = $state(false);
	let intervalId: ReturnType<typeof setInterval> | undefined;

	onMount(() => {
		try {
			const saved = sessionStorage.getItem(`admin_poll_${storageKey}`);
			if (saved === 'true') polling = true;
		} catch {}
	});

	$effect(() => {
		if (intervalId) clearInterval(intervalId);
		intervalId = undefined;
		if (polling) {
			intervalId = setInterval(() => onRefresh(), 60000);
		}
		try {
			sessionStorage.setItem(`admin_poll_${storageKey}`, polling ? 'true' : 'false');
		} catch {}
	});

	onDestroy(() => {
		if (intervalId) clearInterval(intervalId);
	});
</script>

<div class="toolbar">
	<button class="refresh-btn" onclick={onRefresh} disabled={loading}>
		{loading ? 'Refreshing…' : 'Refresh'}
	</button>
	<label class="poll-toggle">
		<input type="checkbox" bind:checked={polling} />
		<span>Auto-refresh (60s)</span>
	</label>
</div>

<style>
	.toolbar {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.refresh-btn {
		padding: 0.45rem 0.9rem;
		font-family: inherit;
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		border: none;
		cursor: pointer;
		transition: opacity 0.15s;
	}
	.refresh-btn:hover:not(:disabled) { opacity: 0.85; }
	.refresh-btn:disabled { opacity: 0.4; cursor: not-allowed; }
	.poll-toggle {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.7rem;
		color: var(--color-stone-500, #857E72);
		cursor: pointer;
		user-select: none;
		white-space: nowrap;
	}
	.poll-toggle input { cursor: pointer; }
</style>
