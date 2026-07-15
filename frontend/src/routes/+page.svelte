<script lang="ts">
	import { onMount } from 'svelte';
	import { roomTypesApi } from '$lib/api/client';
	import type { RoomType } from '$lib/types/api';
	import {
		splitWords,
		prefersReducedMotion,
		initScrollReveals,
		initStaggerReveals,
		initWordStaggers,
	} from '$lib/animations';

	const fallbackRooms = [
		{ id: '1', name: 'Deluxe Suite', base_rate_daily: 180 },
		{ id: '2', name: 'Garden Villa', base_rate_daily: 320 },
		{ id: '3', name: 'Penthouse', base_rate_daily: 480 }
	] as RoomType[];

	const roomImages = [
		'https://images.unsplash.com/photo-1631049307264-da0ec9d70304?w=600&q=80',
		'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80',
		'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80'
	];

	const heroSlides = [
		'https://images.unsplash.com/photo-1566073771259-6a8506099945?w=1600&q=80',
		'https://images.unsplash.com/photo-1584132967334-10e028bd69f7?w=1600&q=80',
		'https://images.unsplash.com/photo-1582719508461-905c673771fd?w=1600&q=80'
	];

	const brandLogos = [
		'Travel + Leisure', 'Condé Nast', 'Forbes Travel',
		'Architectural Digest', 'The New York Times', 'Vogue',
	];

	let roomTypes = $state<RoomType[]>([]);
	let displayRooms = $derived(roomTypes.length > 0 ? roomTypes.slice(0, 3) : fallbackRooms);

	let currentSlide = $state(0);
	let slideDirection = $state<'next' | 'prev'>('next');

	function nextSlide() {
		slideDirection = 'next';
		currentSlide = (currentSlide + 1) % heroSlides.length;
	}

	function prevSlide() {
		slideDirection = 'prev';
		currentSlide = (currentSlide - 1 + heroSlides.length) % heroSlides.length;
	}

	onMount(() => {
		roomTypesApi.list().then((data) => {
			roomTypes = data;
		}).catch(() => {});

		const reduced = prefersReducedMotion();
		const heroInterval = setInterval(nextSlide, 6000);

		if (!reduced) {
			initScrollReveals(reduced);
			initStaggerReveals('[data-stagger-children]', reduced);
			initWordStaggers(reduced);
		} else {
			document.querySelectorAll('[data-reveal]').forEach((el) => {
				(el as HTMLElement).style.opacity = '1';
				(el as HTMLElement).style.transform = 'translateY(0)';
			});
		}

		return () => {
			clearInterval(heroInterval);
		};
	});
</script>

<svelte:head>
	<title>The Lobby — Where Time Stands Still</title>
	<meta name="description" content="A sanctuary of calm nestled in tropical greenery. Twelve rooms, each a private world of comfort and quiet elegance." />
</svelte:head>

