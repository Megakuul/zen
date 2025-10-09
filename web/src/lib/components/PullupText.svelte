<script>
  import { cn } from "$lib/utils/cn";
  import { inview } from "svelte-inview";
  import { Motion } from "svelte-motion";

  /**
   * @type {{
   *  visible: boolean,
   *  delay?: number,
   *  class?: string
   * }}
   */
  let {
    visible = $bindable(),
    delay = 0.05,
    class: className = undefined,
  } = $props();

  let lines = $state({
    "Sleep": "441,504,000 seconds",
    "Work": "441,504,000 seconds",
    "Meals": "110,376,000 seconds",
    "Mother": "55,188,000 seconds",
    "Budgerigar": "13,797,000 seconds",
    "Shopping, etc.": "55,188,000 seconds",
    "Friends, etc.": "165,564,000 seconds",
    "Miss Daria": "27,594,000 seconds",
    "Daydreaming": "13,797,000 seconds",
    "Grand Total": "1,324,512,000 seconds",
  })
  let cumulativeDelay = 0;
</script>

<div class="glass flex flex-col justify-center items-center p-20 rounded-2xl">
  <span use:inview oninview_enter={() => visible = true} class="flex justify-center">
    {#each Object.entries(lines) as [key, value], i}
      {@const keyLetters = key.split("")}
      {@const valueLetters = value.split("")}
      {@const cumulativeDelay = i * (keyLetters.length + valueLetters.length)}
      <span class="flex flex-row justify-start w-full">
        {#each keyLetters.concat(valueLetters) as letter, j}
        <Motion
          variants={{
            hidden: { y: 100, opacity: 0 },
            visible: (i) => ({
              y: 0,
              opacity: 1,
              transition: { delay: (cumulativeDelay + j) * delay },
            })
          }}
          initial="hidden"
          animate={visible ? "visible" : "hidden"}
          custom={i}
          let:motion>
          <span use:motion class="{j === keyLetters.length ? "ml-auto" : ""} font-display text-center text-4xl font-bold tracking-[-0.02em] text-black drop-shadow-sm dark:text-white md:text-4xl md:leading-[5rem]">
            {#if letter === " "}
              <span>&nbsp;</span>
            {:else}
              {letter}
            {/if}
          </span>
        </Motion>
        {/each}
      </span>
    {/each}
  </span>
</div>

