<script>
  import { cn } from "$lib/utils/cn";
  import { inview } from "svelte-inview";
  import { Motion } from "svelte-motion";
    import PullupText from "./PullupText.svelte";

  /**
   * @type {{
   *  class?: string
   * }}
   */
  const { 
    class: className = undefined,
  } = $props();
</script>

<div class="glass flex justify-center items-center p-20 rounded-2xl">
  <PullupText text="24005231234asdf\nasdf"></PullupText>
  <PullupText text="24005231234"></PullupText>
</div>
