<script lang="ts">
	import { toast } from '$lib/stores/toast.svelte';
	import { fade, slide } from 'svelte/transition';

	const icons: Record<string, string> = {
		success: 'M20 6L9 17l-5-5',
		error: 'M18 6L6 18M6 6l12 12',
		info: 'M12 16v-4M12 8h.01',
		warning: 'M12 9v4M12 17h.01',
	};
</script>

{#each toast.toasts as t (t.id)}
	<div
		class="toast toast-{t.type}"
		role="status"
		aria-live="polite"
		transition:slide={{ duration: 300, axis: 'y' }}
	>
		<svg class="toast-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
			<path d={icons[t.type]} stroke-linecap="round" stroke-linejoin="round"/>
		</svg>
		<div class="toast-body">
			<p class="toast-message">{t.message}</p>
			{#if t.description}
				<p class="toast-desc">{t.description}</p>
			{/if}
		</div>
		<button class="toast-close" onclick={() => toast.remove(t.id)} aria-label="Dismiss">
			<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
				<path d="M18 6L6 18M6 6l12 12"/>
			</svg>
		</button>
	</div>
{/each}

<style>
	.toast {
		position: fixed;
		top: 5rem;
		right: 1.25rem;
		z-index: 9999;
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 1rem 1.15rem;
		max-width: 22rem;
		background: var(--color-cream, #FAF8F5);
		border: 1px solid var(--color-stone-200, #E4E1DB);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
		pointer-events: auto;
	}

	.toast-icon {
		flex-shrink: 0;
		margin-top: 0.1rem;
	}

	.toast-success .toast-icon { color: var(--color-sage-600, #4A5D42); }
	.toast-error .toast-icon { color: #9b3a30; }
	.toast-info .toast-icon { color: #3a6b9b; }
	.toast-warning .toast-icon { color: #9b8a30; }

	.toast-body {
		flex: 1;
		min-width: 0;
	}

	.toast-message {
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--color-ink, #1B1917);
		line-height: 1.4;
	}

	.toast-desc {
		font-size: 0.75rem;
		color: var(--color-stone-500, #857E72);
		margin-top: 0.15rem;
		line-height: 1.4;
	}

	.toast-close {
		flex-shrink: 0;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--color-stone-400, #A9A296);
		padding: 0.15rem;
		transition: color 0.15s;
	}

	.toast-close:hover {
		color: var(--color-ink, #1B1917);
	}

	@media (max-width: 639px) {
		.toast {
			top: auto;
			bottom: 1rem;
			right: 1rem;
			left: 1rem;
			max-width: none;
		}
	}
</style>
