import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

gsap.registerPlugin(ScrollTrigger);

export const easeOutExpo = 'cubic-bezier(0.16, 1, 0.3, 1)';
export const easeInOut = 'cubic-bezier(0.65, 0, 0.35, 1)';

export function prefersReducedMotion(): boolean {
	if (typeof window === 'undefined') return false;
	return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

export function splitWords(text: string): { word: string; key: string }[] {
	return text.split(' ').map((w, i) => ({ word: w, key: `${i}-${w}` }));
}

export function initScrollReveals(reduced: boolean) {
	const sections = document.querySelectorAll<HTMLElement>('[data-reveal]');
	sections.forEach((el) => {
		if (reduced) {
			gsap.set(el, { opacity: 1, y: 0 });
			return;
		}
		gsap.fromTo(el,
			{ opacity: 0, y: 40 },
			{
				opacity: 1, y: 0,
				duration: 0.8,
				ease: easeOutExpo,
				scrollTrigger: {
					trigger: el,
					start: 'top 85%',
					toggleActions: 'play none none none',
				},
			}
		);
	});
}

export function initStaggerReveals(selector: string, reduced: boolean) {
	const containers = document.querySelectorAll<HTMLElement>(selector);
	containers.forEach((container) => {
		const items = container.querySelectorAll<HTMLElement>('[data-stagger-item]');
		if (!items.length) return;
		if (reduced) {
			gsap.set(items, { opacity: 1, y: 0 });
			return;
		}
		gsap.fromTo(items,
			{ opacity: 0, y: 30 },
			{
				opacity: 1, y: 0,
				duration: 0.7,
				stagger: 0.08,
				ease: easeOutExpo,
				scrollTrigger: {
					trigger: container,
					start: 'top 85%',
					toggleActions: 'play none none none',
				},
			}
		);
	});
}

export function initWordStaggers(reduced: boolean) {
	const headlines = document.querySelectorAll<HTMLElement>('[data-word-stagger]');
	headlines.forEach((headline) => {
		const words = headline.querySelectorAll<HTMLElement>('[data-word]');
		if (!words.length) return;
		if (reduced) {
			gsap.set(words, { opacity: 1, y: 0 });
			return;
		}
		gsap.fromTo(words,
			{ opacity: 0, y: 20 },
			{
				opacity: 1, y: 0,
				duration: 0.6,
				stagger: 0.06,
				ease: easeOutExpo,
				scrollTrigger: {
					trigger: headline,
					start: 'top 85%',
					toggleActions: 'play none none none',
				},
			}
		);
	});
}
