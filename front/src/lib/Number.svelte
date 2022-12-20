<script lang="ts">
	export let number: number;

	const locale = 'en';

	$: formattedNumber = number.toLocaleString(locale, {
		notation: 'compact',
		compactDisplay: 'short',
	});
</script>

<span class="tabular-nums">{formattedNumber}</span>
