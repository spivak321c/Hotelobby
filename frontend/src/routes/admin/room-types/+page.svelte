<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { RoomType } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let roomTypes = $state<RoomType[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let editing = $state<string | null>(null);
	let form = $state({ name: '', description: '', base_rate_daily: 0, base_rate_hourly: 0, max_occupancy: 2 });
	let saving = $state(false);

	function formatCurrency(n: number): string {
		return `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	async function load() {
		if (!token) return;
		loading = true;
		try {
			roomTypes = await adminApi.listRoomTypes(token);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		form = { name: '', description: '', base_rate_daily: 0, base_rate_hourly: 0, max_occupancy: 2 };
	}

	async function createType() {
		if (!token || !form.name) return;
		saving = true;
		try {
			await adminApi.createRoomType(token, form);
			resetForm();
			showCreate = false;
			await load();
		} catch (e) {
			console.error(e);
		} finally {
			saving = false;
		}
	}

	function startEdit(rt: RoomType) {
		editing = rt.id;
		form = { name: rt.name, description: rt.description || '', base_rate_daily: rt.base_rate_daily, base_rate_hourly: rt.base_rate_hourly || 0, max_occupancy: rt.max_occupancy || 2 };
	}

	async function saveEdit(id: string) {
		if (!token) return;
		saving = true;
		try {
			await adminApi.updateRoomType(token, id, form);
			editing = null;
			resetForm();
			await load();
		} catch (e) {
			console.error(e);
		} finally {
			saving = false;
		}
	}

	async function deleteType(id: string) {
		if (!token || !confirm('Delete this room type?')) return;
		try {
			await adminApi.deleteRoomType(token, id);
			await load();
		} catch (e) {
			console.error(e);
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Room Types — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Room Types</h1>
		<div class="header-actions">
			<AutoRefresh onRefresh={load} storageKey="room-types" {loading} />
			<button class="create-btn" onclick={() => { showCreate = !showCreate; resetForm(); }}>
			{showCreate ? 'Cancel' : '+ New Type'}
		</button>
		</div>
	</div>

	{#if showCreate}
		<div class="create-form">
			<h2 class="form-title">New Room Type</h2>
			<div class="form-grid">
				<label class="form-label full">
					<span>Name</span>
					<input class="form-input" bind:value={form.name} placeholder="e.g. Deluxe Suite" />
				</label>
				<label class="form-label full">
					<span>Description</span>
					<textarea class="form-input" rows="2" bind:value={form.description}></textarea>
				</label>
				<label class="form-label">
					<span>Daily Rate</span>
					<input class="form-input" type="number" bind:value={form.base_rate_daily} />
				</label>
				<label class="form-label">
					<span>Hourly Rate</span>
					<input class="form-input" type="number" bind:value={form.base_rate_hourly} />
				</label>
				<label class="form-label">
					<span>Max Occupancy</span>
					<input class="form-input" type="number" bind:value={form.max_occupancy} />
				</label>
			</div>
			<button class="submit-btn" onclick={createType} disabled={saving || !form.name}>
				{saving ? 'Creating...' : 'Create Type'}
			</button>
		</div>
	{/if}

	{#if loading}
		<div class="loading-msg">Loading room types...</div>
	{:else if roomTypes.length === 0}
		<div class="empty">No room types yet.</div>
	{:else}
		<div class="type-list">
			{#each roomTypes as rt}
				{#if editing === rt.id}
					<div class="type-card editing">
						<div class="form-grid">
							<label class="form-label full">
								<span>Name</span>
								<input class="form-input" bind:value={form.name} />
							</label>
							<label class="form-label full">
								<span>Description</span>
								<textarea class="form-input" rows="2" bind:value={form.description}></textarea>
							</label>
							<label class="form-label">
								<span>Daily Rate</span>
								<input class="form-input" type="number" bind:value={form.base_rate_daily} />
							</label>
							<label class="form-label">
								<span>Hourly Rate</span>
								<input class="form-input" type="number" bind:value={form.base_rate_hourly} />
							</label>
							<label class="form-label">
								<span>Max Occupancy</span>
								<input class="form-input" type="number" bind:value={form.max_occupancy} />
							</label>
						</div>
						<div class="card-actions">
							<button class="submit-btn small" onclick={() => saveEdit(rt.id)} disabled={saving}>
								{saving ? 'Saving...' : 'Save'}
							</button>
							<button class="cancel-link" onclick={() => { editing = null; resetForm(); }}>Cancel</button>
						</div>
					</div>
				{:else}
					<div class="type-card">
						<div class="type-info">
							<h3 class="type-name">{rt.name}</h3>
							{#if rt.description}
								<p class="type-desc">{rt.description}</p>
							{/if}
							<div class="type-rates">
								<span class="rate">{formatCurrency(rt.base_rate_daily)}/day</span>
								{#if rt.base_rate_hourly}
									<span class="rate muted">{formatCurrency(rt.base_rate_hourly)}/hr</span>
								{/if}
								<span class="rate muted">{rt.max_occupancy || 2} guests</span>
							</div>
						</div>
						<div class="card-actions">
							<button class="link-btn" onclick={() => startEdit(rt)}>Edit</button>
							<button class="delete-btn" onclick={() => deleteType(rt.id)}>Delete</button>
						</div>
					</div>
				{/if}
			{/each}
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

	.create-form, .type-card {
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
		margin-bottom: 1rem;
	}

	.type-card.editing { background: var(--color-stone-50, #F7F6F2); }

	.form-title {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
		margin-bottom: 1rem;
	}

	.form-grid {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr;
		gap: 1rem;
	}

	.full { grid-column: 1 / -1; }

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
		margin-top: 1rem;
		transition: opacity 0.15s;
	}

	.submit-btn.small { margin-top: 0; padding: 0.45rem 1rem; }
	.submit-btn:hover:not(:disabled) { opacity: 0.85; }
	.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }

	.type-list {
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.type-card {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1.5rem;
	}

	.type-name {
		font-family: var(--font-display);
		font-size: 1.2rem;
		font-weight: 400;
		margin-bottom: 0.3rem;
	}

	.type-desc {
		font-size: 0.85rem;
		color: var(--color-stone-500, #857E72);
		margin-bottom: 0.5rem;
	}

	.type-rates {
		display: flex;
		gap: 1.5rem;
	}

	.rate {
		font-size: 0.85rem;
		font-weight: 500;
	}

	.muted { color: var(--color-stone-400, #A9A296); font-weight: 400; }

	.card-actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
		flex-shrink: 0;
	}

	.link-btn {
		font-family: inherit;
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--color-sage-700, #40416C);
		background: none;
		border: none;
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
		transition: opacity 0.15s;
	}

	.link-btn:hover { opacity: 0.7; }

	.cancel-link {
		font-family: inherit;
		font-size: 0.75rem;
		color: var(--color-stone-400, #A9A296);
		background: none;
		border: none;
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
		transition: color 0.15s;
	}

	.cancel-link:hover { color: var(--color-ink, #1B1917); }

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

	@media (max-width: 640px) {
		.type-card { flex-direction: column; }
		.form-grid { grid-template-columns: 1fr; }
	}
</style>
