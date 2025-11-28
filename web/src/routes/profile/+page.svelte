<script>
  import { onMount } from 'svelte';
  import { create } from '@bufbuild/protobuf';
  import { goto } from '$app/navigation';
  import { fade } from 'svelte/transition';
  import { Exec } from '$lib/error/error.svelte';
  import { ManagementClient } from '$lib/client/client.svelte';
  import { GetRequestSchema } from '$lib/sdk/v1/manager/management/management_pb';
  import { ClearToken } from '$lib/client/auth.svelte';

  let loading = $state(true);

  /** @type {import('$lib/sdk/v1/manager/user_pb').User|undefined}*/
  let user = $state();

  let edit = $state(false);

  let sentRevoke = $state(false);
  let revokeCode = $state('');

  onMount(async () => {
    await Exec(
      async () => {
        const response = await ManagementClient().get(create(GetRequestSchema, {}));
        user = response.user;
      },
      async () => {
        // explicitly clear in order that /login explicitly fetches a new token.
        // this is important on /profile because this is where /login routes if it believes to have a valid token.
        // so a non-expired token that causes 401 (e.g. after a jwks rotation) would create a loop.
        ClearToken();
        goto('/login');
      },
      processing => (loading = processing),
    );
  });
</script>

<div
  class="flex overflow-hidden flex-col gap-8 p-8 w-full text-base rounded-2xl sm:text-4xl glass min-h-[70vh] max-w-[1800px] sm:p-15"
