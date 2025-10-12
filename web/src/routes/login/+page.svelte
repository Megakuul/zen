<script>
  import { onMount } from 'svelte';
  import { GetToken, Login } from '$lib/client/auth.svelte';
  import { goto } from '$app/navigation';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { fade } from 'svelte/transition';

  let sent = $state(false);

  /** @param {string} verifier */
  async function login(verifier) {
    // TODO: wrap this into a sonner error wrapper
    try {
      await Login(verifier)
      sent = true;
    } catch (e) {
      const err = ConnectError.from(e)
      // TODO: emit a real error here via sandwicher or somethign
      console.error(err.cause)
    }
  }

  onMount(async () => {
    try {
      if (await GetToken()) goto("/profile")
    } catch (e) {
      const err = ConnectError.from(e)
      if (err.code !== Code.Unauthenticated) {
        // TODO: emit a real error here via sandwicher or somethign
        console.error(err.cause)
      } 
    }
  })

  let email = $state("");
  let code = $state("");
</script>

<div class="w-screen h-screen flex justify-center items-center">
  {#if sent}
    <div transition:fade class="glass p-10 rounded-2xl">
      <input bind:value={code} placeholder="Code (XXXX-XXXX)" />
      <button onclick={() => login(`code:${code}`)} class="flex flex-row justify-center items-center gap-4">
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVGp7eHRcaL)" d="M9 1a2 2 0 0 0-2 2H5.5A1.5 1.5 0 0 0 4 4.5V6a2 2 0 0 0 0 4v1.5A1.5 1.5 0 0 0 5.5 13H7a2 2 0 0 0 4 0h1.5a1.5 1.5 0 0 0 1.5-1.5V9h-1a1 1 0 1 1 0-2h1V4.5A1.5 1.5 0 0 0 12.5 3H11a2 2 0 0 0-2-2"/><path fill="url(#SVGEm7MQdlh)" fill-opacity="0.7" d="M9 1a2 2 0 0 0-2 2H5.5A1.5 1.5 0 0 0 4 4.5V6a2 2 0 0 0 0 4v1.5A1.5 1.5 0 0 0 5.5 13H7a2 2 0 0 0 4 0h1.5a1.5 1.5 0 0 0 1.5-1.5V9h-1a1 1 0 1 1 0-2h1V4.5A1.5 1.5 0 0 0 12.5 3H11a2 2 0 0 0-2-2"/><defs><linearGradient id="SVGp7eHRcaL" x1="4" x2="11.698" y1=".222" y2="14.886" gradientUnits="userSpaceOnUse"><stop stop-color="#1ec8b0"/><stop offset="1" stop-color="#2764e7"/></linearGradient><linearGradient id="SVGEm7MQdlh" x1="9.857" x2="13.049" y1="2.719" y2="16.315" gradientUnits="userSpaceOnUse"><stop offset=".533" stop-color="#ff6ce8" stop-opacity="0"/><stop offset="1" stop-color="#ff6ce8"/></linearGradient></defs></g></svg>        
        <span>Confirm code</span>
      </button>
    </div>
  {:else}
    <div transition:fade class="glass p-10 rounded-2xl flex flex-col items-center gap-2">
      <input bind:value={email} placeholder="Email (your.name@email.com)" />
      <button onclick={() => login(`email:${email}`)} class="w-full flex flex-row justify-center items-center gap-4 cursor-pointer p-3 rounded-xl hover:bg-slate-100/5 hover:scale-105 transition-all duration-700">
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVG09UKqbVr)" d="M8.805 8.958L1.994 11l.896-3l-.896-3l6.811 2.042c.95.285.95 1.63 0 1.916"/><path fill="url(#SVGH7ybqcCB)" d="M1.724 1.053a.5.5 0 0 0-.714.545l1.403 4.85a.5.5 0 0 0 .397.354l5.69.953c.268.053.268.437 0 .49l-5.69.953a.5.5 0 0 0-.397.354l-1.403 4.85a.5.5 0 0 0 .714.545l13-6.5a.5.5 0 0 0 0-.894z"/><path fill="url(#SVGUgu7sepB)" d="M1.724 1.053a.5.5 0 0 0-.714.545l1.403 4.85a.5.5 0 0 0 .397.354l5.69.953c.268.053.268.437 0 .49l-5.69.953a.5.5 0 0 0-.397.354l-1.403 4.85a.5.5 0 0 0 .714.545l13-6.5a.5.5 0 0 0 0-.894z"/><defs><linearGradient id="SVGH7ybqcCB" x1="1" x2="12.99" y1="-4.688" y2="11.244" gradientUnits="userSpaceOnUse"><stop stop-color="#3bd5ff"/><stop offset="1" stop-color="#0094f0"/></linearGradient><linearGradient id="SVGUgu7sepB" x1="8" x2="11.641" y1="4.773" y2="14.624" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#dcf8ff" stop-opacity="0"/><stop offset=".769" stop-color="#ff6ce8" stop-opacity="0.7"/></linearGradient><radialGradient id="SVG09UKqbVr" cx="0" cy="0" r="1" gradientTransform="matrix(7.43807 0 0 1.12359 .5 8)" gradientUnits="userSpaceOnUse"><stop stop-color="#0094f0"/><stop offset="1" stop-color="#2052cb"/></radialGradient></defs></g></svg>
        <span>Send verification code</span>
      </button>
    </div>
  {/if}
</div>