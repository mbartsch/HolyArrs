<script lang="ts">
	import { serverWords } from '#lib/core/i18n/server-words.ts';
	import AlertSymbol from './AlertSymbol.svelte';
	import { stateWords, type Attention } from './settings-status.ts';

	let { items }: { items: Attention[] } = $props();
	const title = $derived(stateWords({ state: 'attention', count: items.length }));
</script>

{#if items.length}
	<section id="settings-attention" aria-label={title} data-testid="settings-attention" tabindex="-1">
		<h2>{title}</h2>
		<ul>
			{#each items as item}
				<li data-severity={item.severity}>
					<AlertSymbol tone={item.severity === 'error' ? 'fail' : item.severity === 'warning' ? 'warn' : 'info'} size={20} />
					<span>{serverWords(item.text)}</span>
				</li>
			{/each}
		</ul>
	</section>
{/if}

<style>
	section { padding: var(--s-4); margin-block-end: var(--s-4); border: 1px solid var(--line); border-radius: var(--r-card); background: var(--surface-1); scroll-margin-top: 90px; }
	h2 { margin: 0 0 var(--s-3); font-size: var(--text-sm); }
	ul { list-style: none; padding: 0; margin: 0; display: grid; gap: var(--s-3); }
	li { display: flex; align-items: start; gap: var(--s-3); font-size: var(--text-sm); line-height: 1.5; }
	li span { min-width: 0; overflow-wrap: anywhere; }
</style>