>
  {#if user && edit}
    <input
      transition:fade
      bind:value={user.username}
      type="text"
      autocomplete="username"
      placeholder="Username"
      class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
    />
    <input
      transition:fade
      bind:value={user.email}
      type="email"
      autocomplete="email"
      placeholder="Email"
      class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
    />

    <button class="flex flex-row items-center glass">
      <span>Save Changes</span>
      <!-- prettier-ignore -->
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGLrQLkeeU)" d="M3 13c0-1.113.903-2 2.009-2H15a2 2 0 0 1 2 2c0 1.691-.833 2.966-2.135 3.797C13.583 17.614 11.855 18 10 18s-3.583-.386-4.865-1.203C3.833 15.967 3 14.69 3 13"/><path fill="url(#SVG8guycdmx)" d="M3 13c0-1.113.903-2 2.009-2H15a2 2 0 0 1 2 2c0 1.691-.833 2.966-2.135 3.797C13.583 17.614 11.855 18 10 18s-3.583-.386-4.865-1.203C3.833 15.967 3 14.69 3 13"/><path fill="url(#SVGx4nDzbQv)" fill-opacity="0.5" d="M3 13c0-1.113.903-2 2.009-2H15a2 2 0 0 1 2 2c0 1.691-.833 2.966-2.135 3.797C13.583 17.614 11.855 18 10 18s-3.583-.386-4.865-1.203C3.833 15.967 3 14.69 3 13"/><path fill="url(#SVGO9mfIeXx)" d="M10 2a4 4 0 1 0 0 8a4 4 0 0 0 0-8"/><path fill="url(#SVGOboKIefD)" d="M19 14.5a4.5 4.5 0 1 0-9 0a4.5 4.5 0 0 0 9 0"/><path fill="url(#SVG2dOOCdMt)" fill-rule="evenodd" d="M16.854 12.646a.5.5 0 0 1 0 .708l-3 3a.5.5 0 0 1-.708 0l-1-1a.5.5 0 0 1 .708-.708l.646.647l2.646-2.647a.5.5 0 0 1 .708 0" clip-rule="evenodd"/><defs><linearGradient id="SVGLrQLkeeU" x1="6.329" x2="8.591" y1="11.931" y2="19.153" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVG8guycdmx" x1="10" x2="13.167" y1="10.167" y2="22" gradientUnits="userSpaceOnUse"><stop stop-color="#885edb" stop-opacity="0"/><stop offset="1" stop-color="#e362f8"/></linearGradient><linearGradient id="SVGO9mfIeXx" x1="7.902" x2="11.979" y1="3.063" y2="9.574" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVGOboKIefD" x1="10.321" x2="16.532" y1="11.688" y2="18.141" gradientUnits="userSpaceOnUse"><stop stop-color="#52d17c"/><stop offset="1" stop-color="#22918b"/></linearGradient><linearGradient id="SVG2dOOCdMt" x1="12.938" x2="13.946" y1="12.908" y2="17.36" gradientUnits="userSpaceOnUse"><stop stop-color="#fff"/><stop offset="1" stop-color="#e3ffd9"/></linearGradient><radialGradient id="SVGx4nDzbQv" cx="0" cy="0" r="1" gradientTransform="rotate(90 -.5 15)scale(6.5)" gradientUnits="userSpaceOnUse"><stop offset=".423" stop-color="#30116e"/><stop offset="1" stop-color="#30116e" stop-opacity="0"/></radialGradient></defs></g></svg>
    </button>

    <button class="flex flex-row items-center glass">
      <span>Global Logout</span>
      <!-- prettier-ignore -->
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGM9XbKBKW)" d="M10 18a8 8 0 1 0 0-16a8 8 0 0 0 0 16"/><path fill="url(#SVGyl3FldAV)" fill-opacity="0.7" d="M10 18a8 8 0 1 0 0-16a8 8 0 0 0 0 16"/><path fill="url(#SVG55zqdpiA)" fill-opacity="0.2" d="M10 18a8 8 0 1 0 0-16a8 8 0 0 0 0 16"/><path fill="url(#SVGjdzh4SKb)" fill-rule="evenodd" d="M7.853 2.291a7 7 0 0 0-.816 1.51c-.368.906-.65 1.995-.826 3.199H2.58q-.195.485-.33 1h3.84a22 22 0 0 0 .001 4h-3.84q.135.515.33 1h3.63c.176 1.204.458 2.293.826 3.199a7 7 0 0 0 .816 1.51A8 8 0 0 0 10 18a8 8 0 0 0 2.147-.291a7 7 0 0 0 .816-1.51c.368-.906.65-1.995.826-3.199h3.63q.195-.485.329-1h-3.84a21.6 21.6 0 0 0 0-4h3.84a8 8 0 0 0-.33-1H13.79c-.176-1.204-.458-2.293-.826-3.199a7 7 0 0 0-.816-1.51A8 8 0 0 0 10 2a8 8 0 0 0-2.147.291M7.223 7c.166-1.076.42-2.035.74-2.822c.298-.733.642-1.292 1.003-1.66C9.324 2.153 9.672 2 10 2s.676.153 1.034.518c.36.368.705.927 1.003 1.66c.32.787.574 1.746.74 2.822zM10 18c.328 0 .676-.153 1.034-.518c.36-.368.705-.927 1.003-1.66c.32-.787.574-1.746.74-2.822H7.223c.167 1.076.421 2.035.741 2.822c.298.733.642 1.292 1.003 1.66c.358.365.706.518 1.034.518m-3-8c0 .692.033 1.362.096 2h5.808A21 21 0 0 0 13 10c0-.692-.033-1.362-.096-2H7.096A21 21 0 0 0 7 10" clip-rule="evenodd"/><path fill="url(#SVGbdZWxIYF)" d="M14 18.524c-1.175-.603-2.97-1.945-3-4.524v-2.562c0-.277.225-.497.499-.536c1.37-.193 2.485-1.134 3.066-1.725A.6.6 0 0 1 15 9c.16 0 .32.059.435.177c.58.591 1.696 1.532 3.066 1.725c.274.039.499.26.499.536V14c-.03 2.579-1.825 3.921-3 4.524a6.5 6.5 0 0 1-.87.372a.4.4 0 0 1-.26 0a6.5 6.5 0 0 1-.87-.372"/><defs><radialGradient id="SVGyl3FldAV" cx="0" cy="0" r="1" gradientTransform="rotate(180 7.469 7.063)scale(6.72116)" gradientUnits="userSpaceOnUse"><stop stop-color="#003580"/><stop offset="1" stop-color="#003580" stop-opacity="0"/></radialGradient><radialGradient id="SVG55zqdpiA" cx="0" cy="0" r="1" gradientTransform="matrix(8.37502 6.37502 -6.91548 9.08503 15.25 14)" gradientUnits="userSpaceOnUse"><stop offset=".412" stop-color="#1b44b1"/><stop offset="1" stop-color="#1b44b1" stop-opacity="0"/></radialGradient><radialGradient id="SVGjdzh4SKb" cx="0" cy="0" r="1" gradientTransform="matrix(8.49999 8.5048 -9.08479 9.07965 15 14.002)" gradientUnits="userSpaceOnUse"><stop offset=".445" stop-color="#3bd5ff" stop-opacity="0"/><stop offset=".815" stop-color="#3bd5ff"/></radialGradient><linearGradient id="SVGM9XbKBKW" x1="5.556" x2="17.111" y1="4.667" y2="15.333" gradientUnits="userSpaceOnUse"><stop stop-color="#29c3ff"/><stop offset="1" stop-color="#2052cb"/></linearGradient><linearGradient id="SVGbdZWxIYF" x1="12.5" x2="20.018" y1="9" y2="17.701" gradientUnits="userSpaceOnUse"><stop stop-color="#62be55"/><stop offset="1" stop-color="#1e794a"/></linearGradient></defs></g></svg>
    </button>

    <button class="flex flex-row items-center glass">
      <span>Delete Account</span>
      <!-- prettier-ignore -->
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVG726rTbBX)" d="M8 2a6 6 0 1 1 0 12A6 6 0 0 1 8 2"/><path fill="url(#SVGsV4Dceil)" d="M5.896 5.896a.5.5 0 0 1 .638-.057l.07.057L8 7.293l1.396-1.397a.5.5 0 0 1 .638-.057l.07.057a.5.5 0 0 1 .057.638l-.057.07L8.707 8l1.397 1.396a.5.5 0 0 1 .057.638l-.057.07a.5.5 0 0 1-.638.057l-.07-.057L8 8.707l-1.396 1.397a.5.5 0 0 1-.638.057l-.07-.057a.5.5 0 0 1-.057-.638l.057-.07L7.293 8L5.896 6.604a.5.5 0 0 1-.057-.638z"/><defs><linearGradient id="SVG726rTbBX" x1="3.875" x2="13" y1="2.75" y2="16" gradientUnits="userSpaceOnUse"><stop stop-color="#f83f54"/><stop offset="1" stop-color="#ca2134"/></linearGradient><linearGradient id="SVGsV4Dceil" x1="6.011" x2="8.354" y1="8.199" y2="10.635" gradientUnits="userSpaceOnUse"><stop stop-color="#fdfdfd"/><stop offset="1" stop-color="#fecbe6"/></linearGradient></defs></g></svg>
    </button>
  {:else if user}
    <h1 class="w-full text-3xl font-bold sm:text-6xl">{user.username}</h1>
    <h2>{user.description}</h2>
    <h2>{user.score}</h2>
    <h2>{user.streak}</h2>

    <button onclick={() => (edit = true)} class="flex flex-row items-center glass">
      <span>Edit Profile</span>
      <!-- prettier-ignore -->
      <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32"><g fill="none"><path fill="url(#SVGheAWrbDq)" d="M25.125 18.056v4.166l-8.095 7.75q-.51.027-1.03.028c-3.198 0-6.14-.823-8.315-2.207C5.523 26.417 4 24.393 4 22v-.5A3.5 3.5 0 0 1 7.5 18h17q.32.001.625.056"/><path fill="url(#SVGmPgBWcaU)" d="M25.125 18.056v4.166l-8.095 7.75q-.51.027-1.03.028c-3.198 0-6.14-.823-8.315-2.207C5.523 26.417 4 24.393 4 22v-.5A3.5 3.5 0 0 1 7.5 18h17q.32.001.625.056"/><path fill="url(#SVGXgoIjeqi)" fill-opacity="0.75" d="M25.125 18.056v4.166l-8.095 7.75q-.51.027-1.03.028c-3.198 0-6.14-.823-8.315-2.207C5.523 26.417 4 24.393 4 22v-.5A3.5 3.5 0 0 1 7.5 18h17q.32.001.625.056"/><path fill="url(#SVG1b9p8cSA)" fill-opacity="0.75" d="M25.125 18.056v4.166l-8.095 7.75q-.51.027-1.03.028c-3.198 0-6.14-.823-8.315-2.207C5.523 26.417 4 24.393 4 22v-.5A3.5 3.5 0 0 1 7.5 18h17q.32.001.625.056"/><path fill="url(#SVGbIOE1dGv)" d="M16 16a7 7 0 1 0 0-14a7 7 0 0 0 0 14"/><path fill="url(#SVGNAvoddZR)" d="m21.539 29.469l7.61-7.543v-4.074h-4.073l-7.567 7.64l.308 3.695z"/><path fill="url(#SVGG35LqbVt)" d="m21.539 29.47l.223-.223s-1.726-.661-2.535-1.47c-.809-.81-1.47-2.534-1.47-2.534l-.248.249a2.66 2.66 0 0 0-.686 1.206l-.79 3.051a1 1 0 0 0 1.217 1.22l3.02-.778a2.8 2.8 0 0 0 1.269-.722"/><path fill="url(#SVG0BYmyspL)" d="m27.937 23.14l2.211-2.214a2.88 2.88 0 0 0 .072-4.017a2.88 2.88 0 0 0-4.144-.057l-2.238 2.241z"/><path fill="url(#SVGpCHUpdwg)" d="M25.094 17.838a5.43 5.43 0 0 0 4.106 4.038l-1.55 1.551a5.43 5.43 0 0 1-4.106-4.04z"/><defs><linearGradient id="SVGheAWrbDq" x1="9.707" x2="13.584" y1="19.595" y2="31.977" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVGmPgBWcaU" x1="16" x2="21.429" y1="16.571" y2="36.857" gradientUnits="userSpaceOnUse"><stop stop-color="#885edb" stop-opacity="0"/><stop offset="1" stop-color="#e362f8"/></linearGradient><linearGradient id="SVGbIOE1dGv" x1="12.329" x2="19.464" y1="3.861" y2="15.254" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVGNAvoddZR" x1="20.861" x2="27.044" y1="19.948" y2="26.149" gradientUnits="userSpaceOnUse"><stop stop-color="#ffa43d"/><stop offset="1" stop-color="#fb5937"/></linearGradient><linearGradient id="SVGG35LqbVt" x1="15.174" x2="19.325" y1="26.847" y2="30.975" gradientUnits="userSpaceOnUse"><stop offset=".255" stop-color="#ffd394"/><stop offset="1" stop-color="#ff921f"/></linearGradient><linearGradient id="SVG0BYmyspL" x1="29.502" x2="26.869" y1="17.485" y2="19.969" gradientUnits="userSpaceOnUse"><stop stop-color="#f97dbd"/><stop offset="1" stop-color="#dd3ce2"/></linearGradient><linearGradient id="SVGpCHUpdwg" x1="26.469" x2="22.489" y1="21.664" y2="19.902" gradientUnits="userSpaceOnUse"><stop stop-color="#ff921f"/><stop offset="1" stop-color="#ffe994"/></linearGradient><radialGradient id="SVGXgoIjeqi" cx="0" cy="0" r="1" gradientTransform="matrix(9.6823 -16.28579 8.42641 5.0097 20.724 34.286)" gradientUnits="userSpaceOnUse"><stop stop-color="#0a1852" stop-opacity="0.75"/><stop offset="1" stop-color="#0a1852" stop-opacity="0"/></radialGradient><radialGradient id="SVG1b9p8cSA" cx="0" cy="0" r="1" gradientTransform="matrix(0 -5.5 7.25 0 26 21.5)" gradientUnits="userSpaceOnUse"><stop stop-color="#0a1852" stop-opacity="0.75"/><stop offset="1" stop-color="#0a1852" stop-opacity="0"/></radialGradient></defs></g></svg>
    </button>
  {:else if loading}
    <div class="w-full h-20 rounded-2xl animate-pulse bg-slate-500/10"></div>
    <div class="w-full h-20 rounded-2xl animate-pulse bg-slate-500/10"></div>
    <div class="w-full h-20 rounded-2xl animate-pulse bg-slate-500/10"></div>
  {:else}
    <h1 class="flex flex-row gap-4 items-center text-xl font-bold sm:text-3xl lg:text-6xl">
      <span>An Error occurred</span>
      <!-- prettier-ignore -->
      <svg class="w-8 h-8 lg:w-16 lg:h-16" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g fill="none"><path fill="url(#SVGNCeuCeAB)" d="M10.03 3.659c.856-1.548 3.081-1.548 3.937 0l7.746 14.001c.83 1.5-.255 3.34-1.969 3.34H4.254c-1.715 0-2.8-1.84-1.97-3.34z"/><path fill="url(#SVGHqzv3c0k)" d="M12.997 17A.999.999 0 1 0 11 17a.999.999 0 0 0 1.997 0m-.259-7.852a.75.75 0 0 0-1.493.103l.004 4.501l.007.102a.75.75 0 0 0 1.493-.103l-.004-4.502z"/><defs><linearGradient id="SVGNCeuCeAB" x1="5.125" x2="16.719" y1="-.393" y2="23.477" gradientUnits="userSpaceOnUse"><stop stop-color="#ffcd0f"/><stop offset="1" stop-color="#fe8401"/></linearGradient><linearGradient id="SVGHqzv3c0k" x1="9.336" x2="13.752" y1="8.5" y2="18.405" gradientUnits="userSpaceOnUse"><stop stop-color="#4a4a4a"/><stop offset="1" stop-color="#212121"/></linearGradient></defs></g></svg>
    </h1>
    <p>
      Please reload the page and ensure you are <a class="underline" href="/login">Logged in</a>
    </p>
    <div class="w-full h-20 rounded-2xl animate-pulse bg-slate-500/10"></div>
    <div class="w-full h-20 rounded-2xl animate-pulse bg-slate-500/10"></div>
    <div class="w-full h-20 rounded-2xl animate-pulse bg-slate-500/10"></div>
  {/if}
</div>

