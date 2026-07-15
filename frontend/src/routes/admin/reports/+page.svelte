<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { BookingReport, OccupancyReport, RevenueReport } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let from = $state(new Date(Date.now() - 30 * 86400000).toISOString().slice(0, 10));
	let to = $state(new Date().toISOString().slice(0, 10));
	let activeTab = $state<'bookings' | 'occupancy' | 'revenue'>('bookings');
	let loading = $state(false);

	let bookings = $state<BookingReport | null>(null);
	let occupancy = $state<OccupancyReport | null>(null);
	let revenue = $state<RevenueReport | null>(null);

	function formatCurrency(n: number): string {
		return `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	async function load() {
		if (!token) return;
		loading = true;
		try {
			const [b, o] = await Promise.all([
				adminApi.bookingReport(token, { from, to }),
				adminApi.occupancyReport(token, { from, to })
			]);
			bookings = b;
			occupancy = o;
			try {
				revenue = await adminApi.revenueReport(token, { from, to });
			} catch {
				revenue = null;
			}
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Reports — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Reports</h1>
		<AutoRefresh onRefresh={load} storageKey="reports" {loading} />
	</div>

	<div class="controls">
		<label class="date-label">
			<span>From</span>
			<input class="date-input" type="date" bind:value={from} />
		</label>
		<label class="date-label">
			<span>To</span>
			<input class="date-input" type="date" bind:value={to} />
		</label>
		<button class="load-btn" onclick={load} disabled={loading}>
			{loading ? 'Loading...' : 'Load Reports'}
		</button>
	</div>

	<div class="tabs">
		<button class="tab" class:active={activeTab === 'bookings'} onclick={() => { activeTab = 'bookings'; }}>Bookings</button>
		<button class="tab" class:active={activeTab === 'occupancy'} onclick={() => { activeTab = 'occupancy'; }}>Occupancy</button>
		<button class="tab" class:active={activeTab === 'revenue'} onclick={() => { activeTab = 'revenue'; }}>Revenue</button>
	</div>

	{#if loading}
		<div class="loading-msg">Loading...</div>
	{:else}
		{#if activeTab === 'bookings' && bookings}
			<div class="report-cards">
				<div class="summary-card">
					<span class="summary-label">Period</span>
					<span class="summary-sub">{bookings.from} — {bookings.to}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Total Bookings</span>
					<span class="summary-value">{bookings.total_bookings}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Total Revenue</span>
					<span class="summary-value">{formatCurrency(bookings.total_revenue)}</span>
				</div>
				{#each Object.entries(bookings.by_status) as [status, count]}
					<div class="summary-card">
						<span class="summary-label">{status}</span>
						<span class="summary-value">{count}</span>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'occupancy' && occupancy}
			<div class="report-cards">
				<div class="summary-card">
					<span class="summary-label">Period</span>
					<span class="summary-sub">{occupancy.from} — {occupancy.to}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Total Rooms</span>
					<span class="summary-value">{occupancy.total_rooms}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Occupied Rooms</span>
					<span class="summary-value">{occupancy.occupied_rooms}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Occupancy Rate</span>
					<span class="summary-value">{occupancy.occupancy_rate.toFixed(1)}%</span>
				</div>
			</div>
		{:else if activeTab === 'revenue' && revenue}
			<div class="report-cards">
				<div class="summary-card">
					<span class="summary-label">Period</span>
					<span class="summary-sub">{revenue.from} — {revenue.to}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Total Revenue</span>
					<span class="summary-value">{formatCurrency(revenue.total_revenue)}</span>
				</div>
				<div class="summary-card">
					<span class="summary-label">Cancelled Revenue</span>
					<span class="summary-value">{formatCurrency(revenue.cancelled_revenue)}</span>
				</div>
				{#each Object.entries(revenue.by_status) as [status, amount]}
					<div class="summary-card">
						<span class="summary-label">{status}</span>
						<span class="summary-value">{formatCurrency(amount)}</span>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty">No data for this period.</div>
		{/if}
	{/if}
</div>

<style>
	.admin-page { max-width: 80rem; }

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.5vw, 2rem);
		font-weight: 300;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.controls {
		display: flex;
		gap: 1rem;
		align-items: flex-end;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
	}

	.date-label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.date-label span {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	.date-input {
		padding: 0.5rem 0.75rem;
		font-family: inherit;
		font-size: 0.9rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
		color: var(--color-ink, #1B1917);
	}

	.date-input:focus { outline: none; border-color: var(--color-ink, #1B1917); }

	.load-btn {
		padding: 0.5rem 1.25rem;
		font-family: inherit;
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.03em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		border: none;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.load-btn:hover:not(:disabled) { opacity: 0.85; }
	.load-btn:disabled { opacity: 0.4; cursor: not-allowed; }

	.tabs {
		display: flex;
		gap: 0;
		border-bottom: 1px solid var(--color-stone-200, #E4E1DB);
		margin-bottom: 1.5rem;
	}

	.tab {
		padding: 0.6rem 1.25rem;
		font-family: inherit;
		font-size: 0.8rem;
		font-weight: 400;
		letter-spacing: 0.03em;
		text-transform: uppercase;
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		color: var(--color-stone-400, #A9A296);
		cursor: pointer;
		transition: all 0.15s;
	}

	.tab:hover { color: var(--color-ink, #1B1917); }
	.tab.active { color: var(--color-ink, #1B1917); border-bottom-color: var(--color-ink, #1B1917); }

	.report-cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
		gap: 1rem;
	}

	.summary-card {
		display: flex;
		flex-direction: column;
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
	}

	.summary-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.5rem;
	}

	.summary-sub {
		font-size: 0.9rem;
		color: var(--color-ink, #1B1917);
		margin-bottom: 0.5rem;
	}

	.summary-value {
		font-family: var(--font-display);
		font-size: 1.8rem;
		font-weight: 300;
	}

	.empty, .loading-msg {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}
</style>
