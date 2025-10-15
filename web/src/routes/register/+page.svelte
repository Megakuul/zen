<script>
  import { create } from '@bufbuild/protobuf';
  import { onMount } from 'svelte';
  import { GetToken } from '$lib/client/auth.svelte';
  import { goto } from '$app/navigation';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { fade } from 'svelte/transition';
  import Logo from '$lib/components/Logo.svelte';
  import { ManagementClient } from '$lib/client/client.svelte';
  import { UserSchema } from '$lib/sdk/v1/manager/user_pb';
  import { Exec } from '$lib/error/error.svelte';

  let sent = $state(false);
  let registered = $state(false);
  let loading = $state(false);

    /** 
   * @param {string} verifier
   */
  async function register(verifier) {
    await Exec(async () => {
      const response = await ManagementClient().register({
        user: user,
        captchaId: captchaId,
        captchaDigits: captchaCode,
        verifier: verifier,
      })
      if (captchaId && captchaCode) {
        sent = true;
      } else {
        captchaId = response.captchaId
        captchaBlob = URL.createObjectURL(new Blob(
          [new Uint8Array(response.captchaBlob)], 
          {type: "image/png"}
        ))
      }
    }, true, loading)
  }

  onMount(async () => {
    await Exec(async () => {
      if (await GetToken()) goto("/profile")
    }, false, loading)
  })

  /** @type {import("$lib/sdk/v1/manager/user_pb").User}*/
  let user = $state(create(UserSchema, {}))

  let captchaId = $state("");
  let captchaBlob = $state("");
  let captchaCode = $state("");
  let code = $state("");
</script>

