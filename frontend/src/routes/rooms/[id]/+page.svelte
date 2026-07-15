<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { roomsApi } from '$lib/api/client';
	import type { RoomWithImages, AvailabilityResult } from '$lib/types/api';
	import { useSSE } from '$lib/api/useSSE.svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';

	let roomData = $state<RoomWithImages | null>(null);
	let loading = $state(true);
	let error = $state('');
	let checkIn = $state('');
	let checkOut = $state('');
	let bookingType = $state<'daily' | 'hourly'>('daily');
	let availability = $state<AvailabilityResult | null>(null);
	let checkingAvailability = $state(false);

	const roomId = $derived(page.params.id);

	const fallbackImages = [
		'https://images.unsplash.com/photo-1631049307264-da0ec9d70304?w=800&q=80',
		'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=800&q=80',
		'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=800&q=80'
	];

	function formatDate(d: string): string {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function getToday(): string {
		return new Date().toISOString().split('T')[0];
	}

	async function checkAvail() {
		if (!checkIn || !checkOut || !roomId) {
			availability = null;
			return;
		}
		checkingAvailability = true;
		try {
			const result = await roomsApi.checkAvailability(roomId, {
				check_in: checkIn,
				check_out: checkOut,
				type: bookingType
			});
			availability = result;
		} catch {
			availability = null;
		} finally {
			checkingAvailability = false;
		}
	}

	let showMobileBar = $state(false);
	$effect(() => {
		showMobileBar = !!availability?.available;
	});

	let bookingNights = $derived.by(() => {
		if (!checkIn || !checkOut) return 0;
		const start = new Date(checkIn);
		const end = new Date(checkOut);
		return Math.max(0, Math.ceil((end.getTime() - start.getTime()) / (86_400_000)));
	});

	let priceBreakdown = $derived.by(() => {
		if (!availability?.available || !availability.total_price) return null;
		const total = availability.total_price;
		const rate = bookingType === 'daily' ? roomData?.base_rate_daily ?? 0 : roomData?.base_rate_hourly ?? 0;
		const nights = bookingNights || 1;
		return { rate, nights, total };
	});

	// Auto-check availability whenever dates or booking type change
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;
	$effect(() => {
		checkIn;
		checkOut;
		bookingType;
		if (debounceTimer) clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => checkAvail(), 600);
	});

	// Connect to SSE for real-time availability pushes
	const sse = useSSE(() => auth.getToken());
	$effect(() => {
		const unsub = sse.onEvent((event) => {
			if (event.type === 'availability' && roomId) {
				const ae = event as import('$lib/types/api').AvailabilityEvent;
				const thisRoom = ae.rooms.find(r => r.room_id === roomId);
				if (thisRoom) {
					roomsApi.getWithImages(roomId).then((data) => {
						roomData = data;
					}).catch(() => {});
					if (checkIn && checkOut) {
						checkAvail();
					}
				}
			}
		});
		return () => unsub();
	});

	onMount(() => {
		if (!roomId) return;
		roomsApi.getWithImages(roomId)
			.then((data) => {
				roomData = data;
			})
			.catch(() => {
				error = 'Room not found';
				toast.error('Room not found', 'This room may no longer be available');
			})
			.finally(() => {
				loading = false;
			});
	});
</script>

<svelte:head>
	<title>{roomData?.room.room_number ?? 'Room'} — The Lobby</title>
</svelte:head>

