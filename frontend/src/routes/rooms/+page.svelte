<script lang="ts">
	import { onMount } from 'svelte';
	import { roomTypesApi, roomsApi } from '$lib/api/client';
	import type { RoomType, Room, AvailabilityEvent } from '$lib/types/api';
	import { useSSE } from '$lib/api/useSSE.svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';

	let roomTypes = $state<RoomType[]>([]);
	let allRooms = $state<Room[]>([]);
	let selectedType = $state<string>('');
	let sortBy = $state<string>('default');
	let loading = $state(true);

	const roomImages = $state<Record<string, string>>({});
	let roomAvailability = $state<Record<string, boolean>>({});

	async function loadRooms() {
		const [types, rooms] = await Promise.all([
			roomTypesApi.list().catch(() => {
				toast.error('Failed to load room types', 'Please try again later');
				return [] as RoomType[];
			}),
			roomsApi.list().catch(() => {
				toast.error('Failed to load rooms', 'Please try again later');
				return [] as Room[];
			})
		]);
		roomTypes = types;
		allRooms = rooms;
		loading = false;
		rooms.forEach((room) => {
			roomsApi.getImages(room.id).then((imgs) => {
				const primary = imgs.find((i) => i.is_primary) ?? imgs[0];
				if (primary) roomImages[room.id] = primary.url;
			}).catch(() => {});
		});
	}

	const fallbackImages = [
		'https://images.unsplash.com/photo-1631049307264-da0ec9d70304?w=600&q=80',
		'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80',
		'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80'
	];

	let filteredRooms = $derived(
		selectedType
			? allRooms.filter((r) => r.room_type_id === selectedType)
			: allRooms
	);

	let sortedRooms = $derived.by(() => {
		const rooms = filteredRooms;
		if (sortBy === 'price-asc') {
			return [...rooms].sort((a, b) => {
				const aPrice = getRoomType(a.room_type_id)?.base_rate_daily ?? 0;
				const bPrice = getRoomType(b.room_type_id)?.base_rate_daily ?? 0;
				return aPrice - bPrice;
			});
		}
		if (sortBy === 'price-desc') {
			return [...rooms].sort((a, b) => {
				const aPrice = getRoomType(a.room_type_id)?.base_rate_daily ?? 0;
				const bPrice = getRoomType(b.room_type_id)?.base_rate_daily ?? 0;
				return bPrice - aPrice;
			});
		}
		if (sortBy === 'name') {
			return [...rooms].sort((a, b) => a.room_number.localeCompare(b.room_number));
		}
		return rooms;
	});

	let roomTypeCounts = $derived.by(() => {
		const counts: Record<string, number> = {};
		for (const r of allRooms) {
			const key = r.room_type_id;
			counts[key] = (counts[key] || 0) + 1;
		}
		return counts;
	});

	function getRoomType(typeId: string): RoomType | undefined {
		return roomTypes.find((t) => t.id === typeId);
	}

	function roomStatusText(room: Room): string {
		if (room.status === 'maintenance') return 'Maintenance';
		if (room.status === 'inactive') return 'Unavailable';
		const avail = roomAvailability[room.id];
		if (avail === false) return 'Booked';
		if (avail === true) return 'Available';
		if (room.upcoming_bookings > 0) return `${room.upcoming_bookings} booking${room.upcoming_bookings > 1 ? 's' : ''}`;
		return 'Available';
	}
	function roomStatusClass(room: Room): string {
		if (room.status === 'maintenance') return 'maintenance';
		if (room.status === 'inactive') return 'maintenance';
		const avail = roomAvailability[room.id];
		if (avail === true) return 'available';
		if (avail === false) return 'booked';
		if (room.upcoming_bookings > 0) return 'booked';
		return 'available';
	}

	// SSE — update room availability in-place from per-room payload
	const sse = useSSE(() => auth.getToken());
	$effect(() => {
		const cleanup = sse.onEvent((event) => {
			if (event.type === 'availability') {
				const ae = event as AvailabilityEvent;
				for (const r of ae.rooms) {
					roomAvailability[r.room_id] = r.available;
				}
			}
		});
		return () => cleanup();
	});

	onMount(() => {
		loadRooms();
	});
</script>

