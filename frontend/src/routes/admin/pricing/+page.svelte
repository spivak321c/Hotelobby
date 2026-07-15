<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { RoomPricing, RoomType } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let pricing = $state<RoomPricing[]>([]);
	let roomTypes = $state<RoomType[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let form = $state({ room_type_id: '', rate_type: 'daily' as 'daily' | 'hourly', rate: 0, effective_from: '', effective_to: '' });
	let saving = $state(false);

	function formatCurrency(n: number): string {
		return `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	function typeName(id: string): string {
		return roomTypes.find((t) => t.id === id)?.name || id.slice(0, 8);
	}

	function dateStr(d: string): string {
		return new Date(d).toLocaleDateString();
	}

	async function load() {
		if (!token) return;
		loading = true;
		try {
			const [p, rt] = await Promise.all([
				adminApi.listPricing(token),
				adminApi.listRoomTypes(token).catch(() => [])
			]);
			pricing = p;
			roomTypes = rt;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function createPricing() {
		if (!token || !form.room_type_id || !form.effective_from || !form.effective_to) return;
		saving = true;
		try {
			await adminApi.createPricing(token, form);
			showCreate = false;
			form = { room_type_id: '', rate_type: 'daily', rate: 0, effective_from: '', effective_to: '' };
			await load();
		} catch (e) {
			console.error(e);
		} finally {
			saving = false;
		}
	}

	async function deletePricing(id: string) {
		if (!token || !confirm('Delete this pricing rule?')) return;
		try {
			await adminApi.deletePricing(token, id);
			await load();
		} catch (e) {
			console.error(e);
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Pricing — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Pricing Rules</h1>
		<div class="header-actions">
			<AutoRefresh onRefresh={load} storageKey="pricing" {loading} />
			<button class="create-btn" onclick={() => { showCreate = !showCreate; }}>
			{showCreate ? 'Cancel' : '+ New Rule'}
		</button>
		</div>
	</div>

	{#if showCreate}
		<div class="create-form">
			<h2 class="form-title">New Pricing Rule</h2>
			<div class="form-row">
				<label class="form-label">
					<span>Room Type</span>
					<select class="form-input" bind:value={form.room_type_id}>
						<option value="">Select...</option>
						{#each roomTypes as rt}
							<option value={rt.id}>{rt.name}</option>
						{/each}
					</select>
				</label>
				<label class="form-label">
					<span>Rate Type</span>
					<select class="form-input" bind:value={form.rate_type}>
						<option value="daily">Daily</option>
						<option value="hourly">Hourly</option>
					</select>
				</label>
				<label class="form-label">
					<span>Rate ($)</span>
					<input class="form-input" type="number" bind:value={form.rate} />
				</label>
				<label class="form-label">
					<span>Effective From</span>
					<input class="form-input" type="date" bind:value={form.effective_from} />
				</label>
				<label class="form-label">
					<span>Effective To</span>
					<input class="form-input" type="date" bind:value={form.effective_to} />
				</label>
			</div>
			<button class="submit-btn" onclick={createPricing} disabled={saving}>
				{saving ? 'Creating...' : 'Create Rule'}
			</button>
		</div>
	{/if}

	{#if loading}
		<div class="loading-msg">Loading pricing...</div>
	{:else if pricing.length === 0}
		<div class="empty">No pricing rules yet.</div>
	{:else}
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Room Type</th>
						<th>Rate Type</th>
						<th>Rate</th>
						<th>From</th>
						<th>To</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each pricing as p}
						<tr>
							<td>{typeName(p.room_type_id)}</td>
							<td class="muted">{p.rate_type}</td>
							<td>{formatCurrency(p.rate)}</td>
							<td class="muted">{dateStr(p.effective_range.lower)}</td>
							<td class="muted">{dateStr(p.effective_range.upper)}</td>
							<td>
								<button class="delete-btn" onclick={() => deletePricing(p.id)}>Delete</button>
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
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.5vw, 2rem);
		font-weight: 300;
	}

	.create-btn {
		padding: 0.5rem 1rem;
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

	.create-btn:hover { opacity: 0.85; }

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.create-form {
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
		margin-bottom: 1.5rem;
	}

	.form-title {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
		margin-bottom: 1rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.form-label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.form-label span {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	.form-input {
		padding: 0.6rem 0.75rem;
		font-family: inherit;
		font-size: 0.9rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: var(--color-cream, #FAFAF5);
		color: var(--color-ink, #1B1917);
	}

	.form-input:focus { outline: none; border-color: var(--color-ink, #1B1917); }

	.submit-btn {
		padding: 0.6rem 1.5rem;
		font-family: inherit;
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		border: none;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.submit-btn:hover:not(:disabled) { opacity: 0.85; }
	.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }

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

	.delete-btn {
		font-family: inherit;
		font-size: 0.7rem;
		color: #9b3a30;
		background: none;
		border: none;
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
		transition: opacity 0.15s;
	}

	.delete-btn:hover { opacity: 0.6; }

	.empty, .loading-msg {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}
</style>
