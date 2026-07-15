<script lang="ts">
	import { onMount } from 'svelte';
	import { customerApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Reservation } from '$lib/types/api';

	let reservations = $state<Reservation[]>([]);
	let loading = $state(true);

	const token = $derived(auth.getToken());

	function statusColor(status: string): string {
		const colors: Record<string, string> = {
			confirmed: 'var(--color-sage-700, #40416C)',
			pending: 'var(--color-stone-400, #A9A296)',
			cancelled: '#9b3a30',
			refunded: 'var(--color-stone-400, #A9A296)',
			checked_in: 'var(--color-sage-600, #4A5D42)',
			completed: 'var(--color-stone-400, #A9A296)'
		};
		return colors[status] || 'var(--color-stone-400, #A9A296)';
	}

	function formatDate(d: string): string {
		return new Date(d).toLocaleDateString('en-US', {
			month: 'short', day: 'numeric', year: 'numeric'
		});
	}

	onMount(() => {
		if (!token) return;
		customerApi.listReservations(token)
			.then((data) => { reservations = data; })
			.catch(() => {})
			.finally(() => { loading = false; });
	});
</script>

<svelte:head>
	<title>Dashboard — The Lobby</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="section-tag">My Account <span class="section-tag-line"></span></p>
		<h1 class="page-title">Dashboard</h1>
		<p class="page-desc">View and manage your reservations.</p>
	</header>

	{#if loading}
		<div class="loading-list">
			{#each [1, 2, 3] as _}
				<div class="skeleton-card">
					<div class="skeleton-line"></div>
					<div class="skeleton-line short"></div>
				</div>
			{/each}
		</div>
	{:else if reservations.length === 0}
		<div class="empty">
			<div class="empty-icon">
				<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="0.75">
					<path d="M3 3v18h18"/>
					<path d="M7 16l4-8 4 4 4-6" stroke-linecap="round"/>
				</svg>
			</div>
			<p class="empty-title">No reservations yet</p>
			<p class="empty-desc">When you book a stay, your reservation history will appear here.</p>
			<a href="/rooms" class="empty-link">Browse Rooms</a>
		</div>
	{:else}
		<div class="reservations-list">
			{#each reservations as r}
				<a href="/dashboard/reservations/{r.id}" class="reservation-card">
					<div class="res-header">
						<div class="res-ref">
							<span class="res-ref-label">Reference</span>
							<span class="res-ref-code">{r.reference_code}</span>
						</div>
						<span class="res-status" style="color: {statusColor(r.status)}">
							<span class="res-status-dot" style="background: {statusColor(r.status)}"></span>
							{r.status}
						</span>
					</div>
					<div class="res-details">
						<div class="res-row">
							<span class="res-label">Guest</span>
							<span class="res-value">{r.guest_name}</span>
						</div>
						<div class="res-row">
							<span class="res-label">Total</span>
							<span class="res-value">${r.total_amount}</span>
						</div>
						<div class="res-row">
							<span class="res-label">Created</span>
							<span class="res-value">{formatDate(r.created_at)}</span>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>

<style>
	.page {
		max-width: 56rem;
		margin: 0 auto;
		padding: 6rem 1.5rem 4rem;
	}

	@media (min-width: 640px) {
		.page { padding: 6rem 3rem 4rem; }
	}

	.page-header { margin-bottom: 3rem; }

	.section-tag {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 1rem;
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.section-tag-line {
		display: inline-block;
		width: 2rem;
		height: 1px;
		background: var(--color-brass-400, #B8A475);
		opacity: 0.6;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 3rem);
		font-weight: 300;
		line-height: 1.1;
		margin-bottom: 1rem;
	}

	.page-desc {
		font-size: 1rem;
		line-height: 1.7;
		color: var(--color-stone-500, #857E72);
	}

	.reservations-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.reservation-card {
		display: block;
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		text-decoration: none;
		color: inherit;
		transition: border-color 0.2s;
		position: relative;
	}

	.reservation-card::before {
		content: '';
		position: absolute;
		top: -1px;
		left: 0;
		width: 0;
		height: 2px;
		background: var(--color-brass-400, #B8A475);
		transition: width 0.3s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
	}

	.reservation-card:hover {
		border-color: var(--color-ink, #1B1917);
	}

	.reservation-card:hover::before {
		width: 3rem;
	}

	.res-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
	}

	.res-ref {
		display: flex;
		flex-direction: column;
	}

	.res-ref-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.25rem;
	}

	.res-ref-code {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
	}

	.res-status {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}

	.res-status-dot {
		width: 6px;
		height: 6px;
		border-radius: 999px;
	}

	.res-details {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.res-row {
		display: flex;
		justify-content: space-between;
	}

	.res-label {
		font-size: 0.8rem;
		color: var(--color-stone-500, #857E72);
	}

	.res-value {
		font-size: 0.85rem;
		font-weight: 500;
	}

	.empty {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
	}

	.empty-icon {
		color: var(--color-brass-400, #B8A475);
		opacity: 0.4;
		margin-bottom: 1rem;
	}

	.empty-title {
		font-family: var(--font-display);
		font-size: 1.5rem;
		font-weight: 300;
		color: var(--color-ink, #1B1917);
	}

	.empty-desc {
		font-size: 0.85rem;
		line-height: 1.6;
		color: var(--color-stone-500, #857E72);
		max-width: 24rem;
		margin-bottom: 1.5rem;
	}

	.empty-link {
		display: inline-block;
		margin-top: 1rem;
		padding: 0.7rem 1.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		text-decoration: none;
		transition: opacity 0.2s;
	}

	.empty-link:hover { opacity: 0.85; }

	.loading-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.skeleton-card {
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.skeleton-line {
		height: 1rem;
		background: var(--color-stone-100, #F0EEEA);
		animation: pulse 1.5s ease-in-out infinite;
		margin-bottom: 0.75rem;
	}

	.skeleton-line.short {
		width: 40%;
		margin-bottom: 0;
	}

	@keyframes pulse {
		0%, 100% { opacity: 0.4; }
		50% { opacity: 0.8; }
	}

	@media (max-width: 639px) {
		.page {
			padding: 5rem 1.25rem 3rem;
		}

		.page-header {
			margin-bottom: 2rem;
		}

		.page-title {
			font-size: clamp(1.6rem, 5vw, 2.2rem);
			margin-bottom: 0.75rem;
		}

		.page-desc {
			font-size: 0.9rem;
			line-height: 1.75;
		}

		.reservation-card {
			padding: 1.25rem;
		}

		.res-ref-code {
			font-size: 1.1rem;
		}

		.res-details {
			gap: 0.5rem;
		}

		.res-label {
			font-size: 0.75rem;
		}

		.res-value {
			font-size: 0.8rem;
		}
	}
</style>
