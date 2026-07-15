<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Room, RoomType } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let rooms = $state<Room[]>([]);
	let roomTypes = $state<RoomType[]>([]);
	let loading = $state(true);
	let filterType = $state('');
	let filterStatus = $state('');
	let showCreate = $state(false);
	let newRoom = $state({ room_number: '', room_type_id: '' });
	let creating = $state(false);

	async function load() {
		if (!token) return;
		loading = true;
		try {
			const params: { room_type_id?: string; status?: string } = {};
			if (filterType) params.room_type_id = filterType;
			if (filterStatus) params.status = filterStatus;
			const [r, rt] = await Promise.all([
				adminApi.listRooms(token, params),
				adminApi.listRoomTypes(token).catch(() => [])
			]);
			rooms = r;
			roomTypes = rt;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function statusColor(s: string): string {
		const c: Record<string, string> = {
			available: 'var(--color-sage-700, #40416C)',
			occupied: 'var(--color-stone-500, #857E72)',
			maintenance: '#9b3a30',
			blocked: 'var(--color-stone-400, #A9A296)',
			cleaning: 'var(--color-sage-600, #4A5D42)'
		};
		return c[s] || 'var(--color-stone-400, #A9A296)';
	}

	function typeName(id: string): string {
		return roomTypes.find((t) => t.id === id)?.name || id.slice(0, 8);
	}

	async function createRoom() {
		if (!token || !newRoom.room_number || !newRoom.room_type_id) return;
		creating = true;
		try {
			await adminApi.createRoom(token, newRoom);
			newRoom = { room_number: '', room_type_id: '' };
			showCreate = false;
			await load();
		} catch (e) {
			console.error(e);
		} finally {
			creating = false;
		}
	}

	async function deleteRoom(id: string) {
		if (!token || !confirm('Delete this room?')) return;
		try {
			await adminApi.deleteRoom(token, id);
			await load();
		} catch (e) {
			console.error(e);
		}
	}

	onMount(load);

	$effect(() => {
		filterType;
		filterStatus;
		load();
	});
</script>

<svelte:head>
	<title>Rooms — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Rooms</h1>
		<div class="header-actions">
			<AutoRefresh onRefresh={load} storageKey="rooms" {loading} />
			<select class="filter-select" bind:value={filterType}>
				<option value="">All Types</option>
				{#each roomTypes as rt}
					<option value={rt.id}>{rt.name}</option>
				{/each}
			</select>
			<select class="filter-select" bind:value={filterStatus}>
				<option value="">All Statuses</option>
				<option value="available">Available</option>
				<option value="occupied">Occupied</option>
				<option value="maintenance">Maintenance</option>
				<option value="blocked">Blocked</option>
				<option value="cleaning">Cleaning</option>
			</select>
			<button class="create-btn" onclick={() => { showCreate = !showCreate; }}>
				{showCreate ? 'Cancel' : '+ New Room'}
			</button>
		</div>
	</div>

	{#if showCreate}
		<div class="create-form">
			<h2 class="form-title">New Room</h2>
			<div class="form-row">
				<label class="form-label">
					<span>Room Number</span>
					<input class="form-input" bind:value={newRoom.room_number} placeholder="e.g. 101" />
				</label>
				<label class="form-label">
					<span>Room Type</span>
					<select class="form-input" bind:value={newRoom.room_type_id}>
						<option value="">Select type...</option>
						{#each roomTypes as rt}
							<option value={rt.id}>{rt.name}</option>
						{/each}
					</select>
				</label>
			</div>
			<button
				class="submit-btn"
				onclick={createRoom}
				disabled={creating || !newRoom.room_number || !newRoom.room_type_id}
			>
				{creating ? 'Creating...' : 'Create Room'}
			</button>
		</div>
	{/if}

	{#if loading}
		<p class="loading-msg">Loading rooms...</p>
	{:else if rooms.length === 0}
		<div class="empty">No rooms found.</div>
	{:else}
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Number</th>
						<th>Type</th>
						<th>Status</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each rooms as room}
						<tr>
							<td class="mono">{room.room_number}</td>
							<td>{typeName(room.room_type_id)}</td>
							<td>
								<span class="status-badge" style="color: {statusColor(room.status)}; border-color: currentColor">
									{room.status}
								</span>
							</td>
							<td class="actions">
								<a href="/admin/rooms/{room.id}" class="link-btn">Edit</a>
								<button class="delete-btn" onclick={() => deleteRoom(room.id)}>Delete</button>
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

	.header-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		flex-wrap: wrap;
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
		grid-template-columns: 1fr 1fr;
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

	.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; }

	.status-badge {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		border: 1px solid;
		padding: 0.2rem 0.5rem;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	.link-btn {
		font-size: 0.75rem;
		font-weight: 500;
		text-decoration: none;
		color: var(--color-sage-700, #40416C);
		transition: opacity 0.15s;
	}

	.link-btn:hover { opacity: 0.7; }

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
