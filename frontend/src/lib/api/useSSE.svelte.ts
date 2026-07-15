import { browser } from '$app/environment';
import type { AvailabilityEvent, BookingUpdatedEvent, SSEEvent } from '$lib/types/api';

type SSEEventHandler = (event: SSEEvent) => void;

interface SSEState {
	isConnected: boolean;
	reconnectAttempts: number;
	lastEventId: string | null;
}

const MAX_RECONNECT_DELAY = 30000;
const BASE_RECONNECT_DELAY = 1000;

export function useSSE(getToken: () => string | null) {
	let state = $state<SSEState>({
		isConnected: false,
		reconnectAttempts: 0,
		lastEventId: null
	});

	let events = $state<SSEEvent[]>([]);
	let eventSource: EventSource | null = null;
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	let handlers: SSEEventHandler[] = [];

	function getReconnectDelay(): number {
		return Math.min(
			BASE_RECONNECT_DELAY * Math.pow(2, state.reconnectAttempts),
			MAX_RECONNECT_DELAY
		);
	}

	function connect() {
		if (!browser || eventSource) return;

		const token = getToken();
		const base = import.meta.env?.VITE_API_URL ?? 'http://localhost:8000';
		let url = `${base}/api/events`;
		if (token) {
			url += `?token=${encodeURIComponent(token)}`;
		}

		eventSource = new EventSource(url);

		eventSource.onopen = () => {
			state.isConnected = true;
			state.reconnectAttempts = 0;
		};

		eventSource.addEventListener('availability', (e) => {
			try {
				const data: AvailabilityEvent = { type: 'availability', ...JSON.parse(e.data) };
				events.push(data);
				if (e.lastEventId) state.lastEventId = e.lastEventId;
				handlers.forEach((h) => h(data));
			} catch {
				// ignore malformed events
			}
		});

		eventSource.addEventListener('booking-updated', (e) => {
			try {
				const data: BookingUpdatedEvent = { type: 'booking-updated', ...JSON.parse(e.data) };
				events.push(data);
				if (e.lastEventId) state.lastEventId = e.lastEventId;
				handlers.forEach((h) => h(data));
			} catch {
				// ignore malformed events
			}
		});

		eventSource.onerror = () => {
			state.isConnected = false;
			eventSource?.close();
			eventSource = null;

			// Reconnect with exponential backoff
			const delay = getReconnectDelay();
			state.reconnectAttempts++;
			reconnectTimer = setTimeout(connect, delay);
		};
	}

	function disconnect() {
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
			reconnectTimer = null;
		}
		if (eventSource) {
			eventSource.close();
			eventSource = null;
		}
		state.isConnected = false;
	}

	// Auto-connect on mount, disconnect on unmount
	$effect(() => {
		connect();
		return () => disconnect();
	});

	return {
		get isConnected() {
			return state.isConnected;
		},
		get reconnectAttempts() {
			return state.reconnectAttempts;
		},
		get events() {
			return events;
		},
		get lastEventId() {
			return state.lastEventId;
		},
		onEvent(handler: SSEEventHandler) {
			handlers.push(handler);
			return () => {
				handlers = handlers.filter((h) => h !== handler);
			};
		},
		reconnect() {
			disconnect();
			state.reconnectAttempts = 0;
			connect();
		},
		disconnect
	};
}
