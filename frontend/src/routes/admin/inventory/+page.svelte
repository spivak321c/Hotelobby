<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { RoomTypeInventory, RoomType } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let inventory = $state<RoomTypeInventory[]>([]);
	let roomTypes = $state<RoomType[]>([]);
	let loading = $state(true);
	let selectedDate = $state(new Date().toISOString().slice(0, 10));
	let editing = $state<string | null>(null);
	let editForm = $state({ total_rooms: 0, booked_rooms: 0 });
	let saving = $state(false);

	function dateStr(d: string): string {
		return new Date(d).toLocaleDateString();
	}

	function typeName(id: string): string {
		return roomTypes.find((t) => t.id === id)?.name || id.slice(0, 8);
	}

	async function load() {
		if (!token) return;
		loading = true;
		try {
			const [inv, rt] = await Promise.all([
				adminApi.getInventory(token, { date: selectedDate }),
				adminApi.listRoomTypes(token).catch(() => [])
			]);
			inventory = inv;
			roomTypes = rt;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function startEdit(item: RoomTypeInventory) {
		editing = item.room_type_id;
		editForm = { total_rooms: item.total_rooms, booked_rooms: item.booked_rooms };
	}

	async function saveEdit() {
		if (!token || !editing) return;
		saving = true;
		try {
			await adminApi.updateInventory(token, {
				room_type_id: editing,
				date: selectedDate,
				total_rooms: editForm.total_rooms,
				booked_rooms: editForm.booked_rooms
			});
			editing = null;
			await load();
		} catch (e) {
			console.error(e);
		} finally {
			saving = false;
		}
	}

	onMount(load);

	$effect(() => {
		selectedDate;
		load();
	});
</script>

<svelte:head>
	<title>Inventory — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Room Inventory</h1>
		<div class="header-actions">
			<AutoRefresh onRefresh={load} storageKey="inventory" {loading} />
			<label class="date-label">
				<span>Date</span>
				<input class="date-input" type="date" bind:value={selectedDate} />
			</label>
		</div>
	</div>

	{#if loading}
		<div class="loading-msg">Loading inventory...</div>
	{:else if inventory.length === 0}
		<div class="empty">No inventory data for this date.</div>
	{:else}
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Room Type</th>
						<th>Total Rooms</th>
						<th>Booked</th>
						<th>Available</th>
						<th>Occupancy</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each inventory as item}
						{#if editing === item.room_type_id}
							<tr class="editing-row">
								<td>{typeName(item.room_type_id)}</td>
								<td>
									<input class="inline-input" type="number" bind:value={editForm.total_rooms} />
								</td>
								<td>
									<input class="inline-input" type="number" bind:value={editForm.booked_rooms} />
								</td>
								<td>{editForm.total_rooms - editForm.booked_rooms}</td>
								<td>{editForm.total_rooms > 0 ? Math.round((editForm.booked_rooms / editForm.total_rooms) * 100) : 0}%</td>
								<td class="actions">
									<button class="link-btn" onclick={saveEdit} disabled={saving}>
										{saving ? 'Saving...' : 'Save'}
									</button>
									<button class="cancel-link" onclick={() => { editing = null; }}>Cancel</button>
								</td>
							</tr>
						{:else}
							<tr>
								<td>{typeName(item.room_type_id)}</td>
								<td>{item.total_rooms}</td>
							<td>{item.booked_rooms}</td>
							<td class:sage={item.total_rooms - item.booked_rooms > 0} class:zero={item.total_rooms - item.booked_rooms === 0}>
								{item.total_rooms - item.booked_rooms}
							</td>
								<td>{item.total_rooms > 0 ? Math.round((item.booked_rooms / item.total_rooms) * 100) : 0}%</td>
								<td>
									<button class="link-btn" onclick={() => startEdit(item)}>Edit</button>
								</td>
							</tr>
						{/if}
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
		align-items: flex-end;
		margin-bottom: 2rem;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.5vw, 2rem);
		font-weight: 300;
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

	.sage { color: var(--color-sage-700, #40416C); font-weight: 500; }
	.zero { color: #9b3a30; font-weight: 500; }

	.editing-row { background: var(--color-stone-50, #F7F6F2); }

	.inline-input {
		width: 5rem;
		padding: 0.3rem 0.5rem;
		font-family: inherit;
		font-size: 0.85rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
		color: var(--color-ink, #1B1917);
	}

	.inline-input:focus { outline: none; border-color: var(--color-ink, #1B1917); }

	.actions { display: flex; gap: 0.75rem; align-items: center; }

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
	.link-btn:disabled { opacity: 0.4; cursor: not-allowed; }

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

	.empty, .loading-msg {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}
</style>
