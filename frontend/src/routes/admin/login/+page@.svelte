<script lang="ts">
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { auth } from '$lib/stores/auth.svelte';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);

	async function handleLogin(e: Event) {
		e.preventDefault();
		loading = true;

		try {
			await auth.adminLogin(email, password);
			const saved = browser ? sessionStorage.getItem('admin_last_path') : null;
			goto(saved && saved.startsWith('/admin') ? saved : '/admin');
		} catch {
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Admin Sign In — The Lobby</title>
</svelte:head>

<div class="auth-page">
	<div class="auth-card">
		<p class="section-tag">Admin Access</p>
		<h1 class="auth-title">Admin Sign In</h1>

		<form class="auth-form" onsubmit={handleLogin}>
			{#if $auth.error}
				<div class="form-error">{$auth.error}</div>
			{/if}

			<div class="input-group">
				<label for="email">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="admin@thelobby.com"
					required
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
				/>
			</div>

			<button type="submit" class="submit-btn" disabled={loading}>
				{loading ? 'Signing in...' : 'Sign In'}
			</button>
		</form>

		<p class="auth-footer">
			<a href="/auth/login">Customer Sign In</a>
		</p>
	</div>
</div>

<style>
	.auth-page {
		min-height: 80vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 6rem 1.5rem 4rem;
	}

	.auth-card {
		width: 100%;
		max-width: 24rem;
	}

	.section-tag {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 1rem;
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
	}

	.submit-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.submit-btn:not(:disabled):hover {
		opacity: 0.85;
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
</style>
