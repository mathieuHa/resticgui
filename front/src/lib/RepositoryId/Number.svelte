<script lang="ts">
	export let number: number;
	const locale = "en";

	$: formattedNumber = number.toLocaleString(locale, {
		notation: "compact",
		compactDisplay: "short"
	});
</script>

<style>
	span {
		font-variant-numeric: tabular-nums;
	}
</style>

<span>{formattedNumber}</span>
