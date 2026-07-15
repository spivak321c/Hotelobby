<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { adminApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { ReservationWithBookings } from '$lib/types/api';

	const token = $derived(auth.getToken());
	let reservation = $state<ReservationWithBookings | null>(null);
	let loading = $state(true);
	let statusLoading = $state(false);

	// Reason prompt state
	let pendingStatus = $state<string | null>(null);
	let reasonInput = $state('');

	const statuses = ['pending', 'confirmed', 'checked_in', 'checked_out', 'cancelled', 'refunded'] as const;
	const needsReason = (s: string) => s === 'cancelled' || s === 'refunded';

	function statusColor(s: string): string {
		const c: Record<string, string> = {
			confirmed: 'var(--color-sage-700, #40416C)',
			pending: 'var(--color-stone-400, #A9A296)',
			cancelled: '#9b3a30',
			refunded: 'var(--color-stone-400, #A9A296)',
			checked_in: 'var(--color-sage-600, #4A5D42)',
			checked_out: 'var(--color-stone-400, #A9A296)',
			completed: 'var(--color-stone-400, #A9A296)'
		};
		return c[s] || 'var(--color-stone-400, #A9A296)';
	}

	function formatCurrency(n: number): string {
		return `$${n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	function formatDate(d: string): string {
		return new Date(d).toLocaleDateString('en-US', {
			month: 'short', day: 'numeric', year: 'numeric'
		});
	}

	async function updateStatus(newStatus: string) {
		if (needsReason(newStatus)) {
			pendingStatus = newStatus;
			reasonInput = '';
			return;
		}
		await doUpdateStatus(newStatus);
	}

	async function doUpdateStatus(newStatus: string, reason?: string) {
		if (!token || !reservation) return;
		statusLoading = true;
		try {
			const updated = await adminApi.updateReservationStatus(token, reservation.id, {
				status: newStatus as any,
				reason: reason || undefined
			});
			reservation = { ...reservation, status: updated.status };
		} catch (e) {
			console.error(e);
		} finally {
			statusLoading = false;
		}
	}

	function confirmReasonPrompt() {
		if (!pendingStatus) return;
		const status = pendingStatus;
		pendingStatus = null;
		doUpdateStatus(status, reasonInput);
	}

	function cancelReasonPrompt() {
		pendingStatus = null;
		reasonInput = '';
	}

	onMount(async () => {
		if (!token) return;
		const id = page.params.id ?? '';
		try {
			reservation = await adminApi.getReservation(token, id);
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Reservation {reservation?.reference_code || ''} — Admin</title>
</svelte:head>

<div class="admin-page">
	<nav class="breadcrumb">
		<a href="/admin/reservations">Reservations</a>
		<span>/</span>
		<span>{reservation?.reference_code || '...'}</span>
	</nav>

	{#if loading}
		<div class="loading-msg">Loading reservation...</div>
	{:else if !reservation}
		<div class="empty">Reservation not found.</div>
	{:else}
		<div class="header-row">
			<div>
				<p class="section-tag">Reference</p>
				<h1 class="ref-code">{reservation.reference_code}</h1>
			</div>
			<div class="status-actions">
				<span class="status-badge" style="color: {statusColor(reservation.status)}; border-color: currentColor">
					{reservation.status}
				</span>
			</div>
		</div>

		<div class="content-grid">
			<!-- Info -->
			<div class="card">
				<h2 class="card-title">Guest Information</h2>
				<div class="info-rows">
					<div class="info-row">
						<span class="info-label">Name</span>
						<span class="info-value">{reservation.guest_name}</span>
					</div>
					<div class="info-row">
						<span class="info-label">Email</span>
						<span class="info-value">{reservation.guest_email}</span>
					</div>
					<div class="info-row">
						<span class="info-label">Phone</span>
						<span class="info-value">{reservation.guest_phone}</span>
					</div>
					<div class="info-row">
						<span class="info-label">Total</span>
						<span class="info-value total">{formatCurrency(reservation.total_amount)}</span>
					</div>
					<div class="info-row">
						<span class="info-label">Created</span>
						<span class="info-value">{new Date(reservation.created_at).toLocaleString()}</span>
					</div>
				</div>
			</div>

		<!-- Status Management -->
		<div class="card">
			<h2 class="card-title">Update Status</h2>
			<div class="status-buttons">
				{#each statuses as s}
					<button
						class="status-btn"
						class:current={reservation.status === s}
						disabled={statusLoading || reservation.status === s}
						onclick={() => updateStatus(s)}
						style="--btn-color: {statusColor(s)}"
					>
						{s.replace('_', ' ')}
					</button>
				{/each}
			</div>

			{#if pendingStatus}
				<div class="reason-prompt">
					<p class="reason-prompt-label">
						Reason for <strong>{pendingStatus.replace('_', ' ')}</strong> (optional):
					</p>
					<textarea
						class="reason-input"
						bind:value={reasonInput}
						placeholder="e.g. Guest requested cancellation due to change of plans"
						rows="2"
					></textarea>
					<div class="reason-actions">
						<button
							class="submit-btn small"
							onclick={confirmReasonPrompt}
							disabled={statusLoading}
						>
							{statusLoading ? 'Updating...' : 'Confirm'}
						</button>
						<button class="cancel-link" onclick={cancelReasonPrompt}>Cancel</button>
					</div>
				</div>
			{/if}
		</div>
		</div>

		<!-- Bookings -->
		{#if reservation.bookings?.length}
			<div class="card full-width">
				<h2 class="card-title">Bookings ({reservation.bookings.length})</h2>
				<div class="bookings-list">
					{#each reservation.bookings as booking}
						<div class="booking-row">
							<div class="booking-main">
								<span class="booking-type">{booking.booking_type}</span>
								<span class="booking-id mono">{booking.room_id}</span>
							</div>
							<div class="booking-dates">
								<span>{formatDate(booking.starts_at)}</span>
								<span class="muted">to</span>
								<span>{formatDate(booking.ends_at)}</span>
							</div>
							<span class="booking-amount">{formatCurrency(booking.amount)}</span>
							<span class="booking-status" style="color: {statusColor(booking.status)}">{booking.status}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.admin-page { max-width: 80rem; }

	.breadcrumb {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.75rem;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 2rem;
	}

	.breadcrumb a {
		color: var(--color-stone-400, #A9A296);
		text-decoration: none;
		transition: color 0.15s;
	}

	.breadcrumb a:hover { color: var(--color-ink, #1B1917); }

	.header-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.section-tag {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.15em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.3rem;
	}

	.ref-code {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.5vw, 2rem);
		font-weight: 300;
	}

	.status-badge {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		border: 1px solid;
		padding: 0.3rem 0.8rem;
	}

	.content-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	@media (min-width: 640px) {
		.content-grid { grid-template-columns: 1fr 1fr; }
	}

	.card {
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: #fff;
	}

	.card.full-width { margin-bottom: 2rem; }

	.card-title {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
		margin-bottom: 1.25rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--color-stone-100, #F0EEEA);
	}

	.info-rows {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.info-row {
		display: flex;
		justify-content: space-between;
	}

	.info-label {
		font-size: 0.75rem;
		color: var(--color-stone-400, #A9A296);
	}

	.info-value {
		font-size: 0.9rem;
		font-weight: 500;
		text-align: right;
	}

	.total {
		font-family: var(--font-display);
		font-size: 1.2rem;
		font-weight: 300;
	}

	.status-buttons {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.status-btn {
		padding: 0.45rem 0.9rem;
		font-family: inherit;
		font-size: 0.7rem;
		font-weight: 500;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		background: none;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		color: var(--color-stone-500, #857E72);
		cursor: pointer;
		transition: all 0.15s;
	}

	.status-btn:hover:not(:disabled) {
		border-color: var(--btn-color);
		color: var(--btn-color);
	}

	.status-btn.current {
		border-color: var(--btn-color);
		color: var(--btn-color);
		background: color-mix(in srgb, var(--btn-color) 5%, transparent);
	}

	.status-btn:disabled { opacity: 0.4; cursor: not-allowed; }

	.reason-prompt {
		margin-top: 1.25rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--color-stone-100, #F0EEEA);
	}

	.reason-prompt-label {
		font-size: 0.8rem;
		color: var(--color-stone-500, #857E72);
		margin-bottom: 0.6rem;
	}

	.reason-prompt-label strong {
		color: var(--color-ink, #1B1917);
	}

	.reason-input {
		width: 100%;
		padding: 0.6rem 0.75rem;
		font-family: inherit;
		font-size: 0.85rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: var(--color-cream, #FAFAF5);
		color: var(--color-ink, #1B1917);
		resize: vertical;
	}

	.reason-input:focus { outline: none; border-color: var(--color-ink, #1B1917); }

	.reason-actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
		margin-top: 0.75rem;
	}

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

	.submit-btn.small { padding: 0.45rem 1rem; }
	.submit-btn:hover:not(:disabled) { opacity: 0.85; }
	.submit-btn:disabled { opacity: 0.4; cursor: not-allowed; }

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

	.bookings-list {
		display: flex;
		flex-direction: column;
	}

	.booking-row {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 0.75rem 0;
		border-bottom: 1px solid var(--color-stone-100, #F0EEEA);
		flex-wrap: wrap;
	}

	.booking-row:last-child { border-bottom: none; }

	.booking-main {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		min-width: 10rem;
	}

	.booking-type {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	.booking-dates {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		font-size: 0.85rem;
		flex: 1;
	}

	.booking-amount {
		font-family: var(--font-display);
		font-size: 1rem;
		min-width: 5rem;
		text-align: right;
	}

	.booking-status {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		min-width: 5rem;
		text-align: right;
	}

	.mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.8rem; }
	.muted { color: var(--color-stone-400, #A9A296); }

	.empty, .loading-msg {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}
</style>
