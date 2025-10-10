<script>
  import { goto } from "$app/navigation";
  import { cn } from "$lib/utils/cn";
    import { fade } from "svelte/transition";

  /**
   * @type {{
   *  href: string,
   *  mainCircleSize?: number,
   *  mainCircleOpacity?: number,
   *  numCircles?: number,
   *  class?: string
   * }}
   */
  const { 
    href,
    mainCircleSize = 160,
    mainCircleOpacity = 0.24,
    numCircles = 8,
    class: className = undefined,
  } = $props();

  const titles = [
    "Unleash",
    "your",
    "Time",
    "►"
  ]
  let titleIdx = $state(0)
  let title = $derived(titles[titleIdx])

  $effect.root(() => {
    const interval = setInterval(() => {
      if (titleIdx+1 >= titles.length) titleIdx = 0
      else titleIdx++
    }, 4000)
    return () => clearInterval(interval)
  })
</script>

<div 
  style="width: {mainCircleSize + numCircles * 70}px; height: {mainCircleSize + numCircles * 70}px;" 
  class={cn("relative flex items-center justify-center overflow-hidden", className)}>
  {#each { length: numCircles } as _, i}
    {#if i === 0}
      {#key title}
        <button transition:fade onclick={() => goto(href)}
          class="circle cursor-pointer absolute z-1 font-bold text-3xl sm:text-4xl transition-all duration-700 hover:bg-slate-50/10 rounded-full bg-foreground/30 shadow-xl border [--i:{i}]"
          style="width: {mainCircleSize + i * 70}px;
            height: {mainCircleSize + i * 70}px;
            opacity: {mainCircleOpacity - i * 0.03}; 
            animation-delay: {i * 0.08}s;
            border-style: solid;
            border-width: 1px;
            border-color: rgba(var(--foreground-rgb), {(5 + i * 5) / 100});"
        >{title}</button>
      {/key}
    {:else}
      <div
        class="circle absolute rounded-full bg-foreground/30 shadow-xl border [--i:{i}]"
        style="width: {mainCircleSize + i * 70}px;
          height: {mainCircleSize + i * 70}px;
          opacity: {mainCircleOpacity - i * 0.03}; 
          animation-delay: {i * 0.08}s;
          border-style: solid;
          border-width: 1px;
          border-color: rgba(var(--foreground-rgb), {(5 + i * 5) / 100});"
      ></div>
    {/if}
  {/each}
</div>

<style>
  .circle {
    top: 50%;
    left: 50%;
    animation: animate-ripple 2s ease calc(var(--i, 0)*.2s) infinite;
  }

  @keyframes animate-ripple {
    0%, 100% {
      transform: translate(-50%, -50%) scale(1);
    }
    50% {
      transform: translate(-50%, -50%) scale(0.9);
    }
  }
</style>