<script>
    import { cn } from "$lib/utils/cn";

  /**
   * @type {{
   *  class?: string
   *  svgClass?: string
   * }}
   */
  let {
    class: className = undefined,
    svgClass = undefined,
  } = $props();
</script>

<div class={cn("glass flex justify-center items-center w-min h-min rounded-2xl", className)}>
  <svg class={cn("logo", svgClass)} viewBox="0 0 588 583" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path id="inner" class="layer" d="M126.84 117.601L461.385 117.601L126.841 471.709L461.568 472.717" stroke-width="24"/>
    <path id="center" class="layer" d="M514.436 68.8807L72.7088 68.8807L72.7092 165.238L335.795 165.238L72.7092 417.501L72.7092 521.437L514.437 521.437L514.437 417.501L267.587 417.501L514.437 165.238L514.436 68.8807Z" stroke-width="24"/>
    <path id="outer" class="layer" d="M575.442 12.5819L12.5583 12.582L12.5574 217.207L195.048 217.207L12.5563 385.021L12.5576 570.158L575.442 570.158L575.442 369.863L420.244 369.863L575.442 217.207L575.442 12.5819Z" stroke-width="24"/>
  </svg>
</div>

<style>
  .logo {
    animation: animate-pulse 4s infinite linear;
  }

  @keyframes animate-pulse {
    0%, 100% {
      filter: brightness(40%);
    }
    40% {
      filter: brightness(100%);
    }
  }

  .layer {
    stroke: rgba(255, 255,255, 0.5);
    stroke-width: 20px;
    stroke-linecap: round;
    stroke-linejoin: round;
    animation: animate-layer 4s forwards ease-in-out;
  }

  #outer {
    stroke-dasharray: 3000;
    stroke-dashoffset: 3000;
  }

  #center {
    stroke-dasharray: 2600;
    stroke-dashoffset: -2600;
  }

  #inner {
    stroke-dasharray: 1500;
    stroke-dashoffset: 1500;
  }

  @keyframes animate-layer {
    0% {
      stroke-opacity: 0.5;
    }
    100% {
      stroke-opacity: 1;
      stroke-dashoffset: 0;
    }
  }
</style>