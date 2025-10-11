<script>
  import AgentTable from "$lib/components/AgentTable.svelte";
  import Ripple from "$lib/components/Ripple.svelte";
  import TextGenerateEffect from "$lib/components/TextGenerateEffect.svelte";
  import Logo from "../lib/components/Logo.svelte";
</script>

<Logo></Logo>

<TextGenerateEffect class="flex flex-col items-center w-9/12 text-center mt-10 sm:mt-20 lg:mt-32 2xl:mt-36" durationPerWord={2} segments={[
  {text: "« The first and greatest victory is to conquer yourself »", class: "text-base lg:text-3xl xl:text-5xl 2xl:text-6xl font-medium text-slate-100/60"},
]}></TextGenerateEffect>

<svg class="slow-arrow w-24 h-24 mt-24 sm:mt-64 text-slate-50/30" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m12 20l6-6m-6 6l-6-6m6 6V9.5M12 4v2.5"/></svg>

<Ripple href="/getting-started" class="mt-24 sm:mt-64"></Ripple>

<TextGenerateEffect class="mt-24 sm:mt-64" durationPerWord={1} segments={[
  {text: "Not convinced yet?", class: "text-base lg:text-2xl xl:text-3xl 2xl:text-4xl font-medium text-slate-100/60"}
]}></TextGenerateEffect>

<svg class="arrow w-24 h-24 mt-10 sm:mt-40 text-slate-50/30" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="m12 20l6-6m-6 6l-6-6m6 6V9.5M12 4v2.5"/></svg>

<AgentTable class="mt-24 sm:mt-64"></AgentTable>

<style>
  .arrow {
    animation: fade-in 2s forwards linear, wave-down 4s infinite ease;
  }

  .slow-arrow {
    animation: fade-in 8s forwards linear, wave-down 4s infinite ease;
  }

  @keyframes fade-in {
    0%, 50% {
      opacity: 0;
    }
    100% {
      opacity: 1;
    }
  }

  @keyframes wave-down {
    0%, 100% {
      transform: translateY(0);
    }
    50% {
      transform: translateY(15px);
    }
  }
</style>