<svelte:head>
	<title>Rooms — The Lobby</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<p class="section-tag">Accommodations <span class="section-tag-line"></span></p>
		<h1 class="page-title">Our Rooms</h1>
		<p class="page-desc">Twelve rooms, each a private world of comfort and quiet elegance.</p>
	</header>

	{#if !loading}
		<div class="filters-bar">
			<div class="filters">
				<button
					class="filter-btn"
					class:active={selectedType === ''}
					onclick={() => (selectedType = '')}
				>
					All
					<span class="count-badge">{allRooms.length}</span>
				</button>
				{#each roomTypes as rt}
					<button
						class="filter-btn"
						class:active={selectedType === rt.id}
						onclick={() => (selectedType = rt.id)}
					>
						{rt.name}
						<span class="count-badge">{roomTypeCounts[rt.id] ?? 0}</span>
					</button>
				{/each}
			</div>
			<select class="sort-select" bind:value={sortBy}>
				<option value="default">Default order</option>
				<option value="price-asc">Price: Low to high</option>
				<option value="price-desc">Price: High to low</option>
				<option value="name">Room number</option>
			</select>
		</div>
	{/if}

	{#if loading}
		<div class="rooms-grid">
			{#each [1, 2, 3, 4, 5, 6] as _}
				<div class="room-card">
					<div class="room-card-img">
						<Skeleton width="100%" height="100%" />
					</div>
					<div class="room-card-body">
						<Skeleton width="40%" height="0.65rem" />
						<div style="margin-top: 0.4rem;">
							<Skeleton width="60%" height="1.3rem" />
						</div>
						<div style="margin-top: 0.5rem;">
							<Skeleton width="35%" height="0.8rem" />
						</div>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="rooms-grid">
			{#each sortedRooms as room, i}
				<a href="/rooms/{room.id}" class="room-card">
					<div class="room-card-img">
						<img
							src={roomImages[room.id] || fallbackImages[i % fallbackImages.length]}
							alt={getRoomType(room.room_type_id)?.name ?? 'Room'}
							loading="lazy"
						/>
						<div class="room-card-overlay-price">
							{getRoomType(room.room_type_id) ? `$${getRoomType(room.room_type_id)!.base_rate_daily}/day` : ''}
						</div>
						<div class="room-status" class:available={roomStatusClass(room) === 'available'} class:booked={roomStatusClass(room) === 'booked'} class:maintenance={roomStatusClass(room) === 'maintenance'}>
							{roomStatusText(room)}
						</div>
					</div>
					<div class="room-card-body">
						<div class="room-card-type">
							{getRoomType(room.room_type_id)?.name ?? 'Room'}
						</div>
						<div class="room-card-number">Room {room.room_number}</div>
						<div class="room-card-cta">View Details</div>
					</div>
				</a>
			{:else}
				<div class="empty">
					<p>No rooms available{selectedType ? ' for this type' : ''}.</p>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.page {
		max-width: 80rem;
		margin: 0 auto;
		padding: 6rem 1.5rem 4rem;
	}

	@media (min-width: 640px) {
		.page { padding: 6rem 3rem 4rem; }
	}

	.page-header {
		margin-bottom: 3rem;
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

	.page-title {
		font-family: var(--font-display);
		font-size: clamp(2.5rem, 4vw, 3.5rem);
		font-weight: 300;
		line-height: 1.1;
		margin-bottom: 1rem;
	}

	.page-desc {
		font-size: 1rem;
		line-height: 1.7;
		color: var(--color-stone-500, #857E72);
		max-width: 32rem;
	}

	.filters {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
		margin-bottom: 3rem;
	}

	.filter-btn {
		padding: 0.6rem 1.25rem;
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 500;
		letter-spacing: 0.04em;
		background: transparent;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		border-radius: var(--radius, 2px);
		color: var(--color-stone-500, #857E72);
		cursor: pointer;
		transition: all 0.25s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
	}

	.filter-btn:hover {
		border-color: var(--color-ink, #1B1917);
		color: var(--color-ink, #1B1917);
	}

	.filter-btn.active {
		background: var(--color-ink, #1B1917);
		border-color: var(--color-ink, #1B1917);
		color: #fff;
	}

	.count-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 1.2rem;
		height: 1.2rem;
		padding: 0 0.3rem;
		margin-left: 0.4rem;
		font-size: 0.6rem;
		font-weight: 700;
		border-radius: 999px;
		background: var(--color-stone-200, #E4E1DB);
		color: var(--color-stone-600, #6B655A);
		line-height: 1;
	}

	.filter-btn.active .count-badge {
		background: rgba(255, 255, 255, 0.2);
		color: #fff;
	}

	.sort-select {
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--color-stone-500, #857E72);
		background: transparent;
		border: 1px solid var(--color-stone-200, #E4E1DB);
		padding: 0.5rem 0.75rem;
		cursor: pointer;
		transition: border-color 0.2s;
		outline: none;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23857E72' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.75rem center;
		padding-right: 2rem;
	}

	.sort-select:hover,
	.sort-select:focus {
		border-color: var(--color-ink, #1B1917);
		color: var(--color-ink, #1B1917);
	}

	.rooms-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 1.5rem;
	}

	@media (min-width: 640px) {
		.rooms-grid { grid-template-columns: repeat(2, 1fr); }
	}

	@media (min-width: 1024px) {
		.rooms-grid { grid-template-columns: repeat(3, 1fr); }
	}

	.room-card {
		text-decoration: none;
		color: inherit;
		cursor: pointer;
		transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.room-card:hover {
		transform: translateY(-4px);
	}

	.room-card-img {
		position: relative;
		aspect-ratio: 3 / 4;
		overflow: hidden;
		background: var(--color-stone-100, #F0EEEA);
	}

	.room-card-img img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.6s cubic-bezier(0.16, 1, 0.3, 1);
	}

	.room-card:hover .room-card-img img {
		transform: scale(1.04);
	}

	.room-card-overlay-price {
		position: absolute;
		top: 1rem;
		left: 1rem;
		padding: 0.3rem 0.7rem;
		font-size: 0.7rem;
		font-weight: 600;
		background: var(--color-cream, #FAF8F5);
		color: var(--color-ink, #1B1917);
		border-radius: var(--radius, 2px);
		display: flex;
		align-items: baseline;
		gap: 0.15rem;
	}

	.room-card-overlay-price::before {
		content: '';
		width: 4px;
		height: 4px;
		border-radius: 999px;
		background: var(--color-brass-400, #B8A475);
	}

	.room-status {
		position: absolute;
		bottom: 1rem;
		left: 1rem;
		padding: 0.3rem 0.7rem;
		font-size: 0.6rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(8px);
		color: #fff;
		border-radius: var(--radius, 2px);
		display: flex;
		align-items: center;
		gap: 0.35rem;
	}

	.room-status::before {
		content: '';
		width: 5px;
		height: 5px;
		border-radius: 999px;
		background: currentColor;
	}

	.room-status.available {
		background: rgba(90, 122, 82, 0.8);
	}

	.room-status.booked {
		background: rgba(180, 120, 50, 0.8);
	}

	.room-status.maintenance {
		background: rgba(180, 100, 60, 0.8);
	}

	.room-card-body {
		padding: 1.25rem 0 0;
	}

	.room-card-type {
		font-size: 0.65rem;
		font-weight: 600;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--color-stone-400, #A9A296);
		margin-bottom: 0.3rem;
	}

	.room-card-number {
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
		margin-bottom: 0.5rem;
	}

	.room-card-cta {
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-ink, #1B1917);
		width: fit-content;
		transition: gap 0.3s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}

	.room-card-cta::after {
		content: '→';
		font-size: 0.85rem;
		transition: transform 0.3s var(--ease-out-expo, cubic-bezier(0.16, 1, 0.3, 1));
	}

	.room-card:hover .room-card-cta::after {
		transform: translateX(3px);
	}

	.empty {
		grid-column: 1 / -1;
		text-align: center;
		padding: 4rem 0;
		color: var(--color-stone-400, #A9A296);
	}



	@media (max-width: 639px) {
		.page {
			padding: 5rem 1.25rem 3rem;
		}

		.page-header {
			margin-bottom: 2rem;
		}

		.page-desc {
			font-size: 0.95rem;
			line-height: 1.75;
		}

		.room-card-body {
			padding: 1rem 0 0;
		}

		.room-card-type {
			font-size: 0.6rem;
		}

		.room-card-number {
			font-size: 1.15rem;
		}

		.room-card-cta {
			font-size: 0.65rem;
		}

		.room-card-overlay-price {
			font-size: 0.6rem;
			top: 0.75rem;
			left: 0.75rem;
		}

		.room-status {
			font-size: 0.55rem;
			bottom: 0.75rem;
			left: 0.75rem;
		}

		.rooms-grid {
			gap: 1.25rem;
		}

		.filter-btn {
			padding: 0.5rem 1rem;
			font-size: 0.7rem;
		}

		.filters-bar {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.75rem;
			margin-bottom: 2rem;
		}

		.filters {
			margin-bottom: 0;
			gap: 0.4rem;
		}

		.sort-select {
			font-size: 0.7rem;
			width: 100%;
		}
	}
</style>
