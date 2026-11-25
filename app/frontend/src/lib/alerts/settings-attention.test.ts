import { describe, expect, it } from 'vitest';
import { render } from 'svelte/server';
import SettingsAttention from './SettingsAttention.svelte';
import { fromServer, type Attention } from './settings-status.ts';

const items: Attention[] = [
	{ group: 'general', href: '/settings/this-app', text: 'HolyArrs 1.3.1 is out (this is 1.3.1-mixed.cbf81423)', severity: 'warning' },
	{ group: 'general', href: '/settings/this-app', text: 'A newer HolyArrs is out', severity: 'info' }
];

describe('settings attention details', () => {
	it('shows the actual reasons behind the page count, including informational notices', () => {
		const { body } = render(SettingsAttention, { props: { items } });
		expect(body).toContain('2 need attention');
		for (const item of items) expect(body).toContain(item.text);
		expect(body).toContain('data-severity="warning"');
		expect(body).toContain('data-severity="info"');
		expect(body).toContain('id="settings-attention"');
	});

	it('renders nothing when the page has no issues', () => {
		const { body } = render(SettingsAttention, { props: { items: [] } });
		expect(body).not.toContain('<section');
	});

	it('keeps other pages out and escapes diagnostic text', () => {
		const a = fromServer({ items: [
			{ group: 'general', page: '/settings/this-app', text: '<script>bad()</script>', severity: 'error' },
			{ group: 'copies', page: '/settings/downloaders', text: 'Downloader is offline', severity: 'error' }
		], known: [], unset: {} });
		const { body } = render(SettingsAttention, { props: { items: a.items.filter((i) => i.href === '/settings/this-app') } });
		expect(body).toContain('1 needs attention');
		expect(body).toContain('&lt;script>');
		expect(body).not.toContain('<script>bad()');
		expect(body).not.toContain('Downloader is offline');
	});
});