<div class="w-screen h-screen flex justify-center items-center text-base sm:text-4xl">
  <div class="glass p-4 sm:p-10 rounded-2xl flex flex-col items-center gap-4 sm:gap-8">
    {#if registered}
      <svg class="w-24 h-24 sm:w-32 sm:h-32" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVGIbBqNcEI)" d="M2 8a6 6 0 1 1 12 0A6 6 0 0 1 2 8"/><path fill="url(#SVGrU4MlfYa)" d="M10.12 6.164L7.25 9.042L5.854 7.646a.5.5 0 1 0-.708.708l1.75 1.75a.5.5 0 0 0 .708 0l3.224-3.234a.5.5 0 0 0-.708-.706"/><defs><linearGradient id="SVGIbBqNcEI" x1="2.429" x2="10.71" y1="4.25" y2="12.854" gradientUnits="userSpaceOnUse"><stop stop-color="#52d17c"/><stop offset="1" stop-color="#22918b"/></linearGradient><linearGradient id="SVGrU4MlfYa" x1="6.12" x2="7.076" y1="6.449" y2="11.21" gradientUnits="userSpaceOnUse"><stop stop-color="#fff"/><stop offset="1" stop-color="#e3ffd9"/></linearGradient></defs></g></svg>
      <p class="text-base sm:text-3xl">Registration successful</p>
      <p class="text-base sm:text-3xl"><a href="/login" class="underline">Login</a> to continue</p>
    {:else}
      <Logo class="p-3 sm:p-6" svgClass="w-12 h-12 sm:w-20 sm:h-20"></Logo>
      <h1 class="text-xl sm:text-5xl font-bold text-slate-200/50">Zen Registration</h1>
      {#if sent}
        <input transition:fade bind:value={code} placeholder="Code (XXXX-XXXX)" class="glass text-center p-3 sm:p-5 rounded-xl focus:outline-0" />
        <button transition:fade onclick={() => register(`code:${code}`)} class="glass w-full flex flex-row justify-center items-center gap-4 cursor-pointer p-3 sm:p-4 rounded-xl hover:scale-105 transition-all duration-700">
          {#if loading}
            <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
          {:else}
            <span class="text-emerald-300/60">Confirm code</span>
            <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><g fill="none"><path fill="url(#SVGFnVq5bNm)" d="M30 16c0 7.732-6.268 14-14 14S2 23.732 2 16S8.268 2 16 2s14 6.268 14 14"/><path fill="url(#SVGcpquMdXZ)" d="M22.707 12.707a1 1 0 0 0-1.414-1.414L14.5 18.086l-3.293-3.293a1 1 0 0 0-1.414 1.414l4 4a1 1 0 0 0 1.414 0z"/><defs><linearGradient id="SVGFnVq5bNm" x1="3" x2="22.323" y1="7.25" y2="27.326" gradientUnits="userSpaceOnUse"><stop stop-color="#52d17c"/><stop offset="1" stop-color="#22918b"/></linearGradient><linearGradient id="SVGcpquMdXZ" x1="12.031" x2="14.162" y1="11.969" y2="22.66" gradientUnits="userSpaceOnUse"><stop stop-color="#fff"/><stop offset="1" stop-color="#e3ffd9"/></linearGradient></defs></g></svg>
          {/if}
      </button>
      {:else}
        <input transition:fade bind:value={user.username} placeholder="Username" class="glass p-3 sm:p-5 rounded-xl focus:outline-0" />
        <input transition:fade type="email" bind:value={user.email} placeholder="Email" class="glass p-3 sm:p-5 rounded-xl focus:outline-0" />
        {#if captchaBlob}
          <div class="flex flex-row items-center justify-between gap-4">
            <img src={captchaBlob} alt="captcha" />
            <input bind:value={captchaCode} placeholder="Captcha" class="glass p-3 sm:p-5 rounded-xl focus:outline-0" />
          </div>
        {/if}
        <button transition:fade onclick={() => register(`email:${user.email}`)} class="glass w-full flex flex-row justify-center items-center gap-4 cursor-pointer p-3 sm:p-4 rounded-xl hover:scale-105 transition-all duration-700">
          {#if loading}
            <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
          {:else}
            <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVG09UKqbVr)" d="M8.805 8.958L1.994 11l.896-3l-.896-3l6.811 2.042c.95.285.95 1.63 0 1.916"/><path fill="url(#SVGH7ybqcCB)" d="M1.724 1.053a.5.5 0 0 0-.714.545l1.403 4.85a.5.5 0 0 0 .397.354l5.69.953c.268.053.268.437 0 .49l-5.69.953a.5.5 0 0 0-.397.354l-1.403 4.85a.5.5 0 0 0 .714.545l13-6.5a.5.5 0 0 0 0-.894z"/><path fill="url(#SVGUgu7sepB)" d="M1.724 1.053a.5.5 0 0 0-.714.545l1.403 4.85a.5.5 0 0 0 .397.354l5.69.953c.268.053.268.437 0 .49l-5.69.953a.5.5 0 0 0-.397.354l-1.403 4.85a.5.5 0 0 0 .714.545l13-6.5a.5.5 0 0 0 0-.894z"/><defs><linearGradient id="SVGH7ybqcCB" x1="1" x2="12.99" y1="-4.688" y2="11.244" gradientUnits="userSpaceOnUse"><stop stop-color="#3bd5ff"/><stop offset="1" stop-color="#0094f0"/></linearGradient><linearGradient id="SVGUgu7sepB" x1="8" x2="11.641" y1="4.773" y2="14.624" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#dcf8ff" stop-opacity="0"/><stop offset=".769" stop-color="#ff6ce8" stop-opacity="0.7"/></linearGradient><radialGradient id="SVG09UKqbVr" cx="0" cy="0" r="1" gradientTransform="matrix(7.43807 0 0 1.12359 .5 8)" gradientUnits="userSpaceOnUse"><stop stop-color="#0094f0"/><stop offset="1" stop-color="#2052cb"/></radialGradient></defs></g></svg>
            <span class="text-blue-300/60">Send verification code</span>
          {/if}
        </button>
      {/if}
    {/if}
  </div>
</div>