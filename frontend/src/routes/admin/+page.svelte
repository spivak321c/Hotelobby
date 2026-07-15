<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Reservation, BookingReport, OccupancyReport } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());

	let totalReservations = $state(0);
	let confirmedCount = $state(0);
	let pendingCount = $state(0);
	let cancelledCount = $state(0);
	let totalRevenue = $state(0);
	let recentReservations = $state<Reservation[]>([]);
	let bookingData = $state<BookingReport | null>(null);
	let loading = $state(true);

	function formatCurrency(n: number): string {
		return `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	function statusColor(s: string): string {
		const c: Record<string, string> = {
			confirmed: 'var(--color-sage-700, #40416C)',
			pending: 'var(--color-stone-400, #A9A296)',
			cancelled: '#9b3a30'
		};
		return c[s] || 'var(--color-stone-400, #A9A296)';
	}

	function dateRange(days: number): { from: string; to: string } {
		const to = new Date();
		const from = new Date();
		from.setDate(from.getDate() - days);
		return {
			from: from.toISOString().slice(0, 10),
			to: to.toISOString().slice(0, 10)
		};
	}

	async function load() {
		if (!token) return;
		loading = true;
		try {
			const range = dateRange(30);
			const [reservations, bookingReport] = await Promise.all([
				adminApi.listReservations(token),
				adminApi.bookingReport(token, range).catch(() => null)
			]);

			totalReservations = reservations.length;
			confirmedCount = reservations.filter((r) => r.status === 'confirmed').length;
			pendingCount = reservations.filter((r) => r.status === 'pending').length;
			cancelledCount = reservations.filter((r) => r.status === 'cancelled').length;
			totalRevenue = reservations
				.filter((r) => r.status === 'confirmed')
				.reduce((sum, r) => sum + r.total_amount, 0);

			recentReservations = reservations.slice(0, 5);
			bookingData = bookingReport ?? null;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Admin Dashboard — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Dashboard</h1>
		<AutoRefresh onRefresh={load} storageKey="dashboard" {loading} />
	</div>

	{#if loading}
		<div class="stats-grid">
			{#each [1, 2, 3, 4] as _}
				<div class="stat-card skeleton">
					<div class="skeleton-line"></div>
					<div class="skeleton-line short"></div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="stats-grid">
			<div class="stat-card">
				<span class="stat-label">Total Reservations (30d)</span>
				<span class="stat-value">{totalReservations}</span>
			</div>
			<div class="stat-card">
				<span class="stat-label">Confirmed</span>
				<span class="stat-value sage">{confirmedCount}</span>
			</div>
			<div class="stat-card">
				<span class="stat-label">Pending</span>
				<span class="stat-value">{pendingCount}</span>
			</div>
			<div class="stat-card">
				<span class="stat-label">Revenue (30d)</span>
				<span class="stat-value">{formatCurrency(totalRevenue)}</span>
			</div>
		</div>

		{#if bookingData}
			<div class="report-summary">
				<h2 class="section-title">Booking Report — {bookingData.from} to {bookingData.to}</h2>
				<div class="summary-stats">
					<div class="stat-card">
						<span class="stat-label">Total Bookings</span>
						<span class="stat-value">{bookingData.total_bookings}</span>
					</div>
					<div class="stat-card">
						<span class="stat-label">Total Revenue</span>
						<span class="stat-value">{formatCurrency(bookingData.total_revenue)}</span>
					</div>
					{#each Object.entries(bookingData.by_status) as [status, count]}
						<div class="stat-card">
							<span class="stat-label">{status}</span>
							<span class="stat-value">{count}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if recentReservations.length > 0}
			<div class="recent-section">
				<h2 class="section-title">Recent Reservations</h2>
				<div class="table-wrap">
					<table class="data-table">
						<thead>
							<tr>
								<th>Reference</th>
								<th>Guest</th>
								<th>Total</th>
								<th>Status</th>
								<th>Date</th>
							</tr>
						</thead>
						<tbody>
							{#each recentReservations as r}
								<tr>
									<td class="mono">{r.reference_code}</td>
									<td>{r.guest_name}</td>
									<td>{formatCurrency(r.total_amount)}</td>
									<td>
										<span class="status-dot" style="color: {statusColor(r.status)}">{r.status}</span>
									</td>
									<td class="muted">{new Date(r.created_at).toLocaleDateString()}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.admin-page {
		max-width: 80rem;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.5vw, 2rem);
		font-weight: 300;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-bottom: 2.5rem;
	}

	@media (min-width: 640px) {
		.stats-grid { grid-template-columns: repeat(4, 1fr); }
	}

	.stat-card {
		padding: 1.25rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
	}

	.stat-label {
		display: block;
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.5rem;
	}

	.stat-value {
		font-family: var(--font-display);
		font-size: 1.8rem;
		font-weight: 300;
		display: block;
	}

	.stat-value.sage {
		color: var(--color-sage-700, #40416C);
	}

	.section-title {
		font-family: var(--font-display);
		font-size: 1.2rem;
		font-weight: 300;
		margin-bottom: 1rem;
	}

	/* Report summary */
	.report-summary {
		margin-bottom: 2.5rem;
	}

	.summary-stats {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
	}

	@media (min-width: 640px) {
		.summary-stats { grid-template-columns: repeat(4, 1fr); }
	}

	/* Table */
	.recent-section {
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
	}

	.table-wrap {
		overflow-x: auto;
	}

	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}

	th {
		text-align: left;
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		padding: 0.6rem 0.75rem;
		border-bottom: 1px solid var(--color-stone-200, #E4E1DB);
	}

	td {
		padding: 0.6rem 0.75rem;
		border-bottom: 1px solid var(--color-stone-100, #F0EEEA);
	}

	.mono {
		font-family: 'SF Mono', 'Fira Code', monospace;
		font-size: 0.8rem;
	}

	.muted {
		color: var(--color-stone-400, #A9A296);
	}

	.status-dot {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.skeleton {
		padding: 1.25rem;
	}

	.skeleton-line {
		height: 1rem;
		background: var(--color-stone-100, #F0EEEA);
		animation: pulse 1.5s ease-in-out infinite;
		margin-bottom: 0.5rem;
	}

	.skeleton-line.short { width: 40%; margin-bottom: 0; }

	@keyframes pulse {
		0%, 100% { opacity: 0.4; }
		50% { opacity: 0.8; }
	}
</style>
