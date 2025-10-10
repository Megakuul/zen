<script>
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	
	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children?.()}

<footer class="flex justify-center text-xs sm:text-2xl my-20">
	<a class="font-bold" href="https://github.com/megakuul/zen">Opensource Project</a>&nbsp;
	made with ⏳ and ❤️ by 
		&nbsp;<a class="font-bold" href="https://github.com/megakuul">Megakuul</a>
</footer>