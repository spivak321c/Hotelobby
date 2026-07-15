<script lang="ts">
	import { reservationsApi } from '$lib/api/client';
	import type { ReservationWithBookings } from '$lib/types/api';
	import { toast } from '$lib/stores/toast.svelte';

	let reference = $state('');
	let email = $state('');
	let reservation = $state<ReservationWithBookings | null>(null);
	let loading = $state(false);
	let error = $state('');

	// Cancel flow state
	let showCancel = $state(false);
	let cancelLoading = $state(false);
	let cancelError = $state('');
	let otpSent = $state(false);
	let otp = $state('');
	let cancelReason = $state('');

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

	function isCancellable(status: string): boolean {
		return status === 'pending' || status === 'confirmed';
	}

	async function handleLookup(e: Event) {
		e.preventDefault();
		if (!reference || !email) return;

		loading = true;
		error = '';
		reservation = null;
		resetCancel();

		try {
			reservation = await reservationsApi.lookup(reference, email);
		} catch (err: any) {
			error = err?.message || 'Reservation not found';
			toast.error('Reservation not found', 'Check your reference code and email and try again.');
		} finally {
			loading = false;
		}
	}

	function resetCancel() {
		showCancel = false;
		cancelLoading = false;
		cancelError = '';
		otpSent = false;
		otp = '';
		cancelReason = '';
	}

	async function sendOTP() {
		if (!reservation) return;
		cancelLoading = true;
		cancelError = '';
		try {
			await reservationsApi.requestCancelOTP(reservation.reference_code, email);
			otpSent = true;
		} catch (err: any) {
			cancelError = err?.message || 'Failed to send OTP';
		} finally {
			cancelLoading = false;
		}
	}

	async function confirmCancel(e: Event) {
		e.preventDefault();
		if (!reservation || !otp) return;
		cancelLoading = true;
		cancelError = '';
		try {
			const updated = await reservationsApi.cancel(reservation.reference_code, {
				otp,
				reason: cancelReason || undefined
			});
			reservation = { ...reservation, status: updated.status };
			resetCancel();
			toast.success('Reservation cancelled', 'Your booking has been successfully cancelled.');
		} catch (err: any) {
			cancelError = err?.message || 'Cancellation failed';
			toast.error('Cancellation failed', err?.message || 'Please try again.');
		} finally {
			cancelLoading = false;
		}
	}

	function formatDate(d: string): string {
		return new Date(d).toLocaleDateString('en-US', {
			month: 'short', day: 'numeric', year: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>Look Up Booking — The Lobby</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="section-tag">Manage Booking <span class="section-tag-line"></span></p>
		<h1 class="page-title">Look Up Reservation</h1>
		<p class="page-desc">Enter your booking reference and email to view your reservation details.</p>
	</header>

	<div class="lookup-layout">
		<form class="lookup-form" onsubmit={handleLookup}>
			{#if error}
				<div class="form-error">{error}</div>
			{/if}

			<div class="input-group">
				<label for="ref">Booking Reference</label>
				<input
					id="ref"
					type="text"
					bind:value={reference}
					placeholder="e.g. HB-a1b2c3d4"
					required
				/>
			</div>

			<div class="input-group">
				<label for="email">Email Address</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
				/>
			</div>

			<button type="submit" class="submit-btn" disabled={loading || !reference || !email}>
				{loading ? 'Looking up...' : 'Find Reservation'}
			</button>
		</form>

		{#if reservation}
			<div class="result-card">
				<div class="result-header">
					<div class="result-ref">
						<span class="result-ref-label">Reference</span>
						<span class="result-ref-code">{reservation.reference_code}</span>
					</div>
					<span
						class="result-status"
						style="color: {statusColor(reservation.status)}"
					>
						<span class="result-status-dot" style="background: {statusColor(reservation.status)}"></span>
						{reservation.status}
					</span>
				</div>

				<div class="result-details">
					<div class="result-row">
						<span class="result-label">Guest</span>
						<span class="result-value">{reservation.guest_name}</span>
					</div>
					<div class="result-row">
						<span class="result-label">Email</span>
						<span class="result-value">{reservation.guest_email}</span>
					</div>
					{#if reservation.guest_phone}
						<div class="result-row">
							<span class="result-label">Phone</span>
							<span class="result-value">{reservation.guest_phone}</span>
						</div>
					{/if}
					<div class="result-row">
						<span class="result-label">Total</span>
						<span class="result-value">${reservation.total_amount}</span>
					</div>
					<div class="result-row">
						<span class="result-label">Created</span>
						<span class="result-value">{formatDate(reservation.created_at)}</span>
					</div>
				</div>

				{#if reservation.bookings?.length}
					<div class="result-divider"></div>
					<h3 class="bookings-title">Bookings</h3>
					<div class="bookings-list">
						{#each reservation.bookings as booking}
							<div class="booking-item">
								<div class="booking-row">
									<span class="booking-label">Check In</span>
									<span class="booking-value">{formatDate(booking.starts_at)}</span>
								</div>
								<div class="booking-row">
									<span class="booking-label">Check Out</span>
									<span class="booking-value">{formatDate(booking.ends_at)}</span>
								</div>
								<div class="booking-row">
									<span class="booking-label">Type</span>
									<span class="booking-value">{booking.booking_type}</span>
								</div>
								<div class="booking-row">
									<span class="booking-label">Status</span>
									<span class="booking-value" style="color: {statusColor(booking.status)}">{booking.status}</span>
								</div>
							</div>
						{/each}
					</div>
				{/if}

				{#if isCancellable(reservation.status) && !showCancel}
					<div class="result-divider"></div>
					<button class="cancel-btn" onclick={() => { showCancel = true; sendOTP(); }}>
						Cancel Reservation
					</button>
				{/if}

				{#if showCancel}
					<div class="result-divider"></div>
					<div class="cancel-section">
						<h3 class="cancel-title">Cancel Reservation</h3>
						<p class="cancel-desc">
							{#if otpSent}
								A verification code has been sent to <strong>{email}</strong>. Enter it below to confirm cancellation.
							{:else}
								Sending verification code to your email...
							{/if}
						</p>

						{#if cancelError}
							<div class="form-error">{cancelError}</div>
						{/if}

						{#if otpSent}
							<form class="cancel-form" onsubmit={confirmCancel}>
								<div class="input-group">
									<label for="otp">Verification Code</label>
									<input
										id="otp"
										type="text"
										bind:value={otp}
										placeholder="Enter 6-digit code"
										maxlength="6"
										required
									/>
								</div>
								<div class="input-group">
									<label for="reason">Reason (optional)</label>
									<textarea
										id="reason"
										bind:value={cancelReason}
										placeholder="Why are you cancelling?"
										rows="2"
									></textarea>
								</div>
								<div class="cancel-actions">
									<button type="submit" class="cancel-confirm-btn" disabled={cancelLoading || !otp}>
										{cancelLoading ? 'Cancelling...' : 'Confirm Cancellation'}
									</button>
									<button type="button" class="cancel-cancel-btn" onclick={resetCancel}>
										Keep Reservation
									</button>
								</div>
							</form>
						{/if}

						{#if !otpSent && !cancelLoading}
							<button class="resend-link" onclick={sendOTP}>Resend code</button>
						{/if}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.page {
		max-width: 80rem;
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
		max-width: 32rem;
	}

	.lookup-layout {
		display: grid;
		grid-template-columns: 1fr;
		gap: 3rem;
		max-width: 40rem;
	}

	@media (min-width: 768px) {
		.lookup-layout { grid-template-columns: 1fr 1fr; max-width: none; }
	}

	.lookup-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
		align-self: start;
	}

	.form-error {
		padding: 0.8rem 1rem;
		font-size: 0.8rem;
		background: rgba(180, 60, 50, 0.08);
		color: #9b3a30;
		border: 1px solid rgba(180, 60, 50, 0.15);
	}

	.input-group {
		display: flex;
		flex-direction: column;
	}

	.input-group label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.5rem;
	}

	.input-group input,
	.input-group textarea {
		padding: 0.85rem 1rem;
		font-family: var(--font-body);
		font-size: 0.9rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: transparent;
		color: var(--color-ink, #1B1917);
		outline: none;
		transition: border-color 0.2s;
		resize: vertical;
	}

	.input-group input:focus,
	.input-group textarea:focus {
		border-color: var(--color-ink, #1B1917);
	}

	.submit-btn {
		padding: 0.9rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		border: none;
		cursor: pointer;
		transition: opacity 0.2s;
		position: relative;
		overflow: hidden;
	}

	.submit-btn::after {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 2px;
		background: var(--color-brass-400, #B8A475);
		transform: translateY(-2px);
		transition: transform 0.3s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
	}

	.submit-btn:not(:disabled):hover::after {
		transform: translateY(0);
	}

	.submit-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.submit-btn:not(:disabled):hover {
		opacity: 0.9;
	}

	/* Result */
	.result-card {
		border: 1px solid var(--color-stone-200, #E4E1DB);
		padding: 2rem;
		position: relative;
	}

	.result-card::before {
		content: '';
		position: absolute;
		top: -1px;
		left: 0;
		width: 3rem;
		height: 2px;
		background: var(--color-brass-400, #B8A475);
	}

	.result-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1.5rem;
	}

	.result-ref {
		display: flex;
		flex-direction: column;
	}

	.result-ref-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.25rem;
	}

	.result-ref-code {
		font-family: var(--font-display);
		font-size: 1.5rem;
		font-weight: 400;
	}

	.result-status {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}

	.result-status-dot {
		width: 6px;
		height: 6px;
		border-radius: 999px;
	}

	.result-details {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.result-row {
		display: flex;
		justify-content: space-between;
	}

	.result-label {
		font-size: 0.8rem;
		color: var(--color-stone-500, #857E72);
	}

	.result-value {
		font-size: 0.85rem;
		font-weight: 500;
	}

	.result-divider {
		height: 1px;
		background: var(--color-stone-200, #E4E1DB);
		margin: 1.5rem 0;
	}

	.bookings-title {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
		margin-bottom: 1rem;
	}

	.bookings-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.booking-item {
		padding: 1rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.booking-row {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.3rem;
	}

	.booking-row:last-child { margin-bottom: 0; }

	.booking-label {
		font-size: 0.75rem;
		color: var(--color-stone-500, #857E72);
	}

	.booking-value {
		font-size: 0.8rem;
		font-weight: 500;
	}

	/* Cancel section */
	.cancel-btn {
		width: 100%;
		padding: 0.8rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		background: transparent;
		color: #9b3a30;
		border: 1px solid #9b3a30;
		cursor: pointer;
		transition: background 0.2s;
	}

	.cancel-btn:hover {
		background: rgba(155, 58, 48, 0.08);
	}

	.cancel-section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.cancel-title {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
	}

	.cancel-desc {
		font-size: 0.85rem;
		line-height: 1.6;
		color: var(--color-stone-500, #857E72);
	}

	.cancel-desc strong {
		color: var(--color-ink, #1B1917);
	}

	.cancel-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.cancel-actions {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.cancel-confirm-btn {
		flex: 1;
		padding: 0.8rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		background: #9b3a30;
		color: #fff;
		border: none;
		cursor: pointer;
		transition: opacity 0.2s;
		min-width: 10rem;
	}

	.cancel-confirm-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.cancel-confirm-btn:not(:disabled):hover {
		opacity: 0.85;
	}

	.cancel-cancel-btn {
		padding: 0.8rem 1.2rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		background: transparent;
		color: var(--color-stone-500, #857E72);
		border: 1px solid var(--color-stone-200, #E4E1DB);
		cursor: pointer;
		transition: border-color 0.2s;
	}

	.cancel-cancel-btn:hover {
		border-color: var(--color-ink, #1B1917);
		color: var(--color-ink, #1B1917);
	}

	.resend-link {
		background: none;
		border: none;
		font-family: var(--font-body);
		font-size: 0.8rem;
		color: var(--color-sage-700, #40416C);
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
		padding: 0;
		align-self: flex-start;
	}

	.resend-link:hover {
		opacity: 0.7;
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

		.lookup-layout {
			gap: 2rem;
		}

		.input-group input,
		.input-group textarea {
			font-size: 0.9rem;
		}

		.submit-btn {
			width: 100%;
		}

		.result-card {
			padding: 1.25rem;
		}

		.result-ref-code {
			font-size: 1.2rem;
		}

		.result-label {
			font-size: 0.75rem;
		}

		.result-value {
			font-size: 0.8rem;
		}

		.booking-label {
			font-size: 0.7rem;
		}

		.booking-value {
			font-size: 0.75rem;
		}

		.cancel-actions {
			flex-direction: column;
		}

		.cancel-confirm-btn,
		.cancel-cancel-btn {
			width: 100%;
			min-width: 0;
		}
	}
</style>
