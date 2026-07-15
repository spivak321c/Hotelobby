<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { reservationsApi, paymentsApi, roomsApi, roomTypesApi } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import type { RoomType, PaymentProvider } from '$lib/types/api';

	const roomId = $derived(page.url.searchParams.get('room_id') ?? '');
	const checkIn = $derived(page.url.searchParams.get('check_in') ?? '');
	const checkOut = $derived(page.url.searchParams.get('check_out') ?? '');
	const bookingType = $derived(page.url.searchParams.get('type') ?? 'daily');

	// Step 1: Guest info
	let guestName = $state('');
	let guestEmail = $state('');
	let guestPhone = $state('');
	let roomType = $state<RoomType | null>(null);
	let submitting = $state(false);
	let error = $state('');

	let touched = $state<Record<string, boolean>>({});

	let emailError = $derived(touched.email && guestEmail.length > 0 && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(guestEmail) ? 'Please enter a valid email address' : '');
	let nameError = $derived(touched.name && guestName.length > 0 && guestName.trim().length < 2 ? 'Name must be at least 2 characters' : '');

	function touch(field: string) {
		touched[field] = true;
	}

	function formatCardNumber(val: string): string {
		const digits = val.replace(/\D/g, '').slice(0, 16);
		return digits.replace(/(\d{4})(?=\d)/g, '$1 ');
	}

	function detectCardType(num: string): string {
		const cleaned = num.replace(/\s/g, '');
		if (/^4/.test(cleaned)) return 'Visa';
		if (/^5[1-5]/.test(cleaned)) return 'Mastercard';
		if (/^3[47]/.test(cleaned)) return 'Amex';
		return '';
	}

	function formatCardInput(e: Event) {
		const input = e.target as HTMLInputElement;
		const val = input.value.replace(/\D/g, '').slice(0, 4);
		if (val.length >= 2) {
			cardExpiry = val.slice(0, 2) + '/' + val.slice(2);
		} else {
			cardExpiry = val;
		}
	}

	// Step 2: Payment
	let showPayment = $state(false);
	let reservationId = $state('');
	let reservationRef = $state('');
	let paymentMethod = $state<PaymentProvider>('paystack');
	let processingPayment = $state(false);
	let paymentError = $state('');

	// Card details
	let cardNumber = $state('');
	let cardExpiry = $state('');
	let cardCvv = $state('');
	let cardName = $state('');

	function generateUUIDv7(): string {
		const timestamp = Date.now().toString(16).padStart(12, '0');
		const random = crypto.randomUUID().replace(/-/g, '').slice(0, 12);
		return `${timestamp}-${random.slice(0, 4)}-${random.slice(4, 8)}-${random.slice(8, 12)}`;
	}

	function daysBetween(a: string, b: string): number {
		const d1 = new Date(a);
		const d2 = new Date(b);
		return Math.ceil((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24));
	}

	function hoursBetween(a: string, b: string): number {
		const d1 = new Date(a);
		const d2 = new Date(b);
		return Math.ceil((d2.getTime() - d1.getTime()) / (1000 * 60 * 60));
	}

	let quantity = $derived(
		bookingType === 'daily'
			? daysBetween(checkIn, checkOut)
			: hoursBetween(checkIn, checkOut)
	);

	let totalPrice = $derived(
		roomType
			? bookingType === 'daily'
				? quantity * roomType.base_rate_daily
				: quantity * roomType.base_rate_hourly
			: 0
	);

	function isFormValid(): boolean {
		return guestName.trim().length > 0
			&& guestEmail.trim().length > 0
			&& !!roomId
			&& !!checkIn
			&& !!checkOut;
	}

	function parseCardExpiry(val: string): { month: string; year: string } {
		const cleaned = val.replace(/\D/g, '');
		if (cleaned.length >= 2) {
			return { month: cleaned.slice(0, 2), year: cleaned.slice(2, 4) };
		}
		return { month: cleaned, year: '' };
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!isFormValid()) return;

		submitting = true;
		error = '';

		try {
			const idempotencyKey = generateUUIDv7();
			const reservation = await reservationsApi.create({
				guest_name: guestName,
				guest_email: guestEmail,
				guest_phone: guestPhone || undefined,
				customer_id: auth.isAuthenticated() ? (auth.getCustomerId() ?? undefined) : undefined,
				bookings: [{
					room_id: roomId,
					check_in: checkIn,
					check_out: checkOut,
					booking_type: bookingType as 'daily' | 'hourly',
					expected_occupants: 1
				}],
				payment_method: paymentMethod,
				idempotency_key: idempotencyKey
			}, idempotencyKey);

			reservationId = reservation.id;
			reservationRef = reservation.reference_code;
			showPayment = true;
		} catch (err: any) {
			error = err?.message || 'Failed to create reservation';
			toast.error('Reservation failed', err?.message || 'Something went wrong. Please try again.');
		} finally {
			submitting = false;
		}
	}

	async function handlePayWithCard(e: Event) {
		e.preventDefault();
		if (!reservationId) return;

		processingPayment = true;
		paymentError = '';
		const { month, year } = parseCardExpiry(cardExpiry);

		try {
			await paymentsApi.process({
				reservation_id: reservationId,
				method: 'paystack',
				card_details: {
					number: cardNumber.replace(/\s/g, ''),
					cvv: cardCvv,
					expiry_month: month,
					expiry_year: year
				}
			});
			goto(`/booking/confirmation?ref=${reservationRef}&email=${guestEmail}`);
		} catch (err: any) {
			paymentError = err?.message || 'Payment failed';
			toast.error('Payment failed', err?.message || 'Your card could not be charged. Please try again.');
		} finally {
			processingPayment = false;
		}
	}

	async function handlePayWithCrypto() {
		if (!reservationId) return;
		processingPayment = true;
		paymentError = '';

		try {
			await paymentsApi.process({
				reservation_id: reservationId,
				method: 'crossmint'
			});
			goto(`/booking/confirmation?ref=${reservationRef}&email=${guestEmail}`);
		} catch (err: any) {
			paymentError = err?.message || 'Payment failed';
			toast.error('Payment failed', err?.message || 'Crypto payment could not be processed.');
		} finally {
			processingPayment = false;
		}
	}

	onMount(() => {
		if (roomId) {
			roomsApi.getWithImages(roomId).then((room) => {
				roomTypesApi.get(room.room.room_type_id).then((rt) => {
					roomType = rt;
				}).catch(() => {});
			}).catch(() => {});
		}
	});
