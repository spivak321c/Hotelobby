<script lang="ts">
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import Lenis from 'lenis';
	import gsap from 'gsap';
	import { ScrollTrigger } from 'gsap/ScrollTrigger';
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { auth } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import Toast from '$lib/components/ui/Toast.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	let { children } = $props();
	let mobileMenuOpen = $state(false);

	const isHome = $derived(page.url.pathname === '/');
	const isAdmin = $derived(page.url.pathname.startsWith('/admin'));

	let lenis: Lenis | null = null;

	const navItems: { href: string; label: string; desc: string; auth?: boolean; admin?: boolean }[] = [
		{ href: '/rooms', label: 'Rooms', desc: 'Explore our stays' },
		{ href: '/lookup', label: 'My Booking', desc: 'Find a reservation' },
		{ href: '/dashboard', label: 'Dashboard', desc: 'Your account', auth: true },
		{ href: '/admin', label: 'Admin', desc: 'Manage hotel', auth: true, admin: true },
	];

	function closeMenu() {
		mobileMenuOpen = false;
	}

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) closeMenu();
	}

	function handleLogout() {
		auth.logout();
		goto('/');
		closeMenu();
	}

	$effect(() => {
		if (mobileMenuOpen) {
			document.body.style.overflow = 'hidden';
		} else {
			document.body.style.overflow = '';
		}
		return () => { document.body.style.overflow = ''; };
	});

	$effect(() => {
		function onKeydown(e: KeyboardEvent) {
			if (e.key === 'Escape') closeMenu();
		}
		function onResize() {
			if (window.innerWidth >= 768) closeMenu();
		}
		if (mobileMenuOpen) {
			window.addEventListener('keydown', onKeydown);
			window.addEventListener('resize', onResize);
		}
		return () => {
			window.removeEventListener('keydown', onKeydown);
			window.removeEventListener('resize', onResize);
		};
	});

	onMount(() => {
		gsap.registerPlugin(ScrollTrigger);

		lenis = new Lenis({ autoRaf: false });

		lenis.on('scroll', ScrollTrigger.update);

		gsap.ticker.add((time) => {
			lenis?.raf(time * 1000);
		});
		gsap.ticker.lagSmoothing(0);

		return () => {
			gsap.ticker.lagSmoothing(1);
			gsap.ticker.remove(() => {});
			lenis?.destroy();
		};
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,400;0,500;0,600;1,300;1,400&family=Inter:wght@300;400;500;600&display=swap"
		rel="stylesheet"
	/>
	<title>The Lobby</title>
</svelte:head>

{#snippet navLink(href: string, label: string)}
	<a
		{href}
		class="nav-link text-[0.8rem] font-medium tracking-wide uppercase text-stone-400 hover:text-ink transition-colors duration-200 relative"
	>
		{label}
	</a>
{/snippet}

<div class="min-h-dvh flex flex-col">
	{#if !isAdmin}
		<header
			class="fixed top-0 left-0 right-0 z-50 glass border-b border-stone-200/50 transition-all duration-300"
			class:menu-open={mobileMenuOpen}
		>
			<nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
				<a
					href="/"
					class="font-display text-lg tracking-[0.18em] uppercase text-ink relative z-50"
				>
					The Lobby
				</a>

				<div class="hidden items-center gap-8 md:flex">
					{@render navLink('/rooms', 'Rooms')}
					{@render navLink('/lookup', 'My Booking')}

					{#if $auth.token}
						{#if $auth.role === 'admin'}
							{@render navLink('/admin', 'Admin')}
						{:else}
							{@render navLink('/dashboard', 'Dashboard')}
						{/if}
						<button
							onclick={() => { auth.logout(); goto('/'); }}
							class="text-[0.8rem] font-medium tracking-wide uppercase text-stone-400 hover:text-ink transition-colors duration-200"
						>
							Logout
						</button>
					{:else}
						<a
							href="/auth/login"
							class="px-5 py-2.5 text-[0.75rem] font-semibold tracking-widest uppercase border border-stone-300 text-ink hover:bg-ink hover:text-cream transition-all duration-200"
						>
							Sign In
						</a>
					{/if}
				</div>

				<button
					class="md:hidden relative z-50 flex h-11 w-11 items-center justify-center text-stone-400 hover:text-ink transition-colors duration-200"
					onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
					aria-label={mobileMenuOpen ? 'Close menu' : 'Open menu'}
					aria-expanded={mobileMenuOpen}
				>
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						{#if mobileMenuOpen}
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
						{:else}
							<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
						{/if}
					</svg>
				</button>
			</nav>
		</header>

		<!-- Mobile overlay menu -->
		{#if mobileMenuOpen}
			<div
				class="mobile-overlay"
				transition:fade={{ duration: 250 }}
				role="dialog"
				aria-modal="true"
				aria-label="Navigation menu"
			>
				<div class="mobile-menu-inner" onclick={handleOverlayClick} onkeydown={() => {}} role="presentation">
					<span class="mobile-menu-eyebrow">Menu</span>

					<nav class="mobile-menu-nav">
						{#each navItems as item, i}
							{#if !item.auth || $auth.token}
								{#if !item.admin || $auth.role === 'admin'}
									<a
										href={item.href}
										class="mobile-nav-link"
										style="--i: {i}"
										onclick={closeMenu}
									>
										<span class="mobile-nav-label">{item.label}</span>
										<span class="mobile-nav-desc">{item.desc}</span>
									</a>
								{/if}
							{/if}
						{/each}
					</nav>

					<div class="mobile-menu-foot">
						{#if $auth.token}
							<span class="mobile-foot-label">Signed in</span>
							<button onclick={handleLogout} class="mobile-foot-action">Logout</button>
						{:else}
							<a href="/auth/login" class="mobile-foot-action" onclick={closeMenu}>
								Sign In
							</a>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	{/if}

	<main class="flex-1" class:blurred={mobileMenuOpen}>
		{@render children()}
	</main>

	<Toast />

	{#if !isAdmin}
		<footer class="border-t border-stone-200 bg-cream">
			<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<div class="flex flex-col gap-6 py-12 sm:flex-row sm:items-center sm:justify-between">
					<div class="flex flex-col gap-1">
						<p class="font-display text-lg tracking-[0.18em] uppercase text-ink">The Lobby</p>
						<p class="text-xs tracking-[0.08em] uppercase text-stone-400">Boutique Luxury Resort</p>
					</div>
					<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-8">
						<a href="/rooms" class="text-xs tracking-wide uppercase text-stone-400 hover:text-ink transition-colors">Rooms</a>
						<a href="/lookup" class="text-xs tracking-wide uppercase text-stone-400 hover:text-ink transition-colors">My Booking</a>
						<a href="/auth/login" class="text-xs tracking-wide uppercase text-stone-400 hover:text-ink transition-colors">Sign In</a>
					</div>
					<p class="text-xs text-stone-400 hidden sm:block tabular">© {new Date().getFullYear()} The Lobby Resort</p>
				</div>
			</div>
		</footer>
	{/if}
</div>

<style>
	/* ─── MOBILE OVERLAY ─── */
	.mobile-overlay {
		position: fixed;
		inset: 0;
		z-index: 40;
		background: var(--color-cream, #FAF8F5);
		display: flex;
		align-items: center;
		overflow-y: auto;
	}

	.mobile-menu-inner {
		width: 100%;
		min-height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: center;
		padding: 6rem 2rem 3rem;
		cursor: default;
	}

	/* Eyebrow label */
	.mobile-menu-eyebrow {
		font-size: 0.6rem;
		font-weight: 600;
		letter-spacing: 0.3em;
		text-transform: uppercase;
		color: var(--color-stone-300, #D6CFC2);
		margin-bottom: 2rem;
		animation: mobileIn 0.5s cubic-bezier(0.16, 1, 0.3, 1) 0.05s both;
	}

	/* Nav list */
	.mobile-menu-nav {
		display: flex;
		flex-direction: column;
	}

	.mobile-nav-link {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		padding: 0.75rem 0;
		text-decoration: none;
		cursor: pointer;
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		-webkit-tap-highlight-color: transparent;
		position: relative;
		animation: mobileIn 0.55s cubic-bezier(0.16, 1, 0.3, 1) both;
		animation-delay: calc(0.1s + var(--i, 0) * 0.07s);
	}

	/* Hairline between items */
	.mobile-nav-link::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 0.5px;
		background: var(--color-stone-100, #F0EEEA);
		transform-origin: left;
		animation: ruleIn 0.6s cubic-bezier(0.16, 1, 0.3, 1) both;
		animation-delay: calc(0.15s + var(--i, 0) * 0.07s);
	}

	.mobile-nav-label {
		font-family: var(--font-display);
		font-size: clamp(2rem, 7vw, 2.8rem);
		font-weight: 300;
		letter-spacing: -0.015em;
		line-height: 1.1;
		color: var(--color-ink, #1B1917);
		transition: color 0.25s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.25s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.mobile-nav-desc {
		font-size: 0.78rem;
		font-weight: 400;
		line-height: 1.4;
		color: var(--color-stone-400, #A9A296);
		font-style: italic;
		transition: color 0.25s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.mobile-nav-link:hover .mobile-nav-label {
		color: var(--color-stone-700, #57534E);
	}

	.mobile-nav-link:active .mobile-nav-label {
		opacity: 0.5;
	}

	/* Footer (account) */
	.mobile-menu-foot {
		margin-top: 3rem;
		display: flex;
		align-items: baseline;
		gap: 0.875rem;
		animation: mobileIn 0.5s cubic-bezier(0.16, 1, 0.3, 1) 0.4s both;
	}

	.mobile-foot-label {
		font-size: 0.68rem;
		font-weight: 500;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	.mobile-foot-action {
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-ink, #1B1917);
		text-decoration: none;
		background: none;
		border: none;
		border-bottom: 1px solid var(--color-ink, #1B1917);
		cursor: pointer;
		padding: 2px 0;
		-webkit-tap-highlight-color: transparent;
		transition: opacity 0.25s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.mobile-foot-action:hover {
		opacity: 0.5;
	}

	/* Keyframes */
	@keyframes mobileIn {
		from {
			opacity: 0;
			transform: translateX(-12px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes ruleIn {
		from { transform: scaleX(0); }
		to { transform: scaleX(1); }
	}

	/* ─── NAV STATE ─── */
	header.menu-open {
		background: transparent;
		border-color: transparent;
	}

	:global(.nav-link::after) {
		content: '';
		position: absolute;
		left: 0;
		bottom: -4px;
		width: 0;
		height: 1px;
		background: var(--color-brass-400, #B8A475);
		transition: width 0.3s var(--ease-out-expo);
	}
	:global(.nav-link:hover::after) {
		width: 100%;
	}

	/* ─── REDUCED MOTION ─── */
	@media (prefers-reduced-motion: reduce) {
		.mobile-nav-link,
		.mobile-menu-eyebrow,
		.mobile-menu-foot {
			animation: none;
			opacity: 1;
			transform: none;
		}

		.mobile-nav-link::after {
			animation: none;
			transform: none;
		}
	}
</style>
