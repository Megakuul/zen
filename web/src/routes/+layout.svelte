<script>
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
  import Nav from './Nav.svelte';
	
	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="relative flex flex-col items-center p-4 mt-10 sm:mt-24 overflow-hidden">
	{@render children?.()}
</div>

<Nav></Nav>

<footer class="flex justify-center text-xs sm:text-2xl my-20">
  <p>
    <a class="font-bold" href="https://github.com/megakuul/zen">Opensource Project</a> made with ⏳ and ❤️ by <a class="font-bold" href="https://github.com/megakuul">Megakuul</a>
  </p>
</footer>