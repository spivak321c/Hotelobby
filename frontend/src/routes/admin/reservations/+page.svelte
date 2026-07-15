<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Reservation } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let reservations = $state<Reservation[]>([]);
	let loading = $state(true);
	let filterStatus = $state('');

	async function load() {
		if (!token) return;
		loading = true;
		try {
			const params: { status?: string } = {};
			if (filterStatus) params.status = filterStatus;
			reservations = await adminApi.listReservations(token, params);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function statusColor(s: string): string {
		const c: Record<string, string> = {
			confirmed: 'var(--color-sage-700, #40416C)',
			pending: 'var(--color-stone-400, #A9A296)',
			cancelled: '#9b3a30',
			refunded: 'var(--color-stone-400, #A9A296)',
			checked_in: 'var(--color-sage-600, #4A5D42)',
			completed: 'var(--color-stone-400, #A9A296)'
		};
		return c[s] || 'var(--color-stone-400, #A9A296)';
	}

	function formatCurrency(n: number): string {
		return `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	onMount(load);

	$effect(() => {
		filterStatus;
		load();
	});
</script>

<svelte:head>
	<title>Reservations — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Reservations</h1>
		<div class="header-actions">
			<AutoRefresh onRefresh={load} storageKey="reservations" {loading} />
			<select class="filter-select" bind:value={filterStatus}>
				<option value="">All Statuses</option>
				<option value="pending">Pending</option>
				<option value="confirmed">Confirmed</option>
				<option value="checked_in">Checked In</option>
				<option value="completed">Completed</option>
				<option value="cancelled">Cancelled</option>
				<option value="refunded">Refunded</option>
			</select>
		</div>
	</div>

	{#if loading}
		<div class="loading-msg">Loading...</div>
	{:else if reservations.length === 0}
		<div class="empty">No reservations found.</div>
	{:else}
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Reference</th>
						<th>Guest</th>
						<th>Email</th>
						<th>Total</th>
						<th>Status</th>
						<th>Created</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each reservations as r}
						<tr>
							<td class="mono">{r.reference_code}</td>
							<td>{r.guest_name}</td>
							<td class="muted">{r.guest_email}</td>
							<td>{formatCurrency(r.total_amount)}</td>
							<td>
								<span class="status-badge" style="color: {statusColor(r.status)}; border-color: currentColor">
									{r.status}
								</span>
							</td>
							<td class="muted">{new Date(r.created_at).toLocaleDateString()}</td>
							<td>
								<a href="/admin/reservations/{r.id}" class="link-btn">View</a>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<style>
	.admin-page { max-width: 80rem; }

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.5vw, 2rem);
		font-weight: 300;
	}

	.filter-select {
		padding: 0.5rem 0.75rem;
		font-family: inherit;
		font-size: 0.8rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
		color: var(--color-ink, #1B1917);
		cursor: pointer;
	}

	.filter-select:focus { outline: none; border-color: var(--color-ink, #1B1917); }

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.table-wrap { overflow-x: auto; }

	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
		background: #fff;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	th {
		text-align: left;
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		padding: 0.75rem;
		border-bottom: 1px solid var(--color-stone-200, #E4E1DB);
		background: var(--color-stone-50, #F7F6F2);
	}

	td {
		padding: 0.6rem 0.75rem;
		border-bottom: 1px solid var(--color-stone-100, #F0EEEA);
	}

	.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; }
	.muted { color: var(--color-stone-400, #A9A296); }

	.status-badge {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		border: 1px solid;
		padding: 0.2rem 0.5rem;
	}

	.link-btn {
		font-size: 0.75rem;
		font-weight: 500;
		text-decoration: none;
		color: var(--color-sage-700, #40416C);
		transition: opacity 0.15s;
	}

	.link-btn:hover { opacity: 0.7; }

	.empty {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}

	.loading-msg {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}
</style>