{#if loading}
	<div class="page">
		<Skeleton width="6rem" height="0.75rem" />
		<div class="room-layout" style="margin-top: 1.5rem;">
			<div class="gallery">
				<div class="gallery-item"><Skeleton width="100%" height="100%" /></div>
				<div class="gallery-item"><Skeleton width="100%" height="100%" /></div>
				<div class="gallery-item"><Skeleton width="100%" height="100%" /></div>
				<div class="gallery-item"><Skeleton width="100%" height="100%" /></div>
			</div>
			<div>
				<Skeleton width="40%" height="2.5rem" />
				<div style="margin-top: 1rem;">
					<Skeleton width="30%" height="0.9rem" />
				</div>
				<div style="margin-top: 1.5rem; display: flex; gap: 2rem;">
					<div><Skeleton width="5rem" height="1.5rem" /></div>
					<div><Skeleton width="5rem" height="1.5rem" /></div>
				</div>
			</div>
		</div>
	</div>
{:else if error}
	<div class="error-page">
		<h1>Room not found</h1>
		<a href="/rooms" class="back-link">Back to Rooms</a>
	</div>
{:else if roomData}
	{@const room = roomData.room}
	{@const images = roomData.images}

	<div class="page">
		<a href="/rooms" class="back-link">Back to Rooms</a>

		<div class="room-layout">
			<!-- Image gallery -->
			<div class="gallery">
				{#if images.length > 0}
					{#each images.slice(0, 4) as img}
						<div class="gallery-item">
							<img src={img.url} alt="Room {room.room_number}" loading="lazy" />
						</div>
					{/each}
				{:else}
					{#each fallbackImages.slice(0, 2) as src}
						<div class="gallery-item">
							<img {src} alt="Room {room.room_number}" loading="lazy" />
						</div>
					{/each}
				{/if}
			</div>

			<!-- Info -->
			<div class="room-info">
				<div class="room-meta">
					<span class="room-status" class:available={room.upcoming_bookings === 0 && room.status === 'active'} class:maintenance={room.status === 'maintenance'}>
						{room.upcoming_bookings > 0 ? `${room.upcoming_bookings} upcoming booking${room.upcoming_bookings > 1 ? 's' : ''}` : room.status === 'active' ? 'Available' : room.status}
					</span>
				</div>

				<h1 class="room-title">Room {room.room_number}</h1>
				<p class="room-type-name">{roomData.room_type_name}</p>

				<div class="room-pricing">
					<div class="price-block">
						<span class="price-label">Daily Rate</span>
						<span class="price-value">${roomData.base_rate_daily.toFixed(2)}</span>
					</div>
					<div class="price-block">
						<span class="price-label">Hourly Rate</span>
						<span class="price-value">${roomData.base_rate_hourly.toFixed(2)}</span>
					</div>
				</div>

				<!-- Availability check -->
				<div class="avail-section">
					<h2 class="avail-title">Check Availability</h2>

					<div class="avail-form">
						<div class="toggle-row">
							<button
								class="toggle-btn"
								class:active={bookingType === 'daily'}
								onclick={() => (bookingType = 'daily')}
							>
								Daily
							</button>
							<button
								class="toggle-btn"
								class:active={bookingType === 'hourly'}
								onclick={() => (bookingType = 'hourly')}
							>
								Hourly
							</button>
						</div>

						<div class="input-row">
							<div class="input-group">
								<label for="checkin">Check In</label>
								<input
									id="checkin"
									type="date"
									bind:value={checkIn}
									min={getToday()}
								/>
							</div>
							<div class="input-group">
								<label for="checkout">Check Out</label>
								<input
									id="checkout"
									type="date"
									bind:value={checkOut}
									min={checkIn || getToday()}
								/>
							</div>
						</div>

						{#if checkingAvailability}
							<div class="checking-msg">Checking availability...</div>
						{/if}
					</div>

					{#if availability}
						<div class="avail-result">
							<div class="avail-count">
								{#if availability.available}
									Available for selected dates
									{#if availability.available_rooms === 1}
										<div class="urgency-note">Only 1 room left at this rate!</div>
									{/if}
								{:else if availability.available_rooms > 0}
									This specific room is booked, but other {roomData?.room_type_name || 'rooms of this type'} are available.
								{:else}
									Not available for selected dates
								{/if}
							</div>
							{#if availability.available}
								{#if priceBreakdown}
									<div class="price-breakdown">
										<div class="breakdown-line">
											<span>${priceBreakdown.rate.toFixed(2)}/{bookingType === 'daily' ? 'night' : 'hr'} &times; {priceBreakdown.nights} {bookingType === 'daily' ? 'nights' : 'hours'}</span>
											<span>= ${priceBreakdown.total.toFixed(2)}</span>
										</div>
									</div>
								{/if}
								<div class="avail-price">
									Total: <strong>${availability.total_price}</strong>
								</div>
								<a
									href="/booking?room_id={roomId}&check_in={checkIn}&check_out={checkOut}&type={bookingType}"
									class="book-btn"
								>
									Book Now
								</a>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>

	<!-- Mobile sticky booking bar -->
	{#if showMobileBar}
		<div class="mobile-sticky-bar">
			<div class="mobile-sticky-price">
				<strong>${availability?.total_price ?? ''}</strong>
				<span>total</span>
			</div>
			<a
				href="/booking?room_id={roomId}&check_in={checkIn}&check_out={checkOut}&type={bookingType}"
				class="mobile-sticky-btn"
			>
				Book Now
			</a>
		</div>
	{/if}
{/if}

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

	.room-layout {
		display: grid;
		grid-template-columns: 1fr;
		gap: 3rem;
	}

	@media (min-width: 768px) {
		.room-layout { grid-template-columns: 1.2fr 0.8fr; }
	}

	.gallery {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;
	}

	.gallery-item {
		aspect-ratio: 4 / 3;
		overflow: hidden;
		background: var(--color-stone-100, #F0EEEA);
	}

	.gallery-item:first-child {
		grid-column: 1 / -1;
		aspect-ratio: 16 / 9;
	}

	.gallery-item img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.room-info {
		display: flex;
		flex-direction: column;
	}

	.room-meta {
		margin-bottom: 1rem;
	}

	.room-status {
		display: inline-block;
		padding: 0.3rem 0.7rem;
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		background: var(--color-stone-200, #E4E1DB);
		color: var(--color-stone-600, #6B645A);
	}

	.room-status.available {
		background: rgba(90, 122, 82, 0.15);
		color: var(--color-sage-700, #40416C);
	}

	.room-status.maintenance {
		background: rgba(180, 100, 60, 0.15);
		color: #9b5a30;
	}

	.room-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 3rem);
		font-weight: 300;
		line-height: 1.1;
		margin-bottom: 2rem;
	}

	.room-pricing {
		display: flex;
		gap: 2rem;
		margin-bottom: 3rem;
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.price-block {
		display: flex;
		flex-direction: column;
	}

	.price-label {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.3rem;
	}

	.price-value {
		font-family: var(--font-display);
		font-size: 1.5rem;
		font-weight: 400;
	}

	.price-breakdown {
		margin-bottom: 0.75rem;
	}

	.breakdown-line {
		display: flex;
		justify-content: space-between;
		font-size: 0.8rem;
		color: var(--color-stone-500, #857E72);
		padding: 0.35rem 0;
		border-bottom: 1px dashed var(--color-stone-200, #E4E1DB);
	}

	.urgency-note {
		margin-top: 0.4rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: #b45309;
	}

	.avail-section {
		margin-top: auto;
	}

	.avail-title {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
		margin-bottom: 1.25rem;
	}

	.avail-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.toggle-row {
		display: flex;
		gap: 0;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.toggle-btn {
		flex: 1;
		padding: 0.7rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		background: transparent;
		border: none;
		color: var(--color-stone-500, #857E72);
		cursor: pointer;
		transition: all 0.25s;
	}

	.toggle-btn.active {
		background: var(--color-ink, #1B1917);
		color: #fff;
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
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.4rem;
	}

	.input-group input {
		padding: 0.7rem 0.75rem;
		font-family: var(--font-body);
		font-size: 0.85rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		background: transparent;
		color: var(--color-ink, #1B1917);
		outline: none;
		transition: border-color 0.2s;
	}

	.input-group input:focus {
		border-color: var(--color-ink, #1B1917);
	}

	.checking-msg {
		padding: 0.85rem 0;
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--color-stone-500, #857E72);
		text-align: center;
	}

	.avail-result {
		margin-top: 1.25rem;
		padding: 1.25rem;
		border: 1px solid var(--color-stone-200, #E4E1DB);
	}

	.avail-count {
		font-size: 0.85rem;
		margin-bottom: 0.5rem;
	}

	.avail-price {
		font-size: 0.9rem;
		color: var(--color-stone-500, #857E72);
		margin-bottom: 1rem;
	}

	.avail-price strong {
		font-family: var(--font-display);
		font-size: 1.2rem;
		color: var(--color-ink, #1B1917);
	}

	.book-btn {
		display: block;
		width: 100%;
		padding: 0.85rem;
		text-align: center;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		background: var(--color-sage-700, #40416C);
		color: #fff;
		text-decoration: none;
		transition: opacity 0.2s;
	}

	.book-btn:hover {
		opacity: 0.85;
	}

	/* Error */
	.error-page {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 60vh;
		padding: 2rem;
		text-align: center;
	}

	.error-page h1 {
		font-family: var(--font-display);
		font-size: 2rem;
		font-weight: 300;
		margin-bottom: 1rem;
	}

	@keyframes pulse {
		0%, 100% { opacity: 0.4; }
		50% { opacity: 0.8; }
	}

	@media (max-width: 639px) {
		.page {
			padding: 4rem 1.25rem 6rem;
		}

		.back-link {
			font-size: 0.7rem;
			margin-bottom: 1.5rem;
		}

		.room-layout {
			gap: 2rem;
		}

		.room-title {
			font-size: clamp(1.6rem, 5vw, 2rem);
			margin-bottom: 1.5rem;
		}

		.room-pricing {
			gap: 1.5rem;
			margin-bottom: 2rem;
			padding-bottom: 1.5rem;
		}

		.price-value {
			font-size: 1.25rem;
		}

		.breakdown-line {
			font-size: 0.75rem;
			flex-direction: column;
			align-items: flex-start;
			gap: 0.2rem;
		}

		.avail-title {
			font-size: 1.15rem;
		}

		.input-row {
			grid-template-columns: 1fr;
			gap: 0.75rem;
		}

		.toggle-btn {
			padding: 0.65rem;
			font-size: 0.7rem;
		}

		.input-group input {
			font-size: 0.9rem;
			padding: 0.8rem 0.75rem;
		}

		.avail-result {
			margin-top: 1rem;
			padding: 1rem;
		}

		.book-btn {
			font-size: 0.7rem;
			padding: 0.8rem;
		}
	}

	/* Mobile sticky booking bar */
	.mobile-sticky-bar {
		display: none;
	}

	@media (max-width: 639px) {
		.mobile-sticky-bar {
			display: flex;
			position: fixed;
			bottom: 0;
			left: 0;
			right: 0;
			z-index: 30;
			align-items: center;
			justify-content: space-between;
			padding: 0.75rem 1.25rem;
			background: var(--color-cream, #FAF8F5);
			border-top: 1px solid var(--color-stone-200, #E4E1DB);
			box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.06);
		}

		.mobile-sticky-price {
			display: flex;
			flex-direction: column;
		}

		.mobile-sticky-price strong {
			font-family: var(--font-display);
			font-size: 1.3rem;
			font-weight: 400;
			line-height: 1.2;
		}

		.mobile-sticky-price span {
			font-size: 0.65rem;
			text-transform: uppercase;
			letter-spacing: 0.08em;
			color: var(--color-stone-500, #857E72);
		}

		.mobile-sticky-btn {
			padding: 0.75rem 2rem;
			font-size: 0.7rem;
			font-weight: 600;
			letter-spacing: 0.1em;
			text-transform: uppercase;
			background: var(--color-sage-700, #40416C);
			color: #fff;
			text-decoration: none;
			transition: opacity 0.2s;
		}

		.mobile-sticky-btn:active {
			opacity: 0.85;
		}
	}
</style>
