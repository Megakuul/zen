<script>
  import { onMount } from 'svelte';
  import { GetToken, Login } from '$lib/client/auth.svelte';
  import { goto } from '$app/navigation';
  import { fade } from 'svelte/transition';
  import { create } from '@bufbuild/protobuf';
  import Logo from '$lib/components/Logo.svelte';
  import { Exec } from '$lib/error/error.svelte';
  import { VerifierSchema, VerifierStage } from '$lib/sdk/v1/manager/verifier_pb';
  import { Code, ConnectError } from '@connectrpc/connect';

  let sent = $state(false);
  let loading = $state(false);

  onMount(async () => {
    await Exec(
      async () => {
        if (await GetToken()) goto('/profile');
      },
      async () => {
        /* ignore unauth errors */
      },
      processing => (loading = processing),
    );
  });

  let email = $state('');
  let code = $state('');
</script>

<div class="flex justify-center items-center w-screen text-base sm:text-4xl h-dvh">
  <div class="flex flex-col gap-4 items-center p-4 rounded-2xl sm:gap-8 sm:p-10 glass">
    <Logo class="p-3 sm:p-6" svgClass="w-12 h-12 sm:w-20 sm:h-20"></Logo>
    <h1 class="text-xl font-bold sm:text-5xl text-slate-200/50">Zen Login</h1>
    {#if sent}
      <input
        transition:fade
        bind:value={code}
        autocomplete="one-time-code"
        placeholder="Code (XXXX-XXXX)"
        class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
      />
      <button
        transition:fade
        onclick={() =>
          Exec(
            async () => {
              await Login(
                create(VerifierSchema, { email: email, code: code, stage: VerifierStage.CODE }),
              );
              goto('/profile');
            },
            undefined,
            processing => (loading = processing),
          )}
        style={code === '' ? 'padding: 0px; height: 0px; opacity: 0;' : ''}
        class="flex overflow-hidden flex-row gap-4 justify-center items-center p-3 w-full h-12 rounded-xl transition-all duration-700 cursor-pointer sm:p-4 sm:h-24 hover:scale-105 glass"
      >
        {#if loading}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"> <g stroke="currentColor" stroke-width="1"> <circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"> <animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150" /><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59" /></circle> <animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12" /></g> </svg>
        {:else}
          <span class="text-emerald-300/60">Confirm code</span>
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"> <g fill="none"> <path fill="url(#SVGFnVq5bNm)" d="M30 16c0 7.732-6.268 14-14 14S2 23.732 2 16S8.268 2 16 2s14 6.268 14 14" /><path fill="url(#SVGcpquMdXZ)" d="M22.707 12.707a1 1 0 0 0-1.414-1.414L14.5 18.086l-3.293-3.293a1 1 0 0 0-1.414 1.414l4 4a1 1 0 0 0 1.414 0z" /><defs> <linearGradient id="SVGFnVq5bNm" x1="3" x2="22.323" y1="7.25" y2="27.326" gradientUnits="userSpaceOnUse"> <stop stop-color="#52d17c" /><stop offset="1" stop-color="#22918b" /></linearGradient> <linearGradient id="SVGcpquMdXZ" x1="12.031" x2="14.162" y1="11.969" y2="22.66" gradientUnits="userSpaceOnUse"> <stop stop-color="#fff" /><stop offset="1" stop-color="#e3ffd9" /></linearGradient> </defs> </g> </svg>
        {/if}
      </button>
    {:else}
      <input
        type="email"
        bind:value={email}
        placeholder="Email"
        class="p-3 rounded-xl sm:p-5 glass focus:outline-0"
      />
      <button
        onclick={() =>
          Exec(
            async () => {
              try {
                await Login(create(VerifierSchema, { email: email, stage: VerifierStage.EMAIL }));
              } catch (e) {
                const err = ConnectError.from(e);
                if (err.code === Code.AlreadyExists)
                  sent = true; // code already sent
                else throw e;
              }
              sent = true;
            },
            undefined,
            processing => (loading = processing),
          )}
        style={email === '' ? 'padding: 0px; height: 0px; opacity: 0;' : ''}
        class="flex overflow-hidden flex-row gap-4 justify-center items-center p-3 w-full h-12 rounded-xl transition-all duration-700 cursor-pointer sm:p-4 sm:h-24 hover:scale-105 glass"
      >
        {#if loading}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150" /><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59" /></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12" /></g></svg>
        {:else}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"> <path fill="url(#SVG09UKqbVr)" d="M8.805 8.958L1.994 11l.896-3l-.896-3l6.811 2.042c.95.285.95 1.63 0 1.916" /><path fill="url(#SVGH7ybqcCB)" d="M1.724 1.053a.5.5 0 0 0-.714.545l1.403 4.85a.5.5 0 0 0 .397.354l5.69.953c.268.053.268.437 0 .49l-5.69.953a.5.5 0 0 0-.397.354l-1.403 4.85a.5.5 0 0 0 .714.545l13-6.5a.5.5 0 0 0 0-.894z" /><path fill="url(#SVGUgu7sepB)" d="M1.724 1.053a.5.5 0 0 0-.714.545l1.403 4.85a.5.5 0 0 0 .397.354l5.69.953c.268.053.268.437 0 .49l-5.69.953a.5.5 0 0 0-.397.354l-1.403 4.85a.5.5 0 0 0 .714.545l13-6.5a.5.5 0 0 0 0-.894z" /><defs> <linearGradient id="SVGH7ybqcCB" x1="1" x2="12.99" y1="-4.688" y2="11.244" gradientUnits="userSpaceOnUse"> <stop stop-color="#3bd5ff" /><stop offset="1" stop-color="#0094f0" /></linearGradient> <linearGradient id="SVGUgu7sepB" x1="8" x2="11.641" y1="4.773" y2="14.624" gradientUnits="userSpaceOnUse"> <stop offset=".125" stop-color="#dcf8ff" stop-opacity="0" /><stop offset=".769" stop-color="#ff6ce8" stop-opacity="0.7" /></linearGradient> <radialGradient id="SVG09UKqbVr" cx="0" cy="0" r="1" gradientTransform="matrix(7.43807 0 0 1.12359 .5 8)" gradientUnits="userSpaceOnUse"> <stop stop-color="#0094f0" /><stop offset="1" stop-color="#2052cb" /></radialGradient> </defs> </g> </svg>
          <span class="text-blue-300/60">Send verification code</span>
        {/if}
      </button>
      <p class="text-xs sm:text-base text-slate-50/40">
        No account yet? register <a href="/register" class="underline">here</a>
      </p>
    {/if}
  </div>
</div>
