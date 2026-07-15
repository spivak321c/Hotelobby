<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';

	let full_name = $state('');
	let email = $state('');
	let phone = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);

	function passwordsMatch(): boolean {
		return password === confirmPassword && password.length > 0;
	}

	async function handleRegister(e: Event) {
		e.preventDefault();

		if (!passwordsMatch()) return;

		loading = true;
		try {
			await auth.register(full_name, email, password, phone || undefined);
			goto('/dashboard');
		} catch {
			// error is set in the store
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Create Account — The Lobby</title>
</svelte:head>

<div class="auth-page">
	<div class="auth-split">
		<div class="auth-image">
			<img src="https://images.unsplash.com/photo-1571896349842-33c89424de2d?w=800&q=80" alt="The Lobby Resort" loading="lazy" />
			<div class="auth-image-overlay"></div>
			<div class="auth-image-text">
				<p class="auth-image-tag">The Lobby</p>
				<p class="auth-image-desc">Boutique Luxury Resort</p>
			</div>
		</div>
		<div class="auth-card">
		<p class="section-tag">Join Us <span class="section-tag-line"></span></p>
		<h1 class="auth-title">Create Account</h1>

		<form class="auth-form" onsubmit={handleRegister}>
			{#if $auth.error}
				<div class="form-error">{$auth.error}</div>
			{/if}

			<div class="input-group">
				<label for="name">Full Name</label>
				<input
					id="name"
					type="text"
					bind:value={full_name}
					placeholder="John Doe"
					required
				/>
			</div>

			<div class="input-group">
				<label for="email">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
				/>
			</div>

			<div class="input-group">
				<label for="phone">Phone</label>
				<input
					id="phone"
					type="tel"
					bind:value={phone}
					placeholder="+1 (555) 000-0000"
				/>
			</div>

			<div class="input-group">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					minlength="8"
				/>
			</div>

			<div class="input-group">
				<label for="confirm-password">Confirm Password</label>
				<input
					id="confirm-password"
					type="password"
					bind:value={confirmPassword}
					placeholder="••••••••"
					required
					class:invalid={confirmPassword.length > 0 && !passwordsMatch()}
				/>
				{#if confirmPassword.length > 0 && !passwordsMatch()}
					<span class="field-error">Passwords do not match</span>
				{/if}
			</div>

			<button type="submit" class="submit-btn" disabled={loading || !passwordsMatch()}>
				{loading ? 'Creating account...' : 'Create Account'}
			</button>
		</form>

		<p class="auth-footer">
			Already have an account? <a href="/auth/login">Sign in</a>
		</p>
	</div>
	</div>
</div>

<style>
	.auth-page {
		min-height: 80vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 5rem 1.5rem 4rem;
	}

	.auth-split {
		display: grid;
		grid-template-columns: 1fr;
		max-width: 56rem;
		width: 100%;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	@media (min-width: 768px) {
		.auth-split {
			grid-template-columns: 0.8fr 1.2fr;
		}
	}

	.auth-image {
		position: relative;
		aspect-ratio: 4 / 3;
		overflow: hidden;
		display: none;
	}

	@media (min-width: 768px) {
		.auth-image {
			display: block;
			aspect-ratio: auto;
			min-height: 100%;
		}
	}

	.auth-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.auth-image-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(to bottom, rgba(26, 23, 20, 0.2), rgba(26, 23, 20, 0.55));
	}

	.auth-image-text {
		position: absolute;
		bottom: 2rem;
		left: 2rem;
		color: #fff;
	}

	.auth-image-tag {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		margin-bottom: 0.15rem;
	}

	.auth-image-desc {
		font-size: 0.7rem;
		letter-spacing: 0.15em;
		text-transform: uppercase;
		opacity: 0.7;
	}

	.auth-card {
		padding: 3rem 2rem;
	}

	@media (min-width: 768px) {
		.auth-card {
			padding: 3.5rem 3rem;
		}
	}

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

	.auth-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 2.8rem);
		font-weight: 300;
		line-height: 1.1;
		margin-bottom: 2.5rem;
	}

	.auth-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
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

	.input-group input.invalid {
		border-color: #9b3a30;
	}

	.field-error {
		font-size: 0.75rem;
		color: #9b3a30;
		margin-top: 0.35rem;
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

	.auth-footer {
		margin-top: 2rem;
		text-align: center;
		font-size: 0.85rem;
		color: var(--color-stone-500, #857E72);
	}

	.auth-footer a {
		color: var(--color-ink, #1B1917);
		text-decoration: none;
		border-bottom: 1px solid var(--color-ink, #1B1917);
		padding-bottom: 1px;
		transition: opacity 0.2s;
	}

	.auth-footer a:hover {
		opacity: 0.5;
	}

	@media (max-width: 639px) {
		.auth-page {
			padding: 4rem 1.25rem 3rem;
		}

		.auth-card {
			padding: 2.5rem 1.5rem;
		}

		.auth-title {
			font-size: clamp(1.6rem, 5vw, 2.2rem);
			margin-bottom: 2rem;
		}

		.input-group input {
			font-size: 0.9rem;
		}

		.submit-btn {
			width: 100%;
		}
	}
</style>
