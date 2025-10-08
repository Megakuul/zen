<script>
  import { cn } from "$lib/utils/cn";

  /**
   * @type {{
   *  mainCircleSize?: number,
   *  mainCircleOpacity?: number,
   *  numCircles?: number,
   *  class?: string
   * }}
   */
  const { 
    mainCircleSize = 160,
    mainCircleOpacity = 0.24,
    numCircles = 8,
    class: className = undefined,
  } = $props();
</script>

<div 
  style="width: {mainCircleSize + numCircles * 70}px; height: {mainCircleSize + numCircles * 70}px;" 
  class={cn("relative flex items-center justify-center overflow-hidden", className)}>
  {#each { length: numCircles } as _, i}
    {#if i === 0}
      <button onclick={() => alert("ALARM")}
        class="circle cursor-pointer absolute z-1 text-2xl sm:text-3xl transition-all duration-700 hover:bg-slate-50/10 rounded-full bg-foreground/30 shadow-xl border [--i:{i}]"
        style="width: {mainCircleSize + i * 70}px;
          height: {mainCircleSize + i * 70}px;
          opacity: {mainCircleOpacity - i * 0.03}; 
          animation-delay: {i * 0.08}s;
          border-style: solid;
          border-width: 1px;
          border-color: rgba(var(--foreground-rgb), {(5 + i * 5) / 100});"
      >Unleash</button>
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