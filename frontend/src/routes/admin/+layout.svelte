<script lang="ts">
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import { auth } from '$lib/stores/auth.svelte';

	let { children } = $props();
	let sidebarOpen = $state(false);

	const navItems = [
		{ href: '/admin', label: 'Overview', exact: true },
		{ href: '/admin/reservations', label: 'Reservations', exact: false },
		{ href: '/admin/rooms', label: 'Rooms', exact: false },
		{ href: '/admin/room-types', label: 'Room Types', exact: false },
		{ href: '/admin/pricing', label: 'Pricing', exact: false },
		{ href: '/admin/inventory', label: 'Inventory', exact: false },
		{ href: '/admin/customers', label: 'Customers', exact: false },
		{ href: '/admin/admins', label: 'Admins', exact: false },
		{ href: '/admin/reports', label: 'Reports', exact: false }
	];

	function isActive(href: string, exact: boolean): boolean {
		const path = page.url.pathname;
		return exact ? path === href : path.startsWith(href);
	}

	// Persist the current admin path so re-login can restore the right page
	$effect(() => {
		if (!browser) return;
		const p = page.url.pathname;
		if (p.startsWith('/admin') && $auth.role === 'admin') {
			try { sessionStorage.setItem('admin_last_path', p); } catch {}
		}
	});
</script>

{#if $auth.role === 'admin'}
	<div class="admin-layout">
		<!-- Mobile toggle -->
		<button class="sidebar-toggle" onclick={() => { sidebarOpen = !sidebarOpen; }}>
			{sidebarOpen ? '\u2715' : '\u2630'}
		</button>

		<!-- Overlay -->
		{#if sidebarOpen}
			<div class="sidebar-overlay" onclick={() => { sidebarOpen = false; }} role="presentation"></div>
		{/if}

		<!-- Sidebar -->
		<aside class="sidebar" class:open={sidebarOpen}>
			<div class="sidebar-header">
				<a href="/" class="logo">The Lobby</a>
				<span class="admin-badge">Admin</span>
			</div>

			<nav class="sidebar-nav">
				{#each navItems as item}
					<a
						href={item.href}
						class="nav-link"
						class:active={isActive(item.href, item.exact)}
						onclick={() => { sidebarOpen = false; }}
					>
						{item.label}
					</a>
				{/each}
			</nav>

			<div class="sidebar-footer">
				<a href="/dashboard" class="footer-link">Customer View</a>
				<button class="footer-link" onclick={() => auth.logout()}>
					Sign Out
				</button>
			</div>
		</aside>

		<!-- Main content -->
		<main class="main-content">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="unauthed">
		<p>Admin access required.</p>
		<a href="/admin/login">Admin Sign In</a>
	</div>
{/if}

<style>
	.admin-layout {
		display: flex;
		min-height: 100vh;
		background: var(--color-cream, #FAFAF5);
	}

	/* Sidebar */
	.sidebar {
		width: 15rem;
		background: var(--color-ink, #1B1917);
		color: var(--color-cream, #FAFAF5);
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
		position: fixed;
		top: 0;
		left: 0;
		bottom: 0;
		z-index: 30;
		transition: transform 0.3s ease;
	}

	@media (max-width: 768px) {
		.sidebar {
			transform: translateX(-100%);
		}
		.sidebar.open {
			transform: translateX(0);
		}
	}

	.sidebar-overlay {
		display: none;
	}

	@media (max-width: 768px) {
		.sidebar-overlay {
			display: block;
			position: fixed;
			inset: 0;
			background: rgba(0, 0, 0, 0.5);
			z-index: 25;
		}
	}

	.sidebar-toggle {
		display: none;
		position: fixed;
		top: 1rem;
		left: 1rem;
		z-index: 35;
		background: var(--color-ink, #1B1917);
		color: var(--color-cream, #FAFAF5);
		border: none;
		padding: 0.5rem 0.75rem;
		font-size: 1.1rem;
		cursor: pointer;
	}

	@media (max-width: 768px) {
		.sidebar-toggle { display: block; }
	}

	.sidebar-header {
		padding: 1.5rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}

	.logo {
		font-family: var(--font-display);
		font-size: 1.2rem;
		font-weight: 300;
		color: var(--color-cream, #FAFAF5);
		text-decoration: none;
		display: block;
		margin-bottom: 0.5rem;
	}

	.admin-badge {
		font-size: 0.6rem;
		font-weight: 600;
		letter-spacing: 0.15em;
		text-transform: uppercase;
		color: var(--color-sage-700, #40416C);
		border: 1px solid var(--color-sage-700, #40416C);
		padding: 0.15rem 0.5rem;
	}

	.sidebar-nav {
		flex: 1;
		padding: 1rem 0;
		overflow-y: auto;
	}

	.nav-link {
		display: block;
		padding: 0.6rem 1.5rem;
		font-size: 0.8rem;
		font-weight: 400;
		letter-spacing: 0.02em;
		text-decoration: none;
		color: rgba(255, 255, 255, 0.5);
		transition: all 0.15s;
	}

	.nav-link:hover {
		color: rgba(255, 255, 255, 0.85);
		background: rgba(255, 255, 255, 0.05);
	}

	.nav-link.active {
		color: var(--color-cream, #FAFAF5);
		background: rgba(255, 255, 255, 0.08);
	}

	.sidebar-footer {
		padding: 1rem 1.5rem;
		border-top: 1px solid rgba(255, 255, 255, 0.08);
	}

	.footer-link {
		display: block;
		font-size: 0.75rem;
		color: rgba(255, 255, 255, 0.4);
		text-decoration: none;
		background: none;
		border: none;
		padding: 0.3rem 0;
		cursor: pointer;
		font-family: inherit;
		text-align: left;
		transition: color 0.15s;
	}

	.footer-link:hover {
		color: rgba(255, 255, 255, 0.75);
	}

	/* Main */
	.main-content {
		flex: 1;
		margin-left: 15rem;
		padding: 2rem;
		min-height: 100vh;
	}

	@media (max-width: 768px) {
		.main-content {
			margin-left: 0;
			padding: 1rem;
			padding-top: 3.5rem;
		}
	}

	.unauthed {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		gap: 1rem;
		color: var(--color-stone-500, #857E72);
	}

	.unauthed a {
		padding: 0.6rem 1.2rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		background: var(--color-ink, #1B1917);
		color: #fff;
		text-decoration: none;
		transition: opacity 0.2s;
	}

	.unauthed a:hover { opacity: 0.85; }
</style>
