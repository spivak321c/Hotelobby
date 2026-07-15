<script lang="ts">
	import { onMount } from 'svelte';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Admin } from '$lib/types/api';
	import AutoRefresh from '$lib/components/admin/AutoRefresh.svelte';

	const token = $derived(auth.getToken());
	let admins = $state<Admin[]>([]);
	let loading = $state(true);
	let showCreate = $state(false);
	let form = $state({ full_name: '', email: '', password: '', role: 'front_desk' as string });
	let saving = $state(false);

	const roles = ['super_admin', 'manager', 'front_desk'];

	function roleLabel(r: string): string {
		return r.split('_').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
	}

	async function load() {
		if (!token) return;
		loading = true;
		try {
			admins = await adminApi.listAdmins(token);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	}

	async function createAdmin() {
		if (!token || !form.email || !form.password || !form.full_name) return;
		saving = true;
		try {
			await adminApi.createAdmin(token, form as any);
			showCreate = false;
			form = { full_name: '', email: '', password: '', role: 'front_desk' };
			await load();
		} catch (e) {
			console.error(e);
		} finally {
			saving = false;
		}
	}

	async function deleteAdmin(id: string) {
		if (!token || !confirm('Delete this admin?')) return;
		try {
			await adminApi.deleteAdmin(token, id);
			await load();
		} catch (e) {
			console.error(e);
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Admins — Admin — The Lobby</title>
</svelte:head>

<div class="admin-page">
	<div class="page-header">
		<h1 class="page-title">Admin Accounts</h1>
		<div class="header-actions">
			<AutoRefresh onRefresh={load} storageKey="admins" {loading} />
			<button class="create-btn" onclick={() => { showCreate = !showCreate; }}>
			{showCreate ? 'Cancel' : '+ New Admin'}
		</button>
		</div>
	</div>

	{#if showCreate}
		<div class="create-form">
			<h2 class="form-title">New Admin</h2>
			<div class="form-row">
				<label class="form-label">
					<span>Full Name</span>
					<input class="form-input" bind:value={form.full_name} />
				</label>
				<label class="form-label">
					<span>Email</span>
					<input class="form-input" type="email" bind:value={form.email} />
				</label>
				<label class="form-label">
					<span>Password</span>
					<input class="form-input" type="password" bind:value={form.password} />
				</label>
				<label class="form-label">
					<span>Role</span>
					<select class="form-input" bind:value={form.role}>
						{#each roles as r}
							<option value={r}>{roleLabel(r)}</option>
						{/each}
					</select>
				</label>
			</div>
			<button class="submit-btn" onclick={createAdmin} disabled={saving || !form.email || !form.password || !form.full_name}>
				{saving ? 'Creating...' : 'Create Admin'}
			</button>
		</div>
	{/if}

	{#if loading}
		<div class="loading-msg">Loading admins...</div>
	{:else if admins.length === 0}
		<div class="empty">No admin accounts.</div>
	{:else}
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Name</th>
						<th>Email</th>
						<th>Role</th>
						<th>Status</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each admins as a}
						<tr>
							<td>{a.full_name}</td>
							<td class="muted">{a.email}</td>
							<td>{roleLabel(a.role)}</td>
							<td>
								<span class="status-badge" class:active={a.is_active} class:inactive={!a.is_active}>
									{a.is_active ? 'Active' : 'Inactive'}
								</span>
							</td>
							<td>
								<button class="delete-btn" onclick={() => deleteAdmin(a.id)}>Delete</button>
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

	.status-badge {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		padding: 0.2rem 0.5rem;
		border: 1px solid;
	}

	.status-badge.active {
		color: var(--color-sage-700, #40416C);
		border-color: var(--color-sage-700, #40416C);
	}

	.status-badge.inactive {
		color: var(--color-stone-400, #A9A296);
		border-color: var(--color-stone-400, #A9A296);
	}

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
