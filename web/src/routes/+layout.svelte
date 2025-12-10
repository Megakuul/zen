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
