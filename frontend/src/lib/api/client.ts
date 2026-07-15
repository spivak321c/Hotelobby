import type { ApiResponse, ApiError } from '../types/api';
import type {
	// Auth
	LoginRequest,
	AuthResponse,
	AdminLoginRequest,
	AdminAuthResponse,
	RegisterRequest,
	// Room types
	RoomType,
	Room,
	RoomWithImages,
	RoomImage,
	AvailabilityQuery,
	AvailabilityResult,
	// Reservations
	Reservation,
	ReservationWithBookings,
	CreateReservationPayload,
	CancelReservationPayload,
	// Payments
	Payment,
	ProcessPaymentPayload,
	// Customer
	Customer,
	// Admin
	Admin,
	CreateRoomTypePayload,
	UpdateRoomTypePayload,
	CreateRoomPayload,
	UpdateRoomPayload,
	CreatePricingPayload,
	UpdatePricingPayload,
	RoomPricing,
	RoomTypeInventory,
	UpdateInventoryPayload,
	UpdateReservationStatusPayload,
	CreateAdminPayload,
	UpdateAdminPayload,
	BookingReport,
	OccupancyReport,
	RevenueReport
} from '../types/api';

// ─── Base Client ─────────────────────────────────────────────────
export class ApiException extends Error {
	public code: string;
	constructor(error: ApiError) {
		super(error.message);
		this.code = error.code;
		this.name = 'ApiException';
	}
}

interface FetchOptions extends RequestInit {
	token?: string;
	idempotencyKey?: string;
	_retried?: boolean;
}

function getBaseUrl(): string {
	return import.meta.env?.VITE_API_URL ?? '';
}

// ─── Auto-refresh on 401 ──────────────────────────────────────────
type RefreshFn = () => Promise<string | null>;
let refreshFn: RefreshFn | null = null;
let refreshing: Promise<string | null> | null = null;

export function setRefreshHandler(fn: RefreshFn) {
	refreshFn = fn;
}

async function tryRefresh(): Promise<string | null> {
	if (!refreshFn) return null;
	if (refreshing) return refreshing;
	refreshing = refreshFn().finally(() => {
		refreshing = null;
	});
	return refreshing;
}

async function request<T>(path: string, options: FetchOptions = {}): Promise<T> {
	const headers = new Headers(options.headers);
	headers.set('Content-Type', 'application/json');

	if (options.token) {
		headers.set('Authorization', `Bearer ${options.token}`);
	}
	if (options.idempotencyKey) {
		headers.set('Idempotency-Key', options.idempotencyKey);
	}

	const url = `${getBaseUrl()}${path}`;

	const doFetch = (): Promise<Response> =>
		fetch(url, { ...options, headers, body: options.body });

	let response = await doFetch();

	// If 401, attempt a single token refresh + retry
	if (response.status === 401 && options.token && !options._retried) {
		const newToken = await tryRefresh();
		if (newToken) {
			headers.set('Authorization', `Bearer ${newToken}`);
			response = await doFetch();
			options._retried = true;
		}
	}

	let body: any;
	try {
		body = await response.json();
	} catch {
		if (!response.ok) {
			throw new ApiException({ code: 'http_error', message: `HTTP ${response.status}` });
		}
		throw new ApiException({ code: 'parse_error', message: 'Failed to parse API response' });
	}

	// Handle standardized error envelope: { success: false, error: { code, message } }
	if (!response.ok) {
		if (body && typeof body === 'object' && body.error) {
			const err = body.error;
			if (typeof err === 'object' && err !== null && 'code' in err) {
				throw new ApiException({
					code: err.code || 'request_error',
					message: err.message || 'An error occurred'
				});
			}
			// Legacy fallback: error was a plain string
			throw new ApiException({ code: 'request_error', message: String(err) });
		}
		throw new ApiException({ code: 'http_error', message: `HTTP ${response.status}` });
	}

	// Handle { success, data } envelope
	if (body && typeof body === 'object' && 'success' in body) {
		if (body.success === false || body.error) {
			const err = body.error;
			if (err && typeof err === 'object' && 'code' in err) {
				throw new ApiException({
					code: err.code || 'unknown_error',
					message: err.message || 'Unknown error occurred'
				});
			}
			throw new ApiException({
				code: 'unknown_error',
				message: typeof err === 'string' ? err : 'Unknown error occurred'
			});
		}
		return body.data as T;
	}

	return body as T;
}

function get<T>(path: string, options?: FetchOptions) {
	return request<T>(path, { ...options, method: 'GET' });
}
function post<T>(path: string, body?: unknown, options?: FetchOptions) {
	return request<T>(path, { ...options, method: 'POST', body: body ? JSON.stringify(body) : undefined });
}
function put<T>(path: string, body?: unknown, options?: FetchOptions) {
	return request<T>(path, { ...options, method: 'PUT', body: body ? JSON.stringify(body) : undefined });
}
function del<T>(path: string, options?: FetchOptions) {
	return request<T>(path, { ...options, method: 'DELETE' });
}

