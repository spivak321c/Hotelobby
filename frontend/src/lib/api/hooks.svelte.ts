import { ApiException } from '$lib/api/client';
import type { ApiResponse } from '$lib/types/api';

// ─── useQuery ────────────────────────────────────────────────────
export function useQuery<T>(
	fetcher: () => Promise<T>,
	options?: { immediate?: boolean }
) {
	let data = $state<T | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(false);
	let succeeded = $state(false);

	async function execute() {
		loading = true;
		error = null;
		try {
			data = await fetcher();
			succeeded = true;
		} catch (e) {
			error = e instanceof ApiException ? e.message : 'Request failed';
			succeeded = false;
		} finally {
			loading = false;
		}
	}

	function reset() {
		data = null;
		error = null;
		loading = false;
		succeeded = false;
	}

	if (options?.immediate !== false) {
		$effect(() => {
			execute();
		});
	}

	return {
		get data() {
			return data;
		},
		get error() {
			return error;
		},
		get loading() {
			return loading;
		},
		get succeeded() {
			return succeeded;
		},
		execute,
		reset
	};
}

// ─── useMutation ─────────────────────────────────────────────────
export function useMutation<TArgs, TResult>(
	mutator: (args: TArgs) => Promise<TResult>
) {
	let data = $state<TResult | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(false);

	async function mutate(args: TArgs): Promise<TResult | null> {
		loading = true;
		error = null;
		try {
			data = await mutator(args);
			return data;
		} catch (e) {
			error = e instanceof ApiException ? e.message : 'Mutation failed';
			return null;
		} finally {
			loading = false;
		}
	}

	function reset() {
		data = null;
		error = null;
		loading = false;
	}

	return {
		get data() {
			return data;
		},
		get error() {
			return error;
		},
		get loading() {
			return loading;
		},
		mutate,
		reset
	};
}
