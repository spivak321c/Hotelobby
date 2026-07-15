<script lang="ts">
	let {
		width = '100%',
		height = '1rem',
		rounded = false,
		class: className = '',
	}: {
		width?: string;
		height?: string;
		rounded?: boolean;
		class?: string;
	} = $props();
</script>

<div
	class="skeleton {className}"
	style="width: {width}; height: {height}; border-radius: {rounded ? '9999px' : '0'};"
	aria-hidden="true"
></div>

<style>
	.skeleton {
		background: var(--color-stone-200, #E4E1DB);
		animation: pulse 1.5s ease-in-out infinite;
	}

	@keyframes pulse {
		0%, 100% { opacity: 0.3; }
		50% { opacity: 0.7; }
	}
</style>
