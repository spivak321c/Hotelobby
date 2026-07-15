<script lang="ts">
	import { page } from '$app/state';

	const ref = $derived(page.url.searchParams.get('ref') ?? '');
	const email = $derived(page.url.searchParams.get('email') ?? '');
</script>

<svelte:head>
	<title>Payment Confirmed — The Lobby</title>
</svelte:head>

<div class="confirmation-page">
	<div class="confirmation-card">
		<div class="check-icon">
			<svg class="check-svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
				<path class="check-path" d="M20 6L9 17l-5-5"/>
			</svg>
		</div>

		<p class="confirm-eyebrow">Payment Confirmed</p>
		<h1 class="confirm-title">Your stay is booked</h1>
		<p class="confirm-desc">
			Your payment has been processed and your booking is confirmed.
			A confirmation email has been sent to <strong>{email}</strong>.
		</p>

		<div class="ref-block">
			<span class="ref-label">Reservation Reference</span>
			<span class="ref-code">{ref}</span>
		</div>

		<div class="confirm-actions">
			<a href="/lookup" class="btn-primary">Look Up Booking</a>
			<a href="/" class="btn-ghost">Return Home</a>
		</div>
	</div>
</div>

<style>
	.confirmation-page {
		min-height: 80vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 6rem 1.5rem 4rem;
	}

	.confirmation-card {
		max-width: 28rem;
		text-align: center;
	}

	.check-icon {
		width: 64px;
		height: 64px;
		margin: 0 auto 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		background: var(--color-sage-100, #E8EDE5);
		color: var(--color-sage-700, #40416C);
		position: relative;
	}

	.check-icon::after {
		content: '';
		position: absolute;
		inset: -6px;
		border-radius: 50%;
		border: 1px solid var(--color-sage-300, #B8C8B0);
		opacity: 0;
		animation: ringExpand 1s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1)) 0.4s forwards;
	}

	@keyframes ringExpand {
		to {
			transform: scale(1.3);
			opacity: 0;
		}
	}

	.check-path {
		stroke-dasharray: 30;
		stroke-dashoffset: 30;
		animation: drawCheck 0.6s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1)) 0.3s forwards;
	}

	@keyframes drawCheck {
		to { stroke-dashoffset: 0; }
	}

	.confirm-eyebrow {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.25em;
		text-transform: uppercase;
		color: var(--color-brass-500, #A89260);
		margin-bottom: 0.75rem;
	}

	.confirm-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 2.8rem);
		font-weight: 300;
		line-height: 1.1;
		margin-bottom: 1rem;
	}

	.confirm-desc {
		font-size: 0.95rem;
		line-height: 1.7;
		color: var(--color-stone-500, #857E72);
		margin-bottom: 2.5rem;
	}

	.confirm-desc strong {
		color: var(--color-ink, #1B1917);
	}

	.ref-block {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1.5rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		margin-bottom: 2.5rem;
	}

	.ref-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.5rem;
	}

	.ref-code {
		font-family: var(--font-display);
		font-size: 1.5rem;
		font-weight: 400;
		letter-spacing: 0.05em;
	}

	.confirm-actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
		flex-wrap: wrap;
	}

	.btn-primary {
		display: inline-block;
		padding: 0.85rem 2rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		text-decoration: none;
		transition: opacity 0.2s;
	}

	.btn-primary:hover {
		opacity: 0.85;
	}

	.btn-ghost {
		display: inline-block;
		padding: 0.85rem 2rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		color: var(--color-ink, #1B1917);
		text-decoration: none;
		transition: border-color 0.2s;
	}

	.btn-ghost:hover {
		border-color: var(--color-ink, #1B1917);
	}

	@media (max-width: 639px) {
		.confirmation-page {
			padding: 5rem 1.25rem 3rem;
		}

		.check-icon {
			width: 52px;
			height: 52px;
			margin-bottom: 1.5rem;
		}

		.confirm-title {
			font-size: clamp(1.6rem, 5vw, 2rem);
		}

		.confirm-desc {
			font-size: 0.9rem;
			line-height: 1.75;
			margin-bottom: 2rem;
		}

		.ref-block {
			padding: 1.25rem;
			margin-bottom: 2rem;
		}

		.ref-code {
			font-size: 1.2rem;
		}

		.confirm-actions {
			flex-direction: column;
			gap: 0.75rem;
		}

		.btn-primary,
		.btn-ghost {
			width: 100%;
			text-align: center;
		}
	}
</style>
