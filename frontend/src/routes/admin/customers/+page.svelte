<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Customer } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let customers = $state<Customer[]>([]);
	let loading = $state(true);

	async function load() {
		if (!token) return;
		loading = true;
		try {
			customers = await adminApi.listCustomers(token);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Customers — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Customers</h1>
		<AutoRefresh onRefresh={load} storageKey="customers" {loading} />
	</div>

	{#if loading}
		<div class="loading-msg">Loading customers...</div>
	{:else if customers.length === 0}
		<div class="empty">No customers yet.</div>
	{:else}
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Name</th>
						<th>Email</th>
						<th>Phone</th>
						<th>Joined</th>
					</tr>
				</thead>
				<tbody>
					{#each customers as c}
						<tr>
							<td>{c.full_name}</td>
							<td class="muted">{c.email}</td>
							<td class="muted">{c.phone || '—'}</td>
							<td class="muted">{new Date(c.created_at).toLocaleDateString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
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
		margin-bottom: 2rem;
		flex-wrap: wrap;
		gap: 1rem;
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

	.muted { color: var(--color-stone-400, #A9A296); }

	.empty, .loading-msg {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}
</style>
