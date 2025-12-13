<script>
  import { onMount } from 'svelte';
  import { create } from '@bufbuild/protobuf';
  import { goto } from '$app/navigation';
  import { Exec } from '$lib/error/error.svelte';
  import { ManagementClient } from '$lib/client/client.svelte';
  import {
    DeleteRequestSchema,
    GetRequestSchema,
    UpdateRequestSchema,
  } from '$lib/sdk/v1/manager/management/management_pb';
  import { ClearToken, Logout } from '$lib/client/auth.svelte';
  import { VerifierStage } from '$lib/sdk/v1/manager/verifier_pb';
  import { GetScoreTextDecorator } from '$lib/color/color';
  import Streak from '$lib/components/Streak.svelte';
  import { Code, ConnectError } from '@connectrpc/connect';
  import { browser } from '$app/environment';
  import { EventType } from '$lib/sdk/v1/scheduler/event_pb';
  import EventTypeIcon from '$lib/components/EventTypeIcon.svelte';

  let loading = $state(true);

  /** @type {import('$lib/sdk/v1/manager/user_pb').User|undefined}*/
  let user = $state();

  /** @type {{name: string, id: import('$lib/sdk/v1/scheduler/event_pb').EventType, url: string}[]} */
  let eventTypes = $state([]);

  let dayStartHour = $state(6);

  $effect.root(() => {
    if (!browser) {
      return;
    }

    dayStartHour = Number(localStorage.getItem(`default_day_start`) ?? dayStartHour);

    eventTypes = [];
    // remove typescript enum double mapping (only keep string keys)
    const cleanEventTypes = Object.entries(EventType).filter(([key]) => isNaN(Number(key)));
    for (const [name, id] of cleanEventTypes) {
      const url = localStorage.getItem(`default_music_${id}`);
      eventTypes.push({
        name: name,
        id: Number(id),
        url: url ?? '',
      });
    }
  });

  let edit = $state(false);

  let sentDelete = $state(false);
  let deleteCode = $state('');
  let deleteConfirmation = $state(false);

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

<svelte:head>
  <title>Profile | Zen</title>
  <link rel="canonical" href="https://zen.megakuul.com/profile" />
  <meta property="og:title" content="Zen Profile" />
  <meta property="og:type" content="website" />
  <meta property="og:image" content="https://zen.megakuul.com/favicon.svg" />
</svelte:head>

<div
  class="flex flex-col gap-3 p-8 w-full text-base rounded-2xl sm:gap-4 sm:text-4xl overflow-scroll-hidden glass h-[85dvh] max-w-[1800px] sm:p-15"
