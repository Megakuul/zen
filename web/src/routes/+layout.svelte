<script>
  import '../app.css';
  import favicon from '$lib/assets/favicon.svg';
  import Nav from './Nav.svelte';
  import { Toaster } from 'svelte-sonner';

  let { children } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<Toaster
  position="top-center"
  toastOptions={{
    unstyled: true,
    classes: {
      toast: 'dense-glass bg-slate-800 rounded-2xl p-4 flex flex-row gap-2 items-center',
    },
  }}
></Toaster>

<div class="flex overflow-hidden relative flex-col items-center p-4 mt-10 sm:mt-24">
  {@render children?.()}
</div>

<Nav></Nav>

<footer class="flex flex-col gap-2 justify-center items-center mt-20 mb-40 text-xs sm:text-2xl">
  <p>
    <a class="font-bold" href="https://github.com/megakuul/zen">Opensource Project</a> made with ⏳
    and ❤️ by <a class="font-bold" href="https://github.com/megakuul">Megakuul</a>
  </p>
  <p>
    <a class="underline text-slate-300/30" href="/privacy-policy">Privacy Policy & Legal Notice</a>
  </p>
</footer>

