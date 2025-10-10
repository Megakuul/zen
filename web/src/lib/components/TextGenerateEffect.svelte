<script>
  import { cn } from '$lib/utils/cn';
  import { inview } from 'svelte-inview';
	import { Motion } from 'svelte-motion';

	/** 
	 * @typedef TextSegment
	 * @property {string} text
	 * @property {boolean} [newLine]
	 * @property {string} class
	*/

  /**
   * @type {{
   *  segments: TextSegment[],
   *  durationPerWord: number,
   *  class?: string
   * }}
   */
  const { 
    segments, 
    durationPerWord, 
    class: className = undefined,
  } = $props();

	/** @type {TextSegment[]}*/
	// Flat segment array that every word has its own segment, inheriting other properties from the text segment.
	const wordSegments = segments.flatMap(
		block => block.text.split(' ')
			.map(word => ({ text: word, class: block.class, newLine: block.newLine})))

	/** @type {string} */
	let animationState = $state("hidden");

	const variants = {
		visible: (/** @type {number} */ i) => ({
			opacity: 1,
			transition: {
				delay: i * (durationPerWord / 6),
				duration: durationPerWord
			}
		}),
		hidden: { opacity: 0 }
	};
</script>

<div         
	use:inview
	oninview_change={() => animationState = "visible"}
	class={className}>
	<div class="text-2xl leading-snug tracking-wide">
		<Motion let:motion custom={0} {variants} initial="hidden" animate={'visible'}>
			<div use:motion>
				{#each wordSegments as word, idx}
				<Motion let:motion {variants} custom={idx + 1} initial="hidden" animate={animationState}>
					<span use:motion class={cn("text-inherit", animationState==="hidden" ? "opacity-0" : "", word.class)}>
						{word.text}{' '}
						{#if word.newLine}
						<br>
						{/if}
					</span>
				</Motion>
				{/each}
      </div>
		</Motion>
	</div>
</div>