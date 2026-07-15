import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { authApi, setRefreshHandler } from '$lib/api/client';
import { ApiException } from '$lib/api/client';
import type { Customer, Admin } from '$lib/types/api';

interface AuthState {
	customer: Customer | null;
	admin: Admin | null;
	token: string | null;
	refreshToken: string | null;
	role: 'customer' | 'admin' | null;
	adminRole: string | null;
	loading: boolean;
	error: string | null;
}

const SESSION_KEY = 'hotel_session';

function createAuthStore() {
	// In-memory only — tokens are NEVER persisted to localStorage
	let initial: AuthState = {
		customer: null,
		admin: null,
		token: null,
		refreshToken: null,
		role: null,
		adminRole: null,
		loading: false,
		error: null
	};

	// Restore non-sensitive profile data from sessionStorage so the UI
	// doesn't flash empty after a page reload. Tokens stay in-memory.
	if (browser) {
		const sessionData = sessionStorage.getItem('hotel_session');
		if (sessionData) {
			try {
				const parsed = JSON.parse(sessionData);
				initial = {
					...initial,
					customer: parsed.customer ?? null,
					admin: parsed.admin ?? null,
					role: parsed.role ?? null,
					adminRole: parsed.adminRole ?? null
				};
			} catch {
				// Corrupted — ignore
			}
		}
	}

	const { subscribe, set, update } = writable<AuthState>(initial);

	// Persist only non-sensitive profile data — never tokens
	function persistProfile(state: AuthState) {
		if (!browser) return;
		if (state.token) {
			sessionStorage.setItem(
				SESSION_KEY,
				JSON.stringify({
					customer: state.customer,
					admin: state.admin,
					role: state.role,
					adminRole: state.adminRole
				})
			);
		} else {
			sessionStorage.removeItem(SESSION_KEY);
		}
	}

	let refreshing: Promise<string | null> | null = null;

	return {
		subscribe,

		async register(fullName: string, email: string, password: string, phone?: string) {
			update((s) => ({ ...s, loading: true, error: null }));
			try {
				const res = await authApi.register({ name: fullName, email, password, phone });
				const state: AuthState = {
					customer: res.user,
					admin: null,
					token: res.access_token,
					refreshToken: res.refresh_token,
					role: 'customer',
					adminRole: null,
					loading: false,
					error: null
				};
				persistProfile(state);
				set(state);
			} catch (e) {
				const msg = e instanceof ApiException ? e.message : 'Registration failed';
				update((s) => ({ ...s, loading: false, error: msg }));
				throw e;
			}
		},

		async login(email: string, password: string) {
			update((s) => ({ ...s, loading: true, error: null }));
			try {
				const res = await authApi.login({ email, password });
				const state: AuthState = {
					customer: res.user,
					admin: null,
					token: res.access_token,
					refreshToken: res.refresh_token,
					role: 'customer',
					adminRole: null,
					loading: false,
					error: null
				};
				persistProfile(state);
				set(state);
			} catch (e) {
				const msg = e instanceof ApiException ? e.message : 'Login failed';
				update((s) => ({ ...s, loading: false, error: msg }));
				throw e;
			}
		},

		async adminLogin(email: string, password: string) {
			update((s) => ({ ...s, loading: true, error: null }));
			try {
				const res = await authApi.adminLogin({ email, password });
				const state: AuthState = {
					customer: null,
					admin: res.user,
					token: res.access_token,
					refreshToken: res.refresh_token,
					role: 'admin',
					adminRole: res.user.role,
					loading: false,
					error: null
				};
				persistProfile(state);
				set(state);
			} catch (e) {
				const msg = e instanceof ApiException ? e.message : 'Admin login failed';
				update((s) => ({ ...s, loading: false, error: msg }));
				throw e;
			}
		},

		async refresh(): Promise<string | null> {
			// Use mutex to avoid multiple concurrent refreshes
			if (refreshing) return refreshing;

			refreshing = (async () => {
				let currentRefreshToken: string | null = null;
				subscribe((s) => (currentRefreshToken = s.refreshToken))();

				if (!currentRefreshToken) return null;
				try {
					const res = await authApi.refresh(currentRefreshToken);
					update((s) => {
						const updated: AuthState = {
							...s,
							token: res.access_token,
							refreshToken: res.refresh_token ?? s.refreshToken
						};
						// Tokens stay in-memory; profile is already persisted
						return updated;
					});
					return res.access_token;
				} catch {
					this.logout();
					return null;
				} finally {
					refreshing = null;
				}
			})();

			return refreshing;
		},

		logout() {
			const state: AuthState = {
				customer: null,
				admin: null,
				token: null,
				refreshToken: null,
				role: null,
				adminRole: null,
				loading: false,
				error: null
			};
			persistProfile(state);
			set(state);
		},

		clearError() {
			update((s) => ({ ...s, error: null }));
		},

		getToken(): string | null {
			let token: string | null = null;
			subscribe((s) => (token = s.token))();
			return token;
		},

		getCustomerId(): string | null {
			let id: string | null = null;
			subscribe((s) => (id = s.customer?.id ?? null))();
			return id;
		},

		isAuthenticated(): boolean {
			let authed = false;
			subscribe((s) => (authed = !!s.token))();
			return authed;
		},

		isAdmin(): boolean {
			let admin = false;
			subscribe((s) => (admin = s.role === 'admin'))();
			return admin;
		},

		hasRole(...roles: string[]): boolean {
			let match = false;
			subscribe((s) => (match = roles.includes(s.adminRole ?? '')))();
			return match;
		},

		// Called on page load — attempts silent refresh if a session exists
		async tryRestore() {
			if (!browser) return;
			const hasSession = sessionStorage.getItem(SESSION_KEY);
			if (!hasSession) return;
			// We have profile data but no token — try to refresh via the refresh token cookie
			await this.refresh();
		}
	};
}

export const auth = createAuthStore();

// Register auto-refresh handler so request() can retry once on 401
if (browser) {
	setRefreshHandler(async () => {
		try {
			return await auth.refresh();
		} catch {
			return null;
		}
	});

	// Attempt to clean up any legacy localStorage tokens from older versions
	if (browser) {
		localStorage.removeItem('hotel_auth');
	}

	// Attempt silent token restore on first load
	auth.tryRestore();
}
