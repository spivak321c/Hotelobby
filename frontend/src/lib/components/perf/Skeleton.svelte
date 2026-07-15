<script lang="ts">
	let {
		class: className = '',
		variant = 'text',
		width,
		height,
		count = 1
	}: {
		class?: string;
		variant?: 'text' | 'circular' | 'rectangular' | 'card';
		width?: string;
		height?: string;
		count?: number;
	} = $props();

	const heights: Record<string, string> = {
		text: '1rem',
		circular: '2.5rem',
		rectangular: '10rem',
		card: '12rem'
	};

	const widths: Record<string, string> = {
		text: '100%',
		circular: '2.5rem',
		rectangular: '100%',
		card: '100%'
	};

	const radius: Record<string, string> = {
		text: '0.25rem',
		circular: '50%',
		rectangular: '0.5rem',
		card: '0.75rem'
	};
</script>

{#each { length: count } as _, i}
	<div
		class="animate-pulse bg-surface-alt {className}"
		style:width={width ?? widths[variant]}
		style:height={height ?? heights[variant]}
		style:border-radius={radius[variant]}
		role="presentation"
		aria-hidden="true"
	></div>
{/each}
