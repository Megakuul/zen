<script>
  import { cn } from "$lib/utils/cn";
  import { inview } from "svelte-inview";
  import { Motion } from "svelte-motion";

  /**
   * @type {{
   *  text: string,
   *  delay?: number,
   *  class?: string
   * }}
   */
  const { 
    text, 
    delay = 0.05,
    class: className = undefined,
  } = $props();

  let letters = $derived(text.split(""));
  let visible = $state(false);
</script>

<span use:inview oninview_enter={() => visible = true} class="flex justify-center">
  {#each letters as letter, i}
    <Motion
      variants={{
        hidden: { y: 100, opacity: 0 },
        visible: (i) => ({
          y: 0,
          opacity: 1,
          transition: { delay: i * delay },
        })
      }}
      initial="hidden"
      animate={visible ? "visible" : "hidden"}
      custom={i}
      let:motion
    >
      <span
        class={cn(
          "font-display text-center text-4xl font-bold tracking-[-0.02em] text-black drop-shadow-sm dark:text-white md:text-4xl md:leading-[5rem]",
          className
        )}
        use:motion
      >
        {#if letter === " "}
          <span>&nbsp;</span>
        {:else}
          {letter}
        {/if}
      </span>
    </Motion>
  {/each}
</span>