<div class="home">
	<!-- ═══════════════ HERO ═══════════════ -->
	<section class="hero">
		{#each heroSlides as slide, i}
			<div
				class="hero-slide"
				class:active={currentSlide === i}
				class:prev={slideDirection === 'prev' && currentSlide === i}
				style="background-image: url({slide})"
			></div>
		{/each}

		<div class="hero-overlay"></div>

		<div class="hero-indicators">
			{#each heroSlides as _, i}
				<button
					class="hero-dot"
					class:active={currentSlide === i}
					onclick={() => { slideDirection = i > currentSlide ? 'next' : 'prev'; currentSlide = i; }}
					aria-label="Go to slide {i + 1}"
				></button>
			{/each}
		</div>

		<div class="hero-controls">
			<button onclick={prevSlide} aria-label="Previous slide" class="hero-control">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M15 18l-6-6 6-6"/>
				</svg>
			</button>
			<button onclick={nextSlide} aria-label="Next slide" class="hero-control">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M9 18l6-6-6-6"/>
				</svg>
			</button>
		</div>

		<div class="hero-content">
			<p class="hero-tag">Boutique Luxury Resort</p>
			<h1 class="hero-title">
				{#each splitWords('Where Time Stands Still') as { word, key }}
					<span class="hero-word" style="animation-delay: {0.5 + Number(key.split('-')[0]) * 0.08}s">{word}</span>
				{/each}
			</h1>
			<p class="hero-desc">
				A sanctuary of calm nestled in tropical greenery. Twelve rooms, each a private world of comfort and quiet elegance.
			</p>
			<div class="hero-btns">
				<a href="/rooms" class="btn-primary">Book Your Stay</a>
				<a href="/rooms" class="btn-ghost">Explore Rooms</a>
			</div>
		</div>

		<div class="hero-scroll-hint">
			<span class="scroll-hint-text">Scroll</span>
			<span class="scroll-hint-line"></span>
		</div>
	</section>

	<!-- ═══════════════ MARQUEE ═══════════════ -->
	<section class="marquee-section">
		<div class="marquee-track">
			<div class="marquee-content">
				{#each brandLogos as logo}
					<span class="marquee-item">{logo}</span>
				{/each}
			</div>
			<div class="marquee-content" aria-hidden="true">
				{#each brandLogos as logo}
					<span class="marquee-item">{logo}</span>
				{/each}
			</div>
		</div>
	</section>

	<!-- ═══════════════ FEATURED ROOMS ═══════════════ -->
	<section id="rooms" data-reveal class="section" data-stagger-children>
		<div class="section-header">
			<div>
				<p class="section-tag">Accommodations</p>
				<h2 class="section-title" data-word-stagger>
					{#each splitWords('Curated Rooms') as { word, key }}
						<span data-word class="title-word">{word}</span>
					{/each}
				</h2>
			</div>
			<a href="/rooms" class="section-link">View All</a>
		</div>

		<div class="rooms-track" data-lenis-prevent>
			{#each displayRooms as room, i}
				<a href="/rooms" class="room-card" data-stagger-item>
					<div class="room-card-img">
						<img src={roomImages[i] ?? roomImages[0]} alt={room.name} loading="lazy" />
						<div class="room-card-overlay"></div>
					</div>
					<div class="room-card-info">
						<div class="room-card-name">{room.name}</div>
						<div class="room-card-price">From {'$'}{room.base_rate_daily} / night</div>
					</div>
				</a>
			{/each}
		</div>
	</section>

	<!-- ═══════════════ ABOUT ═══════════════ -->
	<section id="about" data-reveal class="about">
		<div class="about-img">
			<img src="https://images.unsplash.com/photo-1571896349842-33c89424de2d?w=800&q=80" alt="Resort exterior" loading="lazy" />
		</div>
		<div class="about-body">
			<p class="section-tag">Our Story</p>
			<h2 class="about-title" data-word-stagger>
				{#each splitWords('Built on the belief that luxury is quiet') as { word, key }}
					<span data-word class="title-word">{word}</span>
				{/each}
			</h2>
			<p class="about-text">
				Founded in 2019, The Lobby was conceived as an antidote to the noise of modern travel. Every detail — from the hand-selected linens to the morning light in each room — is designed to help you slow down.
			</p>
			<a href="/rooms" class="text-link">Discover More</a>
		</div>
	</section>

	<!-- ═══════════════ AMENITIES ═══════════════ -->
	<section id="amenities" data-reveal class="amenities">
		<p class="section-tag" style="color: var(--stone-400);">Amenities</p>
		<h2 class="section-title" style="color: #fff;" data-word-stagger>
			{#each splitWords('Everything You Need') as { word, key }}
				<span data-word class="title-word">{word}</span>
			{/each}
		</h2>

		<div class="amenities-grid" data-stagger-children>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<path d="M3 3v18h18"/>
						<path d="M7 16l4-8 4 4 4-6"/>
					</svg>
				</div>
				<span class="amenity-label">Restaurant</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<circle cx="12" cy="12" r="3"/>
						<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
					</svg>
				</div>
				<span class="amenity-label">Spa & Wellness</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<path d="M2 20h20M6 20V8a2 2 0 012-2h8a2 2 0 012 2v12"/>
						<path d="M8 8h8"/>
					</svg>
				</div>
				<span class="amenity-label">Pool</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<path d="M18 20V10M12 20V4M6 20v-6"/>
					</svg>
				</div>
				<span class="amenity-label">Fitness Center</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<path d="M5 12h14M12 5l7 7-7 7"/>
					</svg>
				</div>
				<span class="amenity-label">High-Speed Wi-Fi</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<rect x="1" y="3" width="22" height="18" rx="2"/>
						<path d="M9 12h6M12 9v6"/>
					</svg>
				</div>
				<span class="amenity-label">Secure Parking</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
					</svg>
				</div>
				<span class="amenity-label">Room Service</span>
			</div>
			<div class="amenity" data-stagger-item>
				<div class="amenity-icon">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
						<path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07 19.5 19.5 0 01-6-6 19.79 19.79 0 01-3.07-8.67A2 2 0 014.11 2h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L8.09 9.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 16.92z"/>
					</svg>
				</div>
				<span class="amenity-label">Concierge</span>
			</div>
		</div>
	</section>

	<!-- ═══════════════ TESTIMONIAL ═══════════════ -->
	<section id="testimonial" data-reveal class="testimonial">
		<div class="testimonial-marks">
			<svg width="32" height="32" viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="1">
				<path d="M10 16H6a4 4 0 014-4v4m6 0h-4a4 4 0 014-4v4" stroke-linecap="round"/>
			</svg>
		</div>
		<blockquote class="testimonial-quote">
			"The Lobby restored something in me that I didn't know was broken. I left as a different person."
		</blockquote>
		<cite class="testimonial-cite">— Sarah Mitchell, Travel + Leisure</cite>
	</section>

	<!-- ═══════════════ CTA ═══════════════ -->
	<section id="cta" data-reveal class="cta">
		<div class="cta-bg">
			<img src="https://images.unsplash.com/photo-1582719508461-905c673771fd?w=1600&q=80" alt="Resort view" loading="lazy" />
		</div>
		<div class="cta-overlay"></div>
		<div class="cta-content">
			<h2 class="cta-title" data-word-stagger>
				{#each splitWords('Ready for your stay?') as { word, key }}
					<span data-word class="title-word">{word}</span>
				{/each}
			</h2>
			<a href="/rooms" class="btn-primary">Book Now</a>
		</div>
	</section>
</div>

<style>
	/* ─── DESIGN TOKENS ─── */
	:root {
		--ease-out-expo: cubic-bezier(0.16, 1, 0.3, 1);
		--ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
	}

	/* ─── HERO ─── */
	.hero {
		position: relative;
		height: 100vh;
		min-height: 700px;
		display: flex;
		align-items: flex-end;
		background: var(--color-stone-900, #1A1714);
		overflow: hidden;
	}

	.hero-slide {
		position: absolute;
		inset: 0;
		background-size: cover;
		background-position: center;
		opacity: 0;
		transform: scale(1.08);
		transition: opacity 1.4s var(--ease-in-out), transform 7s var(--ease-out-expo);
		will-change: transform, opacity;
	}

	.hero-slide.active {
		opacity: 1;
		transform: scale(1);
	}

	.hero-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to bottom,
			rgba(26, 23, 20, 0.3) 0%,
			rgba(26, 23, 20, 0.15) 40%,
			rgba(26, 23, 20, 0.65) 100%
		);
		z-index: 1;
	}

	.hero-indicators {
		position: absolute;
		bottom: 2rem;
		left: 3rem;
		display: flex;
		gap: 0.5rem;
		z-index: 10;
	}

	.hero-dot {
		width: 2.5rem;
		height: 1px;
		background: rgba(255, 255, 255, 0.2);
		border: none;
		cursor: pointer;
		padding: 0;
		transition: background 0.6s var(--ease-out-expo);
		position: relative;
	}

	.hero-dot.active {
		background: rgba(255, 255, 255, 0.8);
	}

	.hero-dot::after {
		content: '';
		position: absolute;
		top: -6px;
		left: 0;
		right: 0;
		bottom: -6px;
	}

	.hero-controls {
		position: absolute;
		bottom: 1.5rem;
		right: 3rem;
		display: flex;
		gap: 0.375rem;
		z-index: 10;
	}

	.hero-control {
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(255, 255, 255, 0.06);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		color: #fff;
		cursor: pointer;
		transition: all 0.25s var(--ease-out-expo);
	}

	.hero-content {
		position: relative;
		z-index: 2;
		width: 100%;
		max-width: 80rem;
		margin: 0 auto;
		padding: 0 1.5rem 5rem;
	}

	.hero-tag {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: rgba(255, 255, 255, 0.75);
		margin-bottom: 1.25rem;
		animation: heroFadeUp 0.7s var(--ease-out-expo) 0.3s both;
		text-shadow: 0 1px 6px rgba(0, 0, 0, 0.3);
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.hero-tag::after {
		content: '';
		width: 2.5rem;
		height: 1px;
		background: var(--color-brass-400, #B8A475);
		opacity: 0.7;
	}

	.hero-title {
		font-family: var(--font-display);
		font-size: clamp(3rem, 6vw, 5rem);
		font-weight: 400;
		line-height: 1.1;
		color: #fff;
		margin-bottom: 1.5rem;
		display: flex;
		flex-wrap: wrap;
		gap: 0.15em;
		text-shadow: 0 2px 12px rgba(0, 0, 0, 0.25);
	}

	.hero-word {
		opacity: 0;
		transform: translateY(24px);
		animation: heroFadeUp 0.6s var(--ease-out-expo) forwards;
	}

	.hero-desc {
		font-size: 1rem;
		line-height: 1.75;
		color: rgba(255, 255, 255, 0.8);
		max-width: 28rem;
		margin-bottom: 2.5rem;
		animation: heroFadeUp 0.7s var(--ease-out-expo) 0.9s both;
		text-shadow: 0 1px 8px rgba(0, 0, 0, 0.3);
	}

	.hero-btns {
		display: flex;
		gap: 1rem;
		animation: heroFadeUp 0.7s var(--ease-out-expo) 1.1s both;
	}

	.hero-scroll-hint {
		position: absolute;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 10;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		opacity: 0;
		animation: heroFadeIn 0.6s var(--ease-out-expo) 1.8s both;
	}

	.scroll-hint-text {
		font-size: 0.6rem;
		font-weight: 600;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: rgba(255, 255, 255, 0.5);
		text-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
	}

	.scroll-hint-line {
		width: 1px;
		height: 32px;
		background: linear-gradient(to bottom, rgba(255, 255, 255, 0.3), transparent);
		animation: scrollPulse 2.5s var(--ease-in-out) infinite;
	}

	@keyframes scrollPulse {
		0%, 100% { transform: scaleY(0.5); opacity: 0.3; }
		50% { transform: scaleY(1); opacity: 1; }
	}

	.btn-primary {
		display: inline-block;
		padding: 1rem 2.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		background: #fff;
		color: var(--color-ink, #1B1917);
		text-decoration: none;
		transition: all 0.25s var(--ease-out-expo);
		-webkit-tap-highlight-color: transparent;
	}

	.btn-ghost {
		display: inline-block;
		padding: 1rem 2.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		border: 1px solid rgba(255, 255, 255, 0.3);
		color: #fff;
		text-decoration: none;
		transition: all 0.25s var(--ease-out-expo);
		-webkit-tap-highlight-color: transparent;
	}

	/* ─── MARQUEE ─── */
	.marquee-section {
		overflow: hidden;
		padding: 1.5rem 0;
		border-top: 1px solid var(--color-stone-100, #F0EEEA);
		border-bottom: 1px solid var(--color-stone-100, #F0EEEA);
		background: var(--color-cream, #FAF8F5);
	}

	.marquee-track {
		display: flex;
		width: fit-content;
		animation: marqueeScroll 40s linear infinite;
		will-change: transform;
	}

	.marquee-section:hover .marquee-track {
		animation-play-state: paused;
	}

	.marquee-content {
		display: flex;
		align-items: center;
		gap: 3rem;
		padding: 0 1.5rem;
	}

	.marquee-item {
		font-family: var(--font-display);
		font-size: 0.95rem;
		font-weight: 400;
		font-style: italic;
		letter-spacing: 0.05em;
		color: var(--color-stone-400, #A9A296);
		white-space: nowrap;
		transition: color 0.25s var(--ease-out-expo);
		display: flex;
		align-items: center;
		gap: 3rem;
	}

	.marquee-item:not(:last-child)::after {
		content: '✦';
		font-size: 0.55rem;
		color: var(--color-brass-400, #B8A475);
		opacity: 0.5;
		font-style: normal;
	}

	@keyframes marqueeScroll {
		from { transform: translateX(0); }
		to { transform: translateX(-50%); }
	}

	/* ─── SECTIONS ─── */
	.section {
		padding: 6rem 1.5rem;
		max-width: 80rem;
		margin: 0 auto;
	}

	@media (min-width: 640px) {
		.section { padding: 6rem 3rem; }
	}

	.section-tag {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 1rem;
	}

	.section-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 3rem);
		font-weight: 300;
		line-height: 1.12;
		margin-bottom: 3rem;
		display: flex;
		flex-wrap: wrap;
		gap: 0.2em;
	}

	.title-word {
		display: inline-block;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		margin-bottom: 3rem;
	}

	.section-link {
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-ink, #1B1917);
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		transition: gap 0.3s var(--ease-out-expo);
	}

	.section-link::after {
		content: '→';
		font-size: 0.85rem;
		transition: transform 0.3s var(--ease-out-expo);
	}

	.section-link:hover {
		gap: 0.75rem;
	}

	.section-link:hover::after {
		transform: translateX(2px);
	}

	/* ─── ROOMS ─── */
	.rooms-track {
		display: flex;
		gap: 1.5rem;
		overflow-x: auto;
		scroll-snap-type: x mandatory;
		-webkit-overflow-scrolling: touch;
		scrollbar-width: none;
		padding-bottom: 1rem;
	}

	.rooms-track::-webkit-scrollbar {
		display: none;
	}

	.room-card {
		flex: 0 0 85vw;
		scroll-snap-align: start;
		position: relative;
		overflow: hidden;
		cursor: pointer;
		text-decoration: none;
		color: inherit;
	}

	@media (min-width: 640px) {
		.room-card { flex: 0 0 50vw; }
	}

	@media (min-width: 1024px) {
		.room-card { flex: 0 0 calc(33.333% - 1rem); }
	}

	.room-card-img {
		aspect-ratio: 3 / 4;
		overflow: hidden;
		background: var(--color-stone-100, #F0EEEA);
		position: relative;
	}

	.room-card-img img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.7s var(--ease-out-expo);
		will-change: transform;
	}

	.room-card-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(to top, rgba(0, 0, 0, 0.5) 0%, transparent 60%);
		opacity: 0;
		transition: opacity 0.4s var(--ease-out-expo);
	}

	.room-card-info {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		padding: 2rem;
		background: linear-gradient(to top, rgba(0, 0, 0, 0.6), transparent);
		transform: translateY(4px);
		transition: transform 0.35s var(--ease-out-expo);
	}

	.room-card-name {
		font-family: var(--font-display);
		font-size: 1.5rem;
		font-weight: 400;
		color: #fff;
		margin-bottom: 0.3rem;
	}

	.room-card-price {
		font-size: 0.75rem;
		color: var(--color-brass-300, #D4C4A0);
		letter-spacing: 0.05em;
	}

	/* ─── ABOUT ─── */
	.about {
		display: grid;
		grid-template-columns: 1fr;
		gap: 0;
	}

	@media (min-width: 768px) {
		.about { grid-template-columns: 1fr 1fr; }
	}

	.about-img {
		overflow: hidden;
		aspect-ratio: 3 / 4;
	}

	@media (min-width: 768px) {
		.about-img { aspect-ratio: auto; }
	}

	.about-img img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.8s var(--ease-out-expo);
	}

	.about-body {
		display: flex;
		flex-direction: column;
		justify-content: center;
		padding: 4rem 1.5rem;
	}

	@media (min-width: 640px) {
		.about-body { padding: 4rem 3rem; }
	}

	@media (min-width: 768px) {
		.about-body { padding: 4rem; }
	}

	.about-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3vw, 2.8rem);
		font-weight: 300;
		line-height: 1.15;
		margin-bottom: 1.5rem;
		display: flex;
		flex-wrap: wrap;
		gap: 0.2em;
	}

	.about-text {
		font-size: 0.95rem;
		line-height: 1.8;
		color: var(--color-stone-500, #857E72);
		margin-bottom: 2rem;
		max-width: 28rem;
	}

	.text-link {
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-ink, #1B1917);
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		width: fit-content;
		transition: gap 0.3s var(--ease-out-expo);
	}

	.text-link::after {
		content: '→';
		font-size: 0.85rem;
		transition: transform 0.3s var(--ease-out-expo);
	}

	.text-link:hover {
		gap: 0.75rem;
	}

	.text-link:hover::after {
		transform: translateX(2px);
	}

	/* ─── AMENITIES ─── */
	.amenities {
		background: var(--color-stone-900, #1A1714);
		color: #fff;
		padding: 6rem 1.5rem;
		text-align: center;
	}

	@media (min-width: 640px) {
		.amenities { padding: 6rem 3rem; }
	}

	.amenities .section-title {
		justify-content: center;
	}

	.amenities-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1px;
		background: rgba(255, 255, 255, 0.06);
		max-width: 48rem;
		margin: 0 auto;
	}

	@media (min-width: 640px) {
		.amenities-grid { grid-template-columns: repeat(4, 1fr); }
	}

	.amenity {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 2.5rem 1rem;
		background: var(--color-stone-900, #1A1714);
		transition: background 0.3s var(--ease-out-expo);
		-webkit-tap-highlight-color: transparent;
	}

	.amenity-icon {
		width: 2rem;
		height: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		color: rgba(255, 255, 255, 0.15);
		margin-bottom: 0.25rem;
		transition: color 0.3s var(--ease-out-expo);
	}

	.amenity-label {
		font-size: 0.7rem;
		font-weight: 500;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-300, #D1CCC3);
	}

	/* ─── TESTIMONIAL ─── */
	.testimonial {
		padding: 6rem 1.5rem;
		text-align: center;
		background: var(--color-cream, #FAF8F5);
		position: relative;
	}

	@media (min-width: 640px) {
		.testimonial { padding: 6rem 3rem; }
	}

	.testimonial-marks {
		color: var(--color-brass-400, #B8A475);
		margin-bottom: 1.5rem;
		display: flex;
		justify-content: center;
		opacity: 0.6;
	}

	.testimonial-quote {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.8vw, 2.25rem);
		font-weight: 300;
		font-style: italic;
		line-height: 1.45;
		max-width: 40rem;
		margin: 0 auto 1.5rem;
		color: var(--color-ink, #1B1917);
		position: relative;
		padding: 0 2rem;
	}

	.testimonial-quote::before,
	.testimonial-quote::after {
		content: '';
		position: absolute;
		top: 0;
		bottom: 0;
		width: 1px;
		background: var(--color-brass-400, #B8A475);
		opacity: 0.3;
	}
	.testimonial-quote::before { left: 0; }
	.testimonial-quote::after { right: 0; }

	.testimonial-cite {
		font-style: normal;
		font-size: 0.72rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
	}

	/* ─── CTA ─── */
	.cta {
		position: relative;
		height: 60vh;
		min-height: 400px;
		display: flex;
		align-items: center;
		justify-content: center;
		text-align: center;
		overflow: hidden;
	}

	.cta-bg {
		position: absolute;
		inset: 0;
	}

	.cta-bg img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.8s var(--ease-out-expo);
	}

	.cta-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to bottom,
			rgba(26, 23, 20, 0.25),
			rgba(26, 23, 20, 0.55)
		);
	}

	.cta-content {
		position: relative;
		z-index: 2;
		color: #fff;
	}

	.cta-title {
		font-family: var(--font-display);
		font-size: clamp(2.5rem, 5vw, 4rem);
		font-weight: 300;
		margin-bottom: 2rem;
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 0.2em;
	}

	/* ─── MOBILE RESPONSIVE ─── */
	@media (max-width: 639px) {
		.hero {
			min-height: 100dvh;
			min-height: 100svh;
		}

		.hero-content {
			padding: 0 1.25rem 6rem;
		}

		.hero-title {
			gap: 0;
			font-size: clamp(2.6rem, 9vw, 3rem);
			line-height: 1.15;
			font-weight: 400;
			margin-bottom: 1rem;
		}

		.hero-indicators {
			bottom: 1.25rem;
			left: 1.5rem;
		}

		.hero-dot {
			width: 2rem;
		}

		.hero-controls {
			bottom: 1rem;
			right: 1.5rem;
		}

		.hero-control {
			width: 36px;
			height: 36px;
		}

		.hero-scroll-hint {
			display: none;
		}

		.hero-btns {
			flex-direction: column;
			gap: 0.75rem;
		}

		.hero-btns .btn-primary,
		.hero-btns .btn-ghost {
			width: 100%;
			text-align: center;
			padding: 1rem 1.5rem;
		}

		.section-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.section-title {
			margin-bottom: 2rem;
		}

		.room-card {
			flex: 0 0 55vw;
		}

		.room-card-info {
			padding: 1.5rem;
		}

		.about-body {
			padding: 3rem 1.25rem;
		}

		.amenities {
			padding: 4rem 1.25rem;
		}

		.amenity {
			padding: 2rem 0.75rem;
		}

		.amenity-icon {
			width: 2.5rem;
			height: 2.5rem;
		}

		.amenity-icon svg {
			width: 24px;
			height: 24px;
		}

		.testimonial {
			padding: 4rem 1.5rem;
		}

		.cta {
			min-height: 350px;
		}

		.marquee-content {
			gap: 2.5rem;
			padding: 0 1.25rem;
		}

		.marquee-item {
			font-size: 0.75rem;
		}
	}

	@media (max-width: 374px) {
		.hero-title {
			font-size: clamp(2.2rem, 8vw, 3rem);
		}

		.amenities-grid {
			grid-template-columns: 1fr 1fr;
		}

		.amenity {
			padding: 1.5rem 0.5rem;
		}
	}

	/* ─── KEYFRAMES ─── */
	@keyframes heroFadeUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes heroFadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	/* ─── TOUCH DEVICE HOVER GUARD ─── */
	@media (hover: hover) {
		.hero-control:hover {
			background: rgba(255, 255, 255, 0.14);
			border-color: rgba(255, 255, 255, 0.25);
			transform: scale(1.05);
		}

		.room-card:hover .room-card-img img {
			transform: scale(1.05);
		}

		.room-card:hover .room-card-overlay {
			opacity: 1;
		}

		.room-card:hover .room-card-info {
			transform: translateY(0);
		}

		.about:hover .about-img img {
			transform: scale(1.03);
		}

		.cta:hover .cta-bg img {
			transform: scale(1.04);
		}

		.amenity:hover {
			background: rgba(255, 255, 255, 0.04);
		}

		.amenity:hover .amenity-icon {
			color: rgba(255, 255, 255, 0.4);
		}

		.btn-primary:hover {
			background: var(--color-stone-200, #E4E1DB);
			transform: translateY(-2px);
			box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
		}

		.btn-ghost:hover {
			border-color: #fff;
			background: rgba(255, 255, 255, 0.08);
			transform: translateY(-2px);
		}

		.marquee-item:hover {
			color: var(--color-ink, #1B1917);
		}
	}

	/* ─── ACTIVE STATE FOR TOUCH ─── */
	@media (hover: none) {
		.btn-primary:active {
			background: var(--color-stone-200, #E4E1DB);
		}

		.btn-ghost:active {
			border-color: #fff;
			background: rgba(255, 255, 255, 0.08);
		}

		.room-card:active .room-card-img img {
			transform: scale(1.05);
		}

		.room-card:active .room-card-overlay {
			opacity: 1;
		}
	}

	/* ─── REDUCED MOTION ─── */
	@media (prefers-reduced-motion: reduce) {
		.hero-slide {
			transition: opacity 0.3s;
			transform: none;
		}

		.hero-slide.active {
			transform: none;
		}

		.hero-tag,
		.hero-desc,
		.hero-btns {
			animation: heroFadeIn 0.3s both;
		}

		.hero-word {
			animation: heroFadeIn 0.3s forwards;
			font-weight: 400;
		}

		.hero-scroll-hint {
			display: none;
		}

		.marquee-track {
			animation: none;
		}

		.marquee-section:hover .marquee-track {
			animation-play-state: running;
		}

		.hero-control,
		.room-card-img img,
		.room-card-info,
		.room-card-overlay,
		.about-img img,
		.cta-bg img {
			transition: none;
		}
	}
</style>
