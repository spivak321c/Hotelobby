// ─── API Response Envelope ───────────────────────────────────────
export interface ApiError {
	code: string;
	message: string;
}

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: ApiError;
}

export interface PaginatedResponse<T> {
	items: T[];
	total: number;
	page: number;
	per_page: number;
}

// ─── Enums ───────────────────────────────────────────────────────
export type RoomStatus = 'active' | 'maintenance' | 'inactive';
export type BookingType = 'hourly' | 'daily';
export type RateType = 'hourly' | 'daily';
export type ReservationStatus =
	| 'pending'
	| 'confirmed'
	| 'cancelled'
	| 'refunded'
	| 'failed'
	| 'completed';
export type BookingStatus =
	| 'pending'
	| 'confirmed'
	| 'checked_in'
	| 'checked_out'
	| 'cancelled'
	| 'refunded'
	| 'failed';
export type PaymentStatus = 'pending' | 'processing' | 'succeeded' | 'failed' | 'refunded';
export type PaymentProvider = 'paystack' | 'crossmint';
export type AdminRole = 'super_admin' | 'manager' | 'front_desk';

// ─── Domain Models (match backend v2 field names) ────────────────
export interface RoomType {
	id: string;
	name: string;
	description: string;
	base_rate_daily: number;
	base_rate_hourly: number;
	max_occupancy: number;
	is_featured: boolean;
	created_at: string;
	updated_at: string;
}

export interface Room {
	id: string;
	room_type_id: string;
	room_number: string;
	status: RoomStatus;
	upcoming_bookings: number;
	created_at: string;
	updated_at: string;
}

export interface RoomWithImages {
	room: Room;
	room_type_name: string;
	base_rate_daily: number;
	base_rate_hourly: number;
	images: RoomImage[];
}

export interface RoomImage {
	id: string;
	room_id: string;
	url: string;
	is_primary: boolean;
	sort_order: number;
}

export interface RoomPricing {
	id: string;
	room_type_id: string;
	rate_type: RateType;
	rate: number;
	effective_range: DateRange;
}

export interface DateRange {
	lower: string;
	upper: string;
	bounds: string;
}

export interface RoomTypeInventory {
	room_type_id: string;
	date: string;
	total_rooms: number;
	booked_rooms: number;
}

export interface Customer {
	id: string;
	full_name: string;
	email: string;
	phone: string;
	created_at: string;
	updated_at: string;
}

export interface Admin {
	id: string;
	full_name: string;
	email: string;
	role: AdminRole;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface Reservation {
	id: string;
	reference_code: string;
	customer_id?: string;
	guest_name: string;
	guest_email: string;
	guest_phone: string;
	total_amount: number;
	currency: string;
	status: ReservationStatus;
	idempotency_key?: string;
	created_by_admin_id?: string;
	created_at: string;
	updated_at: string;
}

export interface ReservationWithBookings extends Reservation {
	bookings: Booking[];
	payment?: Payment;
}

export interface Booking {
	id: string;
	reservation_id: string;
	room_id: string;
	room_type_id: string;
	booking_type: BookingType;
	starts_at: string;
	ends_at: string;
	status: BookingStatus;
	amount: number;
}

export interface Payment {
	id: string;
	reservation_id: string;
	provider: PaymentProvider;
	provider_reference: string;
	status: PaymentStatus;
	amount: number;
	currency: string;
	metadata?: Record<string, unknown>;
	created_at: string;
	updated_at: string;
}

// ─── Auth Types ──────────────────────────────────────────────────
export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	name: string;
	email: string;
	password: string;
	phone?: string;
}

export interface AuthResponse {
	access_token: string;
	refresh_token: string;
	user: Customer;
}

export interface AdminLoginRequest {
	email: string;
	password: string;
}

export interface AdminAuthResponse {
	access_token: string;
	refresh_token: string;
	user: Admin;
}

// ─── Reservation / Booking DTOs ──────────────────────────────────
export interface CreateBookingPayload {
	room_id: string;
	check_in: string;
	check_out: string;
	booking_type: BookingType;
	expected_occupants: number;
}

export interface CreateReservationPayload {
	guest_name: string;
	guest_email: string;
	guest_phone?: string;
	customer_id?: string;
	bookings: CreateBookingPayload[];
	payment_method: PaymentProvider;
	idempotency_key?: string;
}

export interface CancelReservationPayload {
	otp: string;
	reason?: string;
}

export interface ProcessPaymentPayload {
	reservation_id: string;
	method: PaymentProvider;
	card_details?: {
		number: string;
		cvv: string;
		expiry_month: string;
		expiry_year: string;
	};
}

// ─── Room Lookup ─────────────────────────────────────────────────
export interface AvailabilityQuery {
	check_in: string;
	check_out: string;
	type: BookingType;
}

export interface AvailabilityResult {
	room_type_id: string;
	available: boolean;
	available_rooms: number;
	total_price: number;
	dates: {
		date: string;
		available: number;
	}[];
}

// ─── Admin DTOs ──────────────────────────────────────────────────
export interface CreateRoomTypePayload {
	name: string;
	description?: string;
	base_rate_daily: number;
	base_rate_hourly?: number;
	max_occupancy?: number;
	is_featured?: boolean;
}

export interface UpdateRoomTypePayload {
	name?: string;
	description?: string;
	base_rate_daily?: number;
	base_rate_hourly?: number;
	max_occupancy?: number;
	is_featured?: boolean;
}

export interface CreateRoomPayload {
	room_type_id: string;
	room_number: string;
	status?: RoomStatus;
}

export interface UpdateRoomPayload {
	room_type_id?: string;
	room_number?: string;
	status?: RoomStatus;
}

export interface CreatePricingPayload {
	room_type_id: string;
	rate_type: RateType;
	rate: number;
	effective_from: string;
	effective_to: string;
}

export interface UpdatePricingPayload {
	room_type_id?: string;
	rate_type?: RateType;
	rate?: number;
	effective_from?: string;
	effective_to?: string;
}

export interface UpdateInventoryPayload {
	room_type_id: string;
	date: string;
	total_rooms: number;
	booked_rooms: number;
}

export interface UpdateReservationStatusPayload {
	status: ReservationStatus;
	reason?: string;
}

export interface CreateAdminPayload {
	full_name: string;
	email: string;
	password: string;
	role: AdminRole;
}

export interface UpdateAdminPayload {
	full_name?: string;
	email?: string;
	role?: AdminRole;
	is_active?: boolean;
}

// ─── Reports ─────────────────────────────────────────────────────
export interface BookingReport {
	from: string;
	to: string;
	total_bookings: number;
	total_revenue: number;
	by_status: Record<string, number>;
}

export interface OccupancyReport {
	from: string;
	to: string;
	total_rooms: number;
	occupied_rooms: number;
	occupancy_rate: number;
}

export interface RevenueReport {
	from: string;
	to: string;
	total_revenue: number;
	by_status: Record<string, number>;
	cancelled_revenue: number;
}

// ─── SSE Events ──────────────────────────────────────────────────
export interface RoomAvailability {
	room_id: string;
	room_number: string;
	available: boolean;
}

export interface AvailabilityEvent {
	type: 'availability';
	room_type_id: string;
	date: string;
	rooms: RoomAvailability[];
}

export interface BookingUpdatedEvent {
	type: 'booking-updated';
	reference: string;
	status: ReservationStatus;
}

export type SSEEvent = AvailabilityEvent | BookingUpdatedEvent;
