interface Metric {
	name: string;
	value: number;
	rating: 'good' | 'needs-improvement' | 'poor';
	delta: number;
	id: string;
}

type MetricCallback = (metric: Metric) => void;

function getRating(name: string, value: number): 'good' | 'needs-improvement' | 'poor' {
	const thresholds: Record<string, [number, number]> = {
		LCP: [2500, 4000],
		INP: [200, 500],
		CLS: [0.1, 0.25],
		FCP: [1800, 3000],
		TTFB: [800, 1800]
	};

	const [good, poor] = thresholds[name] ?? [0, 0];
	if (value <= good) return 'good';
	if (value <= poor) return 'needs-improvement';
	return 'poor';
}

function observe(type: string, callback: (entry: PerformanceEntry) => void) {
	try {
		const observer = new PerformanceObserver((list) => {
			for (const entry of list.getEntries()) {
				callback(entry);
			}
		});
		observer.observe({ type, buffered: true });
		return observer;
	} catch {
		return null;
	}
}

export function onLCP(callback: MetricCallback) {
	return observe('largest-contentful-paint', (entry) => {
		const lcpEntry = entry as PerformanceEntry & { startTime: number; size: number; id: string };
		callback({
			name: 'LCP',
			value: lcpEntry.startTime,
			rating: getRating('LCP', lcpEntry.startTime),
			delta: 0,
			id: lcpEntry.id ?? String(Date.now())
		});
	});
}

export function onINP(callback: MetricCallback) {
	return observe('event', (entry) => {
		const eventEntry = entry as PerformanceEntry & { duration: number; interactionId?: string };
		if (eventEntry.duration > 0) {
			callback({
				name: 'INP',
				value: eventEntry.duration,
				rating: getRating('INP', eventEntry.duration),
				delta: 0,
				id: eventEntry.interactionId ?? String(Date.now())
			});
		}
	});
}

export function onCLS(callback: MetricCallback) {
	let sessionValue = 0;
	let sessionEntries: number[] = [];

	return observe('layout-shift', (entry) => {
		const shiftEntry = entry as PerformanceEntry & { hadRecentInput: boolean; value: number };
		if (shiftEntry.hadRecentInput) return;

		const currentSession = shiftEntry.startTime - (sessionEntries[0] ?? 0) > 1000;

		if (currentSession) {
			sessionValue = 0;
			sessionEntries = [];
		}

		sessionValue += shiftEntry.value;
		sessionEntries.push(shiftEntry.startTime);

		callback({
			name: 'CLS',
			value: sessionValue,
			rating: getRating('CLS', sessionValue),
			delta: 0,
			id: String(Date.now())
		});
	});
}

export function onFCP(callback: MetricCallback) {
	return observe('paint', (entry) => {
		if (entry.name === 'first-contentful-paint') {
			callback({
				name: 'FCP',
				value: entry.startTime,
				rating: getRating('FCP', entry.startTime),
				delta: 0,
				id: String(Date.now())
			});
		}
	});
}

export function onTTFB(callback: MetricCallback) {
	return observe('navigation', (entry) => {
		const navEntry = entry as PerformanceNavigationTiming;
		const ttfb = navEntry.responseStart - navEntry.requestStart;
		callback({
			name: 'TTFB',
			value: ttfb,
			rating: getRating('TTFB', ttfb),
			delta: 0,
			id: String(Date.now())
		});
	});
}

export function reportWebVitals(callback: MetricCallback) {
	if (typeof window === 'undefined') return;

	const observers = [onLCP(callback), onINP(callback), onCLS(callback), onFCP(callback), onTTFB(callback)];

	return () => {
		observers.forEach((o) => o?.disconnect());
	};
}
