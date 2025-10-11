<script>
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	
	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="flex flex-col items-center p-4 mt-10 sm:mt-24 overflow-hidden">
	{@render children?.()}
</div>

<footer class="flex justify-center text-xs sm:text-2xl my-20">
	<a class="font-bold" href="https://github.com/megakuul/zen">Opensource Project</a>&nbsp;
	made with ⏳ and ❤️ by 
		&nbsp;<a class="font-bold" href="https://github.com/megakuul">Megakuul</a>
</footer>