>
  {#if user && edit}
    <input
      bind:value={user.username}
      type="text"
      placeholder="Username"
      class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
    />
    <input
      bind:value={user.description}
      type="text"
      placeholder="Statement"
      class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
    />

    <label class="flex flex-row gap-3 justify-center items-center py-2">
      <span class="text-xs sm:max-w-full sm:text-base lg:text-xl max-w-48">
        I want to share my activity stats on the
        <span class="font-bold">public leaderboard</span>
      </span>
      <input bind:checked={user.leaderboard} type="checkbox" class="w-3 h-3 sm:w-5 sm:h-5" />
    </label>

    <div class="flex flex-col gap-3 items-center">
      {#each eventTypes as type}
        <div class="flex flex-row gap-3 items-center sm:gap-4">
          <EventTypeIcon type={type.id} />
          <input
            bind:value={type.url}
            type="url"
            placeholder="music url (spotify, ...)"
            class="p-1 text-xs rounded-lg sm:p-3 sm:rounded-xl lg:text-xl glass focus:outline-0"
          />
        </div>
      {/each}
      <label class="flex flex-row gap-3 justify-center items-center py-2">
        <span class="text-xs sm:max-w-full sm:text-base lg:text-xl max-w-48">
          Calendar day starts at
        </span>
        <input
          bind:value={dayStartHour}
          type="number"
          class="p-1 text-xs text-center rounded-lg sm:p-3 sm:rounded-xl lg:text-xl max-w-12 glass focus:outline-0"
        />
        <span class="text-xs sm:max-w-full sm:text-base lg:text-xl max-w-48"> o'clock </span>
      </label>
    </div>

    <div class="flex flex-col gap-3 items-center mt-auto w-full sm:flex-row sm:gap-4">
      <button
        onclick={() => (edit = false)}
        class="flex flex-row gap-2 justify-center items-center p-2 w-full rounded-xl transition-all duration-700 cursor-pointer hover:scale-105 glass"
      >
        <span>Cancel</span>
        <!-- prettier-ignore -->
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 sm:w-8 sm:h-8" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVG726rTbBX)" d="M8 2a6 6 0 1 1 0 12A6 6 0 0 1 8 2"/><path fill="url(#SVGsV4Dceil)" d="M5.896 5.896a.5.5 0 0 1 .638-.057l.07.057L8 7.293l1.396-1.397a.5.5 0 0 1 .638-.057l.07.057a.5.5 0 0 1 .057.638l-.057.07L8.707 8l1.397 1.396a.5.5 0 0 1 .057.638l-.057.07a.5.5 0 0 1-.638.057l-.07-.057L8 8.707l-1.396 1.397a.5.5 0 0 1-.638.057l-.07-.057a.5.5 0 0 1-.057-.638l.057-.07L7.293 8L5.896 6.604a.5.5 0 0 1-.057-.638z"/><defs><linearGradient id="SVG726rTbBX" x1="3.875" x2="13" y1="2.75" y2="16" gradientUnits="userSpaceOnUse"><stop stop-color="#f83f54"/><stop offset="1" stop-color="#ca2134"/></linearGradient><linearGradient id="SVGsV4Dceil" x1="6.011" x2="8.354" y1="8.199" y2="10.635" gradientUnits="userSpaceOnUse"><stop stop-color="#fdfdfd"/><stop offset="1" stop-color="#fecbe6"/></linearGradient></defs></g></svg>
      </button>

      <button
        onclick={async () =>
          Exec(
            async () => {
              if (browser) localStorage.setItem(`default_day_start`, dayStartHour.toString());
              for (const type of eventTypes) {
                if (browser) localStorage.setItem(`default_music_${type.id}`, type.url);
              }
              await ManagementClient().update(create(UpdateRequestSchema, { user: user }));
              edit = false;
            },
            undefined,
            processing => (loading = processing),
          )}
        class="flex flex-row gap-2 justify-center items-center p-2 w-full rounded-xl transition-all duration-700 cursor-pointer hover:scale-105 glass"
      >
        {#if loading}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
        {:else}
          <span>Save Changes</span>
          <!-- prettier-ignore -->
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 sm:w-8 sm:h-8" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGLrQLkeeU)" d="M3 13c0-1.113.903-2 2.009-2H15a2 2 0 0 1 2 2c0 1.691-.833 2.966-2.135 3.797C13.583 17.614 11.855 18 10 18s-3.583-.386-4.865-1.203C3.833 15.967 3 14.69 3 13"/><path fill="url(#SVG8guycdmx)" d="M3 13c0-1.113.903-2 2.009-2H15a2 2 0 0 1 2 2c0 1.691-.833 2.966-2.135 3.797C13.583 17.614 11.855 18 10 18s-3.583-.386-4.865-1.203C3.833 15.967 3 14.69 3 13"/><path fill="url(#SVGx4nDzbQv)" fill-opacity="0.5" d="M3 13c0-1.113.903-2 2.009-2H15a2 2 0 0 1 2 2c0 1.691-.833 2.966-2.135 3.797C13.583 17.614 11.855 18 10 18s-3.583-.386-4.865-1.203C3.833 15.967 3 14.69 3 13"/><path fill="url(#SVGO9mfIeXx)" d="M10 2a4 4 0 1 0 0 8a4 4 0 0 0 0-8"/><path fill="url(#SVGOboKIefD)" d="M19 14.5a4.5 4.5 0 1 0-9 0a4.5 4.5 0 0 0 9 0"/><path fill="url(#SVG2dOOCdMt)" fill-rule="evenodd" d="M16.854 12.646a.5.5 0 0 1 0 .708l-3 3a.5.5 0 0 1-.708 0l-1-1a.5.5 0 0 1 .708-.708l.646.647l2.646-2.647a.5.5 0 0 1 .708 0" clip-rule="evenodd"/><defs><linearGradient id="SVGLrQLkeeU" x1="6.329" x2="8.591" y1="11.931" y2="19.153" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVG8guycdmx" x1="10" x2="13.167" y1="10.167" y2="22" gradientUnits="userSpaceOnUse"><stop stop-color="#885edb" stop-opacity="0"/><stop offset="1" stop-color="#e362f8"/></linearGradient><linearGradient id="SVGO9mfIeXx" x1="7.902" x2="11.979" y1="3.063" y2="9.574" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#9c6cfe"/><stop offset="1" stop-color="#7a41dc"/></linearGradient><linearGradient id="SVGOboKIefD" x1="10.321" x2="16.532" y1="11.688" y2="18.141" gradientUnits="userSpaceOnUse"><stop stop-color="#52d17c"/><stop offset="1" stop-color="#22918b"/></linearGradient><linearGradient id="SVG2dOOCdMt" x1="12.938" x2="13.946" y1="12.908" y2="17.36" gradientUnits="userSpaceOnUse"><stop stop-color="#fff"/><stop offset="1" stop-color="#e3ffd9"/></linearGradient><radialGradient id="SVGx4nDzbQv" cx="0" cy="0" r="1" gradientTransform="rotate(90 -.5 15)scale(6.5)" gradientUnits="userSpaceOnUse"><stop offset=".423" stop-color="#30116e"/><stop offset="1" stop-color="#30116e" stop-opacity="0"/></radialGradient></defs></g></svg>
        {/if}
      </button>
    </div>

    <div class="flex flex-col gap-3 items-center w-full sm:flex-row sm:gap-4">
      <button
        onclick={async () =>
          Exec(
            async () => {
              await Logout();
              goto('/login');
            },
            undefined,
            processing => (loading = processing),
          )}
        class="flex flex-row gap-2 justify-center items-center p-2 w-full rounded-xl transition-all duration-700 cursor-pointer hover:scale-105 glass"
      >
        {#if loading}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
        {:else}
          <span>Logout from Device</span>
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVGVvaXyd5Q)" d="M8 7a5 5 0 0 0-5 5c0 1.298 1.212 2 2.285 2h5.43C11.788 14 13 13.298 13 12a5 5 0 0 0-5-5"/><path fill="url(#SVGSTjazdBE)" d="M8.5 3.875C8.5 2.938 9.138 2 10.125 2s1.625.938 1.625 1.875s-.638 1.875-1.625 1.875S8.5 4.812 8.5 3.875m-6.125.375C1.388 4.25.75 5.188.75 6.125S1.388 8 2.375 8S4 7.062 4 6.125S3.362 4.25 2.375 4.25m11.25 0c-.987 0-1.625.938-1.625 1.875S12.638 8 13.625 8s1.625-.938 1.625-1.875s-.638-1.875-1.625-1.875M5.875 2c-.987 0-1.625.938-1.625 1.875S4.888 5.75 5.875 5.75S7.5 4.812 7.5 3.875S6.862 2 5.875 2"/><defs><radialGradient id="SVGSTjazdBE" cx="0" cy="0" r="1" gradientTransform="matrix(0 -7.71429 11.6 0 8.403 8.429)" gradientUnits="userSpaceOnUse"><stop stop-color="#eb4824"/><stop offset="1" stop-color="#ff921f"/></radialGradient><linearGradient id="SVGVvaXyd5Q" x1="5.378" x2="8.294" y1="7.931" y2="14.583" gradientUnits="userSpaceOnUse"><stop offset=".125" stop-color="#ff921f"/><stop offset="1" stop-color="#eb4824"/></linearGradient></defs></g></svg>
        {/if}
      </button>

      <button
        onclick={async () =>
          Exec(
            async () => {
              try {
                await ManagementClient().delete(
                  create(DeleteRequestSchema, {
                    verifier: { stage: VerifierStage.EMAIL, email: user?.email },
                  }),
                );
                sentDelete = true;
              } catch (e) {
                const err = ConnectError.from(e);
                if (err.code === Code.AlreadyExists)
                  sentDelete = true; // code already sent
                else throw e;
              }
            },
            undefined,
            processing => (loading = processing),
          )}
        class="flex flex-row gap-2 justify-center items-center p-2 w-full rounded-xl transition-all duration-700 cursor-pointer hover:scale-105 glass"
      >
        {#if loading}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
        {:else}
          <span>Delete Account</span>
          <!-- prettier-ignore -->
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 sm:w-8 sm:h-8" width="24" height="24" viewBox="0 0 24 24"><g fill="none"><path fill="url(#SVG1j4dmetI)" d="M4.5 21.25V15.5H3.007L3 21.25l.007.102A.75.75 0 0 0 3.75 22l.102-.007a.75.75 0 0 0 .648-.743"/><path fill="url(#SVGasxyxeCL)" d="M3.75 2.998a.75.75 0 0 0-.75.75V16a.5.5 0 0 0 .5.5h16.754a.75.75 0 0 0 .6-1.2L16.69 9.75l4.164-5.552a.75.75 0 0 0-.6-1.2z"/><defs><linearGradient id="SVG1j4dmetI" x1="4.5" x2="4.069" y1="24.089" y2="15.729" gradientUnits="userSpaceOnUse"><stop stop-color="#889096"/><stop offset="1" stop-color="#63686e"/></linearGradient><linearGradient id="SVGasxyxeCL" x1="-.939" x2="6.516" y1="-.86" y2="17.385" gradientUnits="userSpaceOnUse"><stop stop-color="#f97dbd"/><stop offset="1" stop-color="#d7257d"/></linearGradient></defs></g></svg>
        {/if}
      </button>
    </div>

    {#if sentDelete}
      <label class="flex flex-row gap-3 justify-start items-center">
        <span class="text-xs sm:max-w-full sm:text-base lg:text-xl max-w-48">
          I confirm that I want to delete my account and all associated data
        </span>
        <input bind:checked={deleteConfirmation} type="checkbox" class="w-3 h-3 sm:w-5 sm:h-5" />
      </label>

      <input
        bind:value={deleteCode}
        autocomplete="one-time-code"
        placeholder="Code (XXXX-XXXX)"
        class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
      />
      <button
        onclick={async () =>
          Exec(
            async () => {
              await ManagementClient().delete(
                create(DeleteRequestSchema, {
                  verifier: { stage: VerifierStage.CODE, email: user?.email, code: deleteCode },
                }),
              );
              await Logout();
              goto('/login');
            },
            undefined,
            processing => (loading = processing),
          )}
        style={deleteCode === '' || !deleteConfirmation
          ? 'padding: 0px; height: 0px; opacity: 0;'
          : ''}
        class="flex overflow-hidden flex-row gap-4 justify-center items-center p-3 w-full h-12 rounded-xl transition-all duration-700 cursor-pointer sm:p-4 sm:h-24 hover:scale-105 glass"
      >
        {#if loading}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
        {:else}
          <span class="text-orange-400/60">Confirm Deletion</span>
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="none"><path fill="url(#SVGNCeuCeAB)" d="M10.03 3.659c.856-1.548 3.081-1.548 3.937 0l7.746 14.001c.83 1.5-.255 3.34-1.969 3.34H4.254c-1.715 0-2.8-1.84-1.97-3.34z"/><path fill="url(#SVGHqzv3c0k)" d="M12.997 17A.999.999 0 1 0 11 17a.999.999 0 0 0 1.997 0m-.259-7.852a.75.75 0 0 0-1.493.103l.004 4.501l.007.102a.75.75 0 0 0 1.493-.103l-.004-4.502z"/><defs><linearGradient id="SVGNCeuCeAB" x1="5.125" x2="16.719" y1="-.393" y2="23.477" gradientUnits="userSpaceOnUse"><stop stop-color="#ffcd0f"/><stop offset="1" stop-color="#fe8401"/></linearGradient><linearGradient id="SVGHqzv3c0k" x1="9.336" x2="13.752" y1="8.5" y2="18.405" gradientUnits="userSpaceOnUse"><stop stop-color="#4a4a4a"/><stop offset="1" stop-color="#212121"/></linearGradient></defs></g></svg>
        {/if}
      </button>
    {/if}
  {:else if user}
    <h1 class="text-3xl font-bold sm:text-6xl">
      {user.username}&nbsp;
      <span class="text-xl sm:text-2xl">
        (<span class="text-slate-100/50">{user.email}</span>)
      </span>
    </h1>
    {#if user.description}
      <p class="text-base sm:text-3xl">
        <span class="font-serif text-xl font-bold sm:text-4xl">“</span>
        <span class="animate-pulse text-slate-200/30">{user.description}</span>
      </p>
    {/if}
    <div class="flex flex-row gap-4 justify-center items-center h-full">
      <p class="text-5xl font-bold text-center select-none sm:text-9xl">
        Score
        <span class={GetScoreTextDecorator(user.score)}>{user.score}</span>
      </p>
      {#if user.streak > 0}
        <Streak streak={Number(user.streak)} enabled={true} title="current streak" />
      {/if}
      {#if user.maxStreak > 0}
        <Streak streak={Number(user.maxStreak)} enabled={false} title="highest streak overall" />
      {/if}
    </div>

    <button
      onclick={() => (edit = true)}
      class="flex flex-row gap-2 justify-center items-center p-3 mt-auto rounded-xl transition-all duration-700 cursor-pointer hover:scale-105 glass"
    >
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
