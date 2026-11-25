<script lang="ts">
 import { t } from '#lib/core/i18n/i18n.svelte.ts';
 import Field from '#lib/ui/Field.svelte';
 let { shared = $bindable(''), movies = $bindable(''), shows = $bindable(''), music = $bindable('') }: { shared?: string; movies?: string; shows?: string; music?: string } = $props();
</script>

<Field id="dl-cat-after" label={t('settings.downloaders.categoryAfterImport')} hint={t('settings.downloaders.categoryAfterImportHint')}>
		<input id="dl-cat-after" class="input" bind:value={shared} maxlength="64" aria-describedby="dl-cat-after-hint" data-testid="dl-cat-after" />
	</Field>
	<Field id="dl-cat-after-movies" label={t('settings.downloaders.categoryAfterImportMovies')} hint={t('settings.downloaders.categoryAfterImportOverrideHint')}>
		<input id="dl-cat-after-movies" class="input" bind:value={movies} placeholder="movies-imported" maxlength="64" aria-describedby="dl-cat-after-movies-hint" data-testid="dl-cat-after-movies" />
	</Field>
	<Field id="dl-cat-after-shows" label={t('settings.downloaders.categoryAfterImportShows')} hint={t('settings.downloaders.categoryAfterImportOverrideHint')}>
		<input id="dl-cat-after-shows" class="input" bind:value={shows} placeholder="shows-imported" maxlength="64" aria-describedby="dl-cat-after-shows-hint" data-testid="dl-cat-after-shows" />
	</Field>
	<Field id="dl-cat-after-music" label={t('settings.downloaders.categoryAfterImportMusic')} hint={t('settings.downloaders.categoryAfterImportOverrideHint')}>
		<input id="dl-cat-after-music" class="input" bind:value={music} placeholder="music-imported" maxlength="64" aria-describedby="dl-cat-after-music-hint" data-testid="dl-cat-after-music" />
	</Field>
