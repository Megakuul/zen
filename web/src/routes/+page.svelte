<script>
  import TextGenerateEffect from "$lib/components/TextGenerateEffect.svelte";
  import Logo from "../lib/components/Logo.svelte";
</script>

<div class="flex flex-col items-center mt-24">
  <Logo></Logo>
  <TextGenerateEffect class="flex flex-col items-center w-9/12 text-center mt-20 lg:mt-32 2xl:mt-36" durationPerWord={2} segments={[
    {text: "« The first and greatest victory is to conquer yourself »", class: "text-xl lg:text-3xl xl:text-5xl 2xl:text-6xl font-medium text-slate-100/60"},
  ]}></TextGenerateEffect>
</div>