</script>

<svelte:head>
	<title>Book — The Lobby</title>
</svelte:head>

<div class="page">
	{#if !showPayment}
		<a href="/rooms/{roomId}" class="back-link">Back to Room</a>
	{/if}

	<!-- Step indicator -->
	<div class="step-indicator">
		<div class="step" class:active={!showPayment} class:completed={showPayment}>
			<span class="step-num">{showPayment ? '✓' : '1'}</span>
			<span class="step-label">Info</span>
		</div>
		<div class="step-line"></div>
		<div class="step" class:active={showPayment}>
			<span class="step-num">2</span>
			<span class="step-label">Payment</span>
		</div>
		<div class="step-line"></div>
		<div class="step">
			<span class="step-num">3</span>
			<span class="step-label">Confirmation</span>
		</div>
	</div>

	<div class="booking-layout">
		<!-- Step 1: Guest Info -->
		{#if !showPayment}
			<div class="booking-form-wrap">
				<h1 class="page-title">Complete Your Booking</h1>

				<form class="booking-form" onsubmit={handleSubmit}>
					{#if error}
						<div class="form-error">{error}</div>
					{/if}

					<h2 class="form-section-title">Guest Information</h2>

					<div class="input-group">
						<label for="name">Full Name</label>
						<input
							id="name"
							type="text"
							bind:value={guestName}
							placeholder="John Doe"
							required
							onblur={() => touch('name')}
							class:input-error={nameError}
						/>
						{#if nameError}
							<span class="field-hint error">{nameError}</span>
						{/if}
					</div>

					<div class="input-row">
						<div class="input-group">
							<label for="email">Email</label>
							<input
								id="email"
								type="email"
								bind:value={guestEmail}
								placeholder="you@example.com"
								required
								onblur={() => touch('email')}
								class:input-error={emailError}
							/>
							{#if emailError}
								<span class="field-hint error">{emailError}</span>
							{/if}
						</div>
						<div class="input-group">
							<label for="phone">Phone</label>
							<input
								id="phone"
								type="tel"
								bind:value={guestPhone}
								placeholder="+1 (555) 000-0000"
							/>
						</div>
					</div>

					<h2 class="form-section-title">Payment Method</h2>
					<div class="payment-toggle">
						<button
							type="button"
							class="toggle-btn"
							class:active={paymentMethod === 'paystack'}
							onclick={() => { paymentMethod = 'paystack'; }}
						>
							Card (Paystack)
						</button>
						<button
							type="button"
							class="toggle-btn"
							class:active={paymentMethod === 'crossmint'}
							onclick={() => { paymentMethod = 'crossmint'; }}
						>
							Crypto (Crossmint)
						</button>
					</div>

					<button type="submit" class="submit-btn" disabled={submitting || !isFormValid()}>
						{submitting ? 'Creating reservation...' : 'Proceed to Payment'}
					</button>

					<p class="form-note">
						By confirming, you agree to our booking terms. A confirmation email will be sent to your address.
					</p>
				</form>
			</div>
		{/if}

		<!-- Step 2: Payment -->
		{#if showPayment}
			<div class="payment-form-wrap">
				<div class="back-indicator">
					Reservation <strong>{reservationRef}</strong> created
				</div>
				<h1 class="page-title">Complete Payment</h1>
				<p class="payment-desc">
					{paymentMethod === 'paystack' ? 'Enter your card details to pay securely.' : 'Pay with cryptocurrency via Crossmint.'}
				</p>

				{#if paymentError}
					<div class="form-error">{paymentError}</div>
				{/if}

				{#if paymentMethod === 'paystack'}
					<form class="payment-form" onsubmit={handlePayWithCard}>
						<div class="input-group">
							<label for="cardName">Cardholder Name</label>
							<input
								id="cardName"
								type="text"
								bind:value={cardName}
								placeholder="John Doe"
								required
							/>
						</div>
						<div class="input-group">
							<label for="cardNumber">Card Number</label>
							<input
								id="cardNumber"
								type="text"
								bind:value={cardNumber}
								placeholder="4242 4242 4242 4242"
								maxlength="19"
								required
							/>
						</div>
						<div class="input-row">
							<div class="input-group">
								<label for="cardExpiry">Expiry</label>
								<input
									id="cardExpiry"
									type="text"
									value={cardExpiry}
									oninput={formatCardInput}
									placeholder="MM/YY"
									maxlength="5"
									required
								/>
							</div>
							<div class="input-group">
								<label for="cardCvv">CVV</label>
								<input
									id="cardCvv"
									type="text"
									bind:value={cardCvv}
									placeholder="123"
									maxlength="4"
									required
								/>
							</div>
						</div>
						<button type="submit" class="submit-btn" disabled={processingPayment}>
							{processingPayment ? 'Processing...' : `Pay $${totalPrice} with Card`}
						</button>
					</form>
				{:else}
					<div class="crypto-section">
						<p class="crypto-info">
							You'll be redirected to Crossmint to complete your crypto payment.
						</p>
						<button class="submit-btn" onclick={handlePayWithCrypto} disabled={processingPayment}>
							{processingPayment ? 'Processing...' : `Pay $${totalPrice} with Crypto`}
						</button>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Summary -->
		<div class="booking-summary">
			<h2 class="summary-title">Booking Summary</h2>

			<div class="summary-details">
				<div class="summary-row">
					<span class="summary-label">Check In</span>
					<span class="summary-value">{checkIn || '—'}</span>
				</div>
				<div class="summary-row">
					<span class="summary-label">Check Out</span>
					<span class="summary-value">{checkOut || '—'}</span>
				</div>
				<div class="summary-row">
					<span class="summary-label">Type</span>
					<span class="summary-value">{bookingType}</span>
				</div>
				<div class="summary-row">
					<span class="summary-label">Room Type</span>
					<span class="summary-value">{roomType?.name ?? '—'}</span>
				</div>
			</div>

			<div class="summary-divider"></div>

			<div class="summary-pricing">
				<div class="summary-row">
					<span class="summary-label">
						{quantity} × {bookingType === 'daily' ? `$${roomType?.base_rate_daily ?? 0}/day` : `$${roomType?.base_rate_hourly ?? 0}/hr`}
					</span>
					<span class="summary-value">${totalPrice}</span>
				</div>
			</div>

			<div class="summary-total">
				<span>Total</span>
				<span>${totalPrice}</span>
			</div>

			{#if showPayment}
				<div class="summary-ref">
					<span class="summary-ref-label">Reference</span>
					<span class="summary-ref-code">{reservationRef}</span>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.page {
		max-width: 80rem;
		margin: 0 auto;
		padding: 5rem 1.5rem 4rem;
	}

	@media (min-width: 640px) {
		.page { padding: 5rem 3rem 4rem; }
	}

	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--color-stone-500, #857E72);
		text-decoration: none;
		margin-bottom: 2rem;
		transition: gap 0.3s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
	}

	.back-link::before {
		content: '←';
		font-size: 0.9rem;
		transition: transform 0.3s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
	}

	.back-link:hover {
		color: var(--color-ink, #1B1917);
		gap: 0.65rem;
	}

	.back-link:hover::before {
		transform: translateX(-2px);
	}

	.back-indicator {
		font-size: 0.75rem;
		color: var(--color-sage-700, #40416C);
		margin-bottom: 1rem;
	}

	/* Step indicator */
	.step-indicator {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 2.5rem;
	}

	.step {
		display: flex;
		align-items: center;
		gap: 0.45rem;
	}

	.step-num {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 1.6rem;
		height: 1.6rem;
		font-size: 0.7rem;
		font-weight: 600;
		border-radius: 50%;
		border: 1px solid var(--color-stone-300, #D1CCC3);
		color: var(--color-stone-400, #A9A296);
		transition: all 0.25s;
	}

	.step.active .step-num {
		background: var(--color-ink, #1B1917);
		border-color: var(--color-ink, #1B1917);
		color: #fff;
	}

	.step.active .step-label {
		color: var(--color-ink, #1B1917);
	}

	.step.completed .step-num {
		background: var(--color-sage-600, #4A5D42);
		border-color: var(--color-sage-600, #4A5D42);
		color: #fff;
	}

	.step-label {
		font-size: 0.65rem;
		font-weight: 500;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		display: none;
	}

	@media (min-width: 640px) {
		.step-label { display: block; }
	}

	.step-line {
		flex: 1;
		height: 1px;
		background: var(--color-stone-200, #E4E1DB);
		max-width: 3rem;
	}

	/* Field hint */
	.field-hint {
		font-size: 0.7rem;
		margin-top: 0.3rem;
		line-height: 1.4;
	}

	.field-hint.error {
		color: #9b3a30;
	}

	.input-group input.input-error {
		border-color: #9b3a30;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 2.8rem);
		font-weight: 300;
		line-height: 1.1;
		margin-bottom: 2.5rem;
	}

	.payment-desc {
		font-size: 0.9rem;
		color: var(--color-stone-500, #857E72);
		margin-bottom: 2rem;
		line-height: 1.6;
	}

	.booking-layout {
		display: grid;
		grid-template-columns: 1fr;
		gap: 3rem;
	}

	@media (min-width: 768px) {
		.booking-layout { grid-template-columns: 1.2fr 0.8fr; }
	}

	.booking-form, .payment-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.form-section-title {
		font-family: var(--font-display);
		font-size: 1.2rem;
		font-weight: 400;
		margin-top: 1rem;
	}

	.form-error {
		padding: 0.8rem 1rem;
		font-size: 0.8rem;
		background: rgba(180, 60, 50, 0.08);
		color: #9b3a30;
		border: 1px solid rgba(180, 60, 50, 0.15);
	}

	.input-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
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

	.input-group input {
		padding: 0.85rem 1rem;
		font-family: var(--font-body);
		font-size: 0.9rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: transparent;
		color: var(--color-ink, #1B1917);
		outline: none;
		transition: border-color 0.2s;
	}

	.input-group input:focus {
		border-color: var(--color-ink, #1B1917);
	}

	.input-group input::placeholder {
		color: var(--color-stone-300, #D1CCC3);
	}

	.payment-toggle {
		display: flex;
		gap: 0.5rem;
	}

	.toggle-btn {
		flex: 1;
		padding: 0.7rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		background: transparent;
		color: var(--color-stone-500, #857E72);
		border: 1px solid var(--color-stone-200, #E4E1DB);
		cursor: pointer;
		transition: all 0.15s;
	}

	.toggle-btn.active {
		border-color: var(--color-ink, #1B1917);
		color: var(--color-ink, #1B1917);
		background: var(--color-stone-50, #F7F6F2);
	}

	.toggle-btn:hover:not(.active) {
		border-color: var(--color-stone-400, #A9A296);
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
		transition: opacity 0.2s, transform 0.2s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
		margin-top: 0.5rem;
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

	@media (max-width: 639px) {
		.submit-btn {
			width: 100%;
		}
	}

	.form-note {
		font-size: 0.8rem;
		color: var(--color-stone-400, #A9A296);
		line-height: 1.6;
	}

	.crypto-section {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.crypto-info {
		font-size: 0.9rem;
		color: var(--color-stone-500, #857E72);
		line-height: 1.6;
	}

	/* Summary */
	.booking-summary {
		border: 1px solid var(--color-stone-200, #E4E1DB);
		padding: 2rem;
		position: sticky;
		top: 5rem;
		align-self: start;
	}

	.booking-summary::before {
		content: '';
		position: absolute;
		top: -1px;
		left: 0;
		width: 3rem;
		height: 2px;
		background: var(--color-brass-400, #B8A475);
	}

	.summary-title {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
		margin-bottom: 1.5rem;
	}

	.summary-details {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.summary-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
	}

	.summary-label {
		font-size: 0.8rem;
		color: var(--color-stone-500, #857E72);
	}

	.summary-value {
		font-size: 0.85rem;
		font-weight: 500;
	}

	.summary-divider {
		height: 1px;
		background: var(--color-stone-200, #E4E1DB);
		margin: 1.25rem 0;
	}

	.summary-pricing {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.summary-total {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		padding-top: 1.25rem;
		border-top: 1px solid var(--color-stone-200, #E4E1DB);
		margin-top: 1.25rem;
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
	}

	.summary-ref {
		margin-top: 1.25rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--color-stone-100, #F0EEEA);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
	}

	.summary-ref-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	.summary-ref-code {
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
	}

	@media (max-width: 639px) {
		.page {
			padding: 4rem 1.25rem 3rem;
		}

		.back-link {
			font-size: 0.7rem;
			margin-bottom: 1.5rem;
		}

		.page-title {
			font-size: clamp(1.6rem, 5vw, 2.2rem);
			margin-bottom: 1.75rem;
		}

		.booking-layout {
			gap: 2rem;
		}

		.input-row {
			grid-template-columns: 1fr;
			gap: 0.75rem;
		}

		.input-group input {
			font-size: 0.9rem;
		}

		.form-section-title {
			font-size: 1.1rem;
		}

		.payment-toggle {
			flex-direction: column;
			gap: 0.4rem;
		}

		.booking-summary {
			padding: 1.5rem;
		}

		.summary-title {
			font-size: 1.15rem;
			margin-bottom: 1.25rem;
		}

		.summary-total {
			font-size: 1.15rem;
		}
	}
</style>
