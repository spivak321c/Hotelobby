type ToastType = 'success' | 'error' | 'info' | 'warning';

interface Toast {
	id: string;
	type: ToastType;
	message: string;
	description?: string;
	duration: number;
}

let toasts = $state<Toast[]>([]);

function add(type: ToastType, message: string, description?: string, duration = 4000) {
	const id = crypto.randomUUID();
	toasts.push({ id, type, message, description, duration });
	if (duration > 0) {
		setTimeout(() => remove(id), duration);
	}
	return id;
}

function remove(id: string) {
	const idx = toasts.findIndex((t) => t.id === id);
	if (idx !== -1) toasts.splice(idx, 1);
}

function success(message: string, description?: string) {
	return add('success', message, description);
}
function error(message: string, description?: string) {
	return add('error', message, description, 6000);
}
function info(message: string, description?: string) {
	return add('info', message, description);
}
function warning(message: string, description?: string) {
	return add('warning', message, description, 5000);
}

export const toast = {
	get toasts() {
		return toasts;
	},
	add,
	remove,
	success,
	error,
	info,
	warning,
};