// ─── Auth API ────────────────────────────────────────────────────
export const authApi = {
	register: (data: RegisterRequest) => post<AuthResponse>('/api/auth/register', data),
	login: (data: LoginRequest) => post<AuthResponse>('/api/auth/login', data),
	adminLogin: (data: AdminLoginRequest) => post<AdminAuthResponse>('/api/auth/admin/login', data),
	refresh: (token: string) =>
		post<{ access_token: string; refresh_token: string; user: Customer | Admin }>('/api/auth/refresh', undefined, { token }),
	logout: (token: string) => post<void>('/api/auth/logout', undefined, { token })
};

// ─── Room Types API ──────────────────────────────────────────────
export const roomTypesApi = {
	list: () => get<RoomType[]>('/api/room-types'),
	get: (id: string) => get<RoomType>(`/api/room-types/${id}`)
};

// ─── Rooms API ───────────────────────────────────────────────────
export const roomsApi = {
	list: (params?: { room_type_id?: string; status?: string }) => {
		const query = new URLSearchParams();
		if (params?.room_type_id) query.set('room_type_id', params.room_type_id);
		if (params?.status) query.set('status', params.status);
		const qs = query.toString();
		return get<Room[]>(`/api/rooms${qs ? `?${qs}` : ''}`);
	},
	get: (id: string) => get<Room>(`/api/rooms/${id}`),
	getImages: (id: string) => get<RoomImage[]>(`/api/rooms/${id}/images`),
	getWithImages: (id: string) => get<RoomWithImages>(`/api/rooms/${id}`),
	checkAvailability: (id: string, query: AvailabilityQuery) => {
		const params = new URLSearchParams({
			check_in: query.check_in,
			check_out: query.check_out,
			type: query.type
		});
		return get<AvailabilityResult>(`/api/rooms/${id}/availability?${params}`);
	}
};

// ─── Reservations API ────────────────────────────────────────────
export const reservationsApi = {
	create: (data: CreateReservationPayload, idempotencyKey?: string) =>
		post<Reservation>('/api/reservations', data, idempotencyKey ? { idempotencyKey } : undefined),
	lookup: (reference: string, email: string) => {
		const params = new URLSearchParams({ email });
		return get<ReservationWithBookings>(`/api/reservations/${reference}?${params}`);
	},
	requestCancelOTP: (reference: string, email: string) =>
		post<void>(`/api/reservations/${reference}/cancel/otp`, { email }),
	cancel: (reference: string, data: CancelReservationPayload) =>
		post<Reservation>(`/api/reservations/${reference}/cancel`, data)
};

// ─── Payments API ────────────────────────────────────────────────
export const paymentsApi = {
	process: (data: ProcessPaymentPayload) => post<Payment>('/api/payments', data),
	check: (reference: string) => get<Payment>(`/api/payments/${reference}`)
};

// ─── Customer API (auth required) ────────────────────────────────
export const customerApi = {
	getProfile: (token: string) => get<Customer>('/api/customer/profile', { token }),
	updateProfile: (token: string, data: Partial<Pick<Customer, 'full_name' | 'phone'>>) =>
		put<Customer>('/api/customer/profile', data, { token }),
	listReservations: (token: string) =>
		get<Reservation[]>('/api/customer/reservations', { token }),
	getReservation: (token: string, id: string) =>
		get<ReservationWithBookings>(`/api/customer/reservations/${id}`, { token })
};

