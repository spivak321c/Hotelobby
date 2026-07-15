<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { customerApi, reservationsApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { ReservationWithBookings } from '$lib/types/api';

	let reservation = $state<ReservationWithBookings | null>(null);
	let loading = $state(true);
	let cancelling = $state(false);
	let cancelReason = $state('');
	let showCancelForm = $state(false);
	let otpSent = $state(false);
	let otp = $state('');
	let otpSending = $state(false);
	let otpError = $state('');

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

	function formatDateTime(d: string): string {
		return new Date(d).toLocaleDateString('en-US', {
			month: 'short', day: 'numeric', year: 'numeric',
			hour: 'numeric', minute: '2-digit'
		});
	}

	async function requestOTP() {
		if (!reservation || !reservation.guest_email) return;
		otpSending = true;
		otpError = '';
		try {
			await reservationsApi.requestCancelOTP(reservation.reference_code, reservation.guest_email);
			otpSent = true;
		} catch (e) {
			otpError = e instanceof Error ? e.message : 'Failed to send OTP';
		} finally {
			otpSending = false;
		}
	}

	async function confirmCancel() {
		if (!reservation) return;
		cancelling = true;
		otpError = '';
		try {
			await reservationsApi.cancel(reservation.reference_code, {
				otp,
				reason: cancelReason || undefined
			});
			reservation = { ...reservation, status: 'cancelled' };
			showCancelForm = false;
			otpSent = false;
			otp = '';
		} catch (e) {
			otpError = e instanceof Error ? e.message : 'Failed to cancel reservation';
		} finally {
			cancelling = false;
		}
	}

	onMount(async () => {
		if (!token) return;
		const id = page.params.id ?? '';
		try {
			reservation = await customerApi.getReservation(token, id) as any;
		} catch (e) {
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Reservation {reservation?.reference_code || ''} — The Lobby</title>
</svelte:head>

<div class="page">
	<nav class="breadcrumb">
		<a href="/dashboard">Dashboard</a>
		<span class="breadcrumb-sep">/</span>
		<span>Reservation</span>
	</nav>

	{#if loading}
		<div class="skeleton">
			<div class="skeleton-line"></div>
			<div class="skeleton-line short"></div>
			<div class="skeleton-line"></div>
		</div>
	{:else if !reservation}
		<div class="empty">Reservation not found.</div>
	{:else}
		<div class="header-row">
			<div>
				<p class="section-tag">Reference</p>
				<h1 class="ref-code">{reservation.reference_code}</h1>
			</div>
			<span class="status" style="color: {statusColor(reservation.status)}">
				{reservation.status}
			</span>
		</div>

		<div class="info-grid">
			<div class="info-block">
				<span class="info-label">Guest</span>
				<span class="info-value">{reservation.guest_name}</span>
			</div>
			<div class="info-block">
				<span class="info-label">Email</span>
				<span class="info-value">{reservation.guest_email}</span>
			</div>
			<div class="info-block">
				<span class="info-label">Phone</span>
				<span class="info-value">{reservation.guest_phone}</span>
			</div>
			<div class="info-block">
				<span class="info-label">Total</span>
				<span class="info-value total">${reservation.total_amount}</span>
			</div>
			<div class="info-block">
				<span class="info-label">Payment</span>
				<span class="info-value">{reservation.payment?.provider ?? '—'}</span>
			</div>
			<div class="info-block">
				<span class="info-label">Created</span>
				<span class="info-value">{formatDateTime(reservation.created_at)}</span>
			</div>
		</div>

		{#if reservation.bookings?.length}
			<div class="bookings-section">
				<h2 class="section-title">Bookings</h2>
				<div class="bookings-list">
					{#each reservation.bookings as booking}
						<div class="booking-card">
							<div class="booking-header">
								<span class="booking-type">{booking.booking_type}</span>
								<span class="booking-amount">${booking.amount}</span>
							</div>
							<div class="booking-dates">
								<div class="booking-date">
									<span class="date-label">Starts</span>
									<span class="date-value">{formatDate(booking.starts_at)}</span>
								</div>
								<div class="booking-date">
									<span class="date-label">Ends</span>
									<span class="date-value">{formatDate(booking.ends_at)}</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if reservation.status === 'confirmed' || reservation.status === 'pending'}
			<div class="cancel-section">
				{#if showCancelForm}
					<div class="cancel-form">
						<h3 class="cancel-title">Cancel Reservation</h3>
						<p class="cancel-desc">A one-time passcode will be sent to {reservation.guest_email}</p>
						<p class="cancel-desc">Reason (optional)</p>
						<textarea
							class="cancel-input"
							rows="3"
							placeholder="Why are you cancelling?"
							bind:value={cancelReason}
						></textarea>
						{#if otpError}
							<p class="cancel-error">{otpError}</p>
						{/if}
						{#if otpSent}
							<p class="cancel-desc">Enter the OTP sent to your email</p>
							<input
								class="cancel-input"
								type="text"
								placeholder="Enter OTP"
								bind:value={otp}
							/>
							<div class="cancel-actions">
								<button
									class="cancel-btn secondary"
									onclick={() => { showCancelForm = false; cancelReason = ''; otpSent = false; otp = ''; otpError = ''; }}
								>
									Keep Reservation
								</button>
								<button
									class="cancel-btn primary"
									onclick={confirmCancel}
									disabled={cancelling || !otp}
								>
									{cancelling ? 'Cancelling...' : 'Confirm Cancellation'}
								</button>
							</div>
						{:else}
							<div class="cancel-actions">
								<button
									class="cancel-btn secondary"
									onclick={() => { showCancelForm = false; cancelReason = ''; }}
								>
									Keep Reservation
								</button>
								<button
									class="cancel-btn primary"
									onclick={requestOTP}
									disabled={otpSending}
								>
									{otpSending ? 'Sending...' : 'Request Cancel OTP'}
								</button>
							</div>
						{/if}
					</div>
				{:else}
					<button class="cancel-trigger" onclick={() => { showCancelForm = true; }}>
						Cancel Reservation
					</button>
				{/if}
			</div>
		{/if}
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
		transition: color 0.2s;
	}

	.breadcrumb a:hover { color: var(--color-ink, #1B1917); }

	.breadcrumb-sep {
		color: var(--color-stone-300, #CCC9C1);
	}

	.header-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2.5rem;
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.section-tag {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.15em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.5rem;
	}

	.ref-code {
		font-family: var(--font-display);
		font-size: clamp(1.8rem, 3vw, 2.5rem);
		font-weight: 300;
	}

	.status {
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		padding: 0.4rem 1rem;
		border: 1px solid currentColor;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1.5rem;
		margin-bottom: 2.5rem;
	}

	@media (min-width: 640px) {
		.info-grid { grid-template-columns: repeat(3, 1fr); }
	}

	.info-block {
		display: flex;
		flex-direction: column;
	}

	.info-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.35rem;
	}

	.info-value {
		font-size: 0.95rem;
		font-weight: 500;
	}

	.total {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 300;
	}

	.section-title {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 300;
		margin-bottom: 1rem;
	}

	.bookings-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-bottom: 2.5rem;
	}

	.booking-card {
		padding: 1.25rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.booking-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.75rem;
	}

	.booking-type {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	.booking-amount {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
	}

	.booking-dates {
		display: flex;
		gap: 2rem;
	}

	.booking-date {
		display: flex;
		flex-direction: column;
	}

	.date-label {
		font-size: 0.65rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.2rem;
	}

	.date-value {
		font-size: 0.9rem;
	}

	.cancel-section {
		padding-top: 2rem;
		border-top: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.cancel-trigger {
		font-size: 0.75rem;
		font-weight: 500;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		color: #9b3a30;
		background: none;
		border: 1px solid #9b3a30;
		padding: 0.7rem 1.5rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.cancel-trigger:hover {
		background: #9b3a30;
		color: #fff;
	}

	.cancel-form {
		padding: 1.5rem;
		border: 1px solid #9b3a30;
		background: rgba(155, 58, 48, 0.03);
	}

	.cancel-title {
		font-family: var(--font-display);
		font-size: 1.2rem;
		font-weight: 400;
		margin-bottom: 0.75rem;
		color: #9b3a30;
	}

	.cancel-desc {
		font-size: 0.75rem;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.5rem;
	}

	.cancel-input {
		width: 100%;
		padding: 0.75rem;
		font-family: inherit;
		font-size: 0.9rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: var(--color-cream, #FAFAF5);
		color: var(--color-ink, #1B1917);
		resize: vertical;
	}

	.cancel-input:focus {
		outline: none;
		border-color: var(--color-stone-400, #A9A296);
	}

	.cancel-actions {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
	}

	.cancel-btn {
		padding: 0.7rem 1.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		cursor: pointer;
		transition: all 0.2s;
	}

	.cancel-btn.secondary {
		background: none;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		color: var(--color-stone-500, #857E72);
	}

	.cancel-btn.secondary:hover {
		border-color: var(--color-ink, #1B1917);
		color: var(--color-ink, #1B1917);
	}

	.cancel-btn.primary {
		background: #9b3a30;
		border: 1px solid #9b3a30;
		color: #fff;
	}

	.cancel-btn.primary:hover { opacity: 0.85; }
	.cancel-btn.primary:disabled { opacity: 0.5; cursor: not-allowed; }

	.cancel-error {
		font-size: 0.75rem;
		color: #9b3a30;
		margin-bottom: 0.75rem;
	}

	.empty {
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}

	.skeleton {
		padding: 2rem 0;
	}

	.skeleton-line {
		height: 1.2rem;
		background: var(--color-stone-100, #F0EEEA);
		animation: pulse 1.5s ease-in-out infinite;
		margin-bottom: 1rem;
	}

	.skeleton-line.short { width: 30%; margin-bottom: 0; }

	@keyframes pulse {
		0%, 100% { opacity: 0.4; }
		50% { opacity: 0.8; }
	}
</style>