// ─── Admin API ───────────────────────────────────────────────────
export const adminApi = {
	// Room Types
	listRoomTypes: (token: string) => get<RoomType[]>('/api/admin/room-types', { token }),
	createRoomType: (token: string, data: CreateRoomTypePayload) =>
		post<RoomType>('/api/admin/room-types', data, { token }),
	updateRoomType: (token: string, id: string, data: UpdateRoomTypePayload) =>
		put<RoomType>(`/api/admin/room-types/${id}`, data, { token }),
	deleteRoomType: (token: string, id: string) =>
		del<void>(`/api/admin/room-types/${id}`, { token }),

	// Rooms
	listRooms: (token: string, params?: { room_type_id?: string; status?: string }) => {
		const query = new URLSearchParams();
		if (params?.room_type_id) query.set('room_type_id', params.room_type_id);
		if (params?.status) query.set('status', params.status);
		const qs = query.toString();
		return get<Room[]>(`/api/admin/rooms${qs ? `?${qs}` : ''}`, { token });
	},
	createRoom: (token: string, data: CreateRoomPayload) =>
		post<Room>('/api/admin/rooms', data, { token }),
	updateRoom: (token: string, id: string, data: UpdateRoomPayload) =>
		put<Room>(`/api/admin/rooms/${id}`, data, { token }),
	deleteRoom: (token: string, id: string) =>
		del<void>(`/api/admin/rooms/${id}`, { token }),

	// Images
	uploadImage: (token: string, roomId: string, file: File) => {
		const form = new FormData();
		form.append('image', file);
		return post<RoomImage>(`/api/admin/rooms/${roomId}/images`, form, { token });
	},
	deleteImage: (token: string, roomId: string, imageId: string) =>
		del<void>(`/api/admin/rooms/${roomId}/images/${imageId}`, { token }),
	reorderImages: (token: string, roomId: string, imageIds: string[]) =>
		put<void>(`/api/admin/rooms/${roomId}/images/reorder`, { image_ids: imageIds }, { token }),

	// Pricing
	listPricing: (token: string, params?: { room_type_id?: string }) => {
		const query = new URLSearchParams();
		if (params?.room_type_id) query.set('room_type_id', params.room_type_id);
		const qs = query.toString();
		return get<RoomPricing[]>(`/api/admin/room-pricing${qs ? `?${qs}` : ''}`, { token });
	},
	createPricing: (token: string, data: CreatePricingPayload) =>
		post<RoomPricing>('/api/admin/room-pricing', data, { token }),
	updatePricing: (token: string, id: string, data: UpdatePricingPayload) =>
		put<RoomPricing>(`/api/admin/room-pricing/${id}`, data, { token }),
	deletePricing: (token: string, id: string) =>
		del<void>(`/api/admin/room-pricing/${id}`, { token }),

	// Inventory
	getInventory: (token: string, params: { date: string; room_type_id?: string }) => {
		const query = new URLSearchParams({ date: params.date });
		if (params.room_type_id) query.set('room_type_id', params.room_type_id);
		return get<RoomTypeInventory[]>(`/api/admin/inventory?${query}`, { token });
	},
	updateInventory: (token: string, data: UpdateInventoryPayload) =>
		put<void>('/api/admin/inventory', data, { token }),

	// Reservations
	listReservations: (token: string, params?: { status?: string; from?: string; to?: string }) => {
		const query = new URLSearchParams();
		if (params?.status) query.set('status', params.status);
		if (params?.from) query.set('from', params.from);
		if (params?.to) query.set('to', params.to);
		const qs = query.toString();
		return get<Reservation[]>(`/api/admin/reservations${qs ? `?${qs}` : ''}`, { token });
	},
	getReservation: (token: string, id: string) =>
		get<ReservationWithBookings>(`/api/admin/reservations/${id}`, { token }),
	updateReservationStatus: (token: string, id: string, data: UpdateReservationStatusPayload) =>
		put<Reservation>(`/api/admin/reservations/${id}/status`, data, { token }),

	// Customers
	listCustomers: (token: string) => get<Customer[]>('/api/admin/customers', { token }),
	getCustomer: (token: string, id: string) =>
		get<Customer>(`/api/admin/customers/${id}`, { token }),

	// Admins
	listAdmins: (token: string) => get<Admin[]>('/api/admin/admins', { token }),
	createAdmin: (token: string, data: CreateAdminPayload) =>
		post<Admin>('/api/admin/admins', data, { token }),
	updateAdmin: (token: string, id: string, data: UpdateAdminPayload) =>
		put<Admin>(`/api/admin/admins/${id}`, data, { token }),
	deleteAdmin: (token: string, id: string) =>
		del<void>(`/api/admin/admins/${id}`, { token }),

	// Reports
	bookingReport: (token: string, params: { from: string; to: string }) => {
		const query = new URLSearchParams(params);
		return get<BookingReport>(`/api/admin/reports/bookings?${query}`, { token });
	},
	occupancyReport: (token: string, params: { from: string; to: string }) => {
		const query = new URLSearchParams(params);
		return get<OccupancyReport>(`/api/admin/reports/occupancy?${query}`, { token });
	},
	revenueReport: (token: string, params: { from: string; to: string }) => {
		const query = new URLSearchParams(params);
		return get<RevenueReport>(`/api/admin/reports/revenue?${query}`, { token });
	}
};

// ─── SSE API ─────────────────────────────────────────────────────
export function createSSEConnection(
	token: string | null,
	onEvent: (event: MessageEvent) => void,
	onError?: (error: Event) => void
): EventSource {
	const base = getBaseUrl();
	const url = token ? `${base}/api/events?token=${token}` : `${base}/api/events`;
	const es = new EventSource(url);
	es.onmessage = onEvent;
	es.onerror = onError ?? null;
	return es;
}
