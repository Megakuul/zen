<script>
  import { create } from '@bufbuild/protobuf';
  import { PlanningClient } from '$lib/client/client.svelte';
  import {
    GetRequestSchema,
    UpsertRequestSchema,
  } from '$lib/sdk/v1/scheduler/planning/planning_pb';
  import { Exec } from '$lib/error/error.svelte';
  import { EventSchema, EventType } from '$lib/sdk/v1/scheduler/event_pb';
  import { browser } from '$app/environment';
  import Event from './Event.svelte';
  import { DeleteRequestSchema } from '$lib/sdk/v1/scheduler/planning/planning_pb';

  const kitchenFormatter = new Intl.DateTimeFormat(undefined, {
    hour: 'numeric',
    minute: '2-digit',
    hour12: false,
  });

  // factor applied to event seconds to get the pixels on the canvas.
  let shrinkFactor = $state(0.01);

  const morningThreshold = 6;

  /** @type {import("$lib/sdk/v1/scheduler/event_pb").Event[]}*/
  let events = $state([]);

  let loading = $state(false);

  let day = $state(new Date());

  let morning = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), morningThreshold),
  );

  let evening = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 23, 59, 59),
  );

  $effect(() => {
    Exec(
      async () => {
        const response = await PlanningClient().get(
          create(GetRequestSchema, {
            since: BigInt(morning.getTime() / 1000),
            until: BigInt(evening.getTime() / 1000),
          }),
        );
        events = response.events;
      },
      undefined,
      processing => (loading = processing),
    );
  });

  let newEventName = $state('');
  let newEventType = $state(EventType.CHILL);
  let newEventMusic = $derived.by(() => {
    if (browser) return localStorage.getItem(`default_music_${newEventType.toString()}`) ?? '';
    else '';
  });
  let newEventStart = $derived.by(() => {
    if (events.length > 0) return Number(events.at(-1)?.stopTime);
    else return morning.getTime() / 1000;
  });
  let newEventStop = $derived(newEventStart + 3600);

  let newEvent = $derived(
    create(EventSchema, {
      type: newEventType,
      name: newEventName,
      musicUrl: newEventMusic,
      startTime: BigInt(newEventStart),
      stopTime: BigInt(newEventStop),
    }),
  );

  // updateEvents applies the user modified event list to the database.
  // the server automatically drops the old events that are still referenced by the event.id
  // and creates new events with an id of start_time.
  function updateEvents() {
    // update start/stop times so that they align with the sorting order.
    events.forEach((event, i, events) => {
      const duration = event.stopTime - event.startTime;
      if (i < 1) event.startTime = BigInt(morning.getTime() / 1000);
      else event.startTime = events[i - 1].stopTime;
      event.stopTime = event.startTime + duration;
    });

    Exec(
      async () => {
        for (let event of events) {
          await PlanningClient().upsert(
            create(UpsertRequestSchema, {
              event: event,
            }),
          );
        }
        day = new Date(); // reset day to retrigger database load
        // this is required because when changing the events start_time
        // there is a mismatch between the id and the start_time
        // (used by the server to backtrack events even if their start_time changed).
      },
      undefined,
      processing => (loading = processing),
    );
  }
</script>

<div class="flex flex-col gap-2 items-center text-base sm:text-4xl">
  <div class="flex flex-row justify-between w-full">
    <div class="w-full h-[700px] overflow-scroll-hidden">
      <div
        style="height: {((+evening - +morning + 1 * 60 * 60 * 1000) / 1000) * shrinkFactor}px"
        class="flex relative flex-col gap-1 items-center py-3 pr-4 pl-10 w-full"
      >
        {#each { length: evening.getHours() - morning.getHours() }, i}
          {@const hour = (1 + i) * 60 * 60 * 1000}
          <div
            style="top: {(hour / 1000) * shrinkFactor}px"
            class="flex absolute right-0 flex-row gap-2 items-center w-full"
          >
            <span class="text-xs text-slate-50/20">
              {kitchenFormatter.format(new Date(morning.getTime() + hour))}
            </span>
            <hr class="w-full rounded-full border-none h-[2px] bg-slate-50/5" />
          </div>
        {/each}
        {#each events as event}
          <Event {event} height={Number(event.stopTime - event.startTime) * shrinkFactor} />
        {/each}
        {#if newEventName}
          <Event event={newEvent} height={Number(newEventStop - newEventStart) * shrinkFactor} />
        {/if}
      </div>
    </div>
    <input
      type="range"
      name="scale"
      bind:value={shrinkFactor}
      step="any"
      min={0.01}
      max={0.04}
      class="ml-3 [writing-mode:vertical-lr]"
    />
  </div>
  <div class="flex flex-row gap-4 items-center">
    <input
      bind:value={newEventName}
      placeholder="Next Event"
      class="p-3 text-center rounded-xl sm:p-5 glass focus:outline-0"
    />
    <button
      aria-label="type"
      class="p-4 text-center rounded-xl cursor-pointer sm:p-6 glass"
      onclick={() => {
        if (newEventType + 1 > EventType.MAX_FOCUS) {
          newEventType = EventType.CHILL;
        } else {
          newEventType++;
        }
      }}
    >
      {#if newEventType === EventType.CHILL}
        <p title="Relax" class="w-full h-full transition-all duration-700 hover:scale-125">
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><circle cx="9.5" cy="6.5" r="5.5" fill="url(#SVGml5grb2k)"/><circle cx="9.5" cy="6.5" r="5.5" fill="url(#SVGsvrtDgzX)" fill-opacity="0.8"/><path fill="url(#SVGf2AycIHP)" d="M6.239 2.07L9.73 5.563l.825-.824A3 3 0 0 1 10 3c0-.71.247-1.363.66-1.877a5.5 5.5 0 0 1 1.057.342a1.996 1.996 0 0 0-.44 2.551l1.743-1.743q.385.322.707.707l-1.743 1.743a1.996 1.996 0 0 0 2.55-.44q.224.505.344 1.057A3 3 0 0 1 13 6a3 3 0 0 1-1.738-.555l-.824.825l3.491 3.491q-.299.405-.665.75L9.73 6.976L6.854 9.854l-.708-.708L9.023 6.27L5.49 2.736a5.5 5.5 0 0 1 .749-.665" opacity="0.6"/><path fill="url(#SVGOccyWTeh)" d="M2.5 6A1.5 1.5 0 0 0 1 7.5v1A6.5 6.5 0 0 0 7.5 15h1a1.5 1.5 0 0 0 1.5-1.5v-1A6.5 6.5 0 0 0 3.5 6z"/><path fill="url(#SVGDVfHgdwi)" fill-opacity="0.9" d="M2.5 6A1.5 1.5 0 0 0 1 7.5v1A6.5 6.5 0 0 0 7.5 15h1a1.5 1.5 0 0 0 1.5-1.5v-1A6.5 6.5 0 0 0 3.5 6z"/><path fill="#ffc470" d="M5.104 9.396a.5.5 0 1 0-.708.708l1.5 1.5a.5.5 0 0 0 .708-.708z"/><defs><radialGradient id="SVGml5grb2k" cx="0" cy="0" r="1" gradientTransform="rotate(-90 14.355 3.374)scale(19.2202)" gradientUnits="userSpaceOnUse"><stop stop-color="#eb4824"/><stop offset=".978" stop-color="#ff921f"/></radialGradient><radialGradient id="SVGsvrtDgzX" cx="0" cy="0" r="1" gradientTransform="matrix(4.58333 -4.58333 6.27026 6.27026 5.375 10.625)" gradientUnits="userSpaceOnUse"><stop offset=".588" stop-color="#aa1d2d"/><stop offset=".931" stop-color="#eb4824" stop-opacity="0.1"/></radialGradient><radialGradient id="SVGOccyWTeh" cx="0" cy="0" r="1" gradientTransform="matrix(6.95452 9.40913 -8.50845 6.2888 1.818 6.409)" gradientUnits="userSpaceOnUse"><stop offset=".24" stop-color="#ae5606"/><stop offset="1" stop-color="#944600"/></radialGradient><radialGradient id="SVGDVfHgdwi" cx="0" cy="0" r="1" gradientTransform="rotate(10.938 -88.628 31.08)scale(13.3182)" gradientUnits="userSpaceOnUse"><stop offset=".626" stop-color="#ffa43d" stop-opacity="0"/><stop offset=".927" stop-color="#ffa43d"/></radialGradient><linearGradient id="SVGf2AycIHP" x1="13.66" x2="7.761" y1="2.51" y2="8.408" gradientUnits="userSpaceOnUse"><stop offset=".713" stop-color="#8e250b"/><stop offset=".903" stop-color="#8e250b" stop-opacity="0"/></linearGradient></defs></g></svg>
        </p>
      {:else if newEventType === EventType.PAUSE}
        <p title="Pause" class="transition-all duration-700 hover:scale-125">
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="none"><path fill="url(#SVGVT6WPNCe)" d="M13.5 1A4.5 4.5 0 0 0 9 5.5V7a1 1 0 0 0 1 1h.944l-.02.191c-.046.452-.109 1.062-.172 1.7c-.123 1.255-.252 2.663-.252 3.109a2 2 0 1 0 4 0c0-.446-.129-1.854-.252-3.11a304 304 0 0 0-.23-2.24L14 7.473V1.5a.5.5 0 0 0-.5-.5"/><path fill="url(#SVGNsmhjeCs)" d="M6.723 1.054a.5.5 0 0 1 .265.335C7.006 1.468 7.5 3.582 7.5 5c0 .95-.442 1.797-1.13 2.346c-.25.2-.37.418-.37.6v.486q0 .035.004.066c.034.248.157 1.169.272 2.124c.113.937.224 1.959.224 2.378a2 2 0 1 1-4 0c0-.42.111-1.44.224-2.378c.115-.955.238-1.876.272-2.124L3 8.432v-.486c0-.182-.12-.4-.37-.6A3 3 0 0 1 1.5 5c0-1.413.49-3.516.512-3.61A.505.505 0 0 1 2.505 1c.28 0 .507.227.507.507v2.998A.495.495 0 1 0 4 4.5v-3a.5.5 0 0 1 1 0v3.026a.495.495 0 0 0 .99-.021v-3c0-.279.226-.505.506-.505c.022 0 .12 0 .227.054"/><defs><linearGradient id="SVGVT6WPNCe" x1="8.154" x2="21.198" y1="1.875" y2="6.749" gradientUnits="userSpaceOnUse"><stop stop-color="#6ce0ff"/><stop offset="1" stop-color="#0067bf"/></linearGradient><linearGradient id="SVGNsmhjeCs" x1=".577" x2="14.483" y1="1.875" y2="7.543" gradientUnits="userSpaceOnUse"><stop stop-color="#6ce0ff"/><stop offset="1" stop-color="#0067bf"/></linearGradient></defs></g></svg>
        </p>
      {:else if newEventType === EventType.MIN_FOCUS}
        <p title="Normal Focus" class="transition-all duration-700 hover:scale-125">
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVG6sNa2bIm)" d="m3 13l-1-1v-2a8 8 0 1 1 16 0v2l-1 1l-1-1v-2a6 6 0 0 0-12 0v2z"/><path fill="url(#SVG2jUJYdbZ)" d="M4.5 11H2v5a2 2 0 0 0 2 2h1v-6.5a.5.5 0 0 0-.5-.5"/><path fill="url(#SVG2jUJYdbZ)" d="M17.5 11H15v7h1a2 2 0 0 0 2-2v-4.5a.5.5 0 0 0-.5-.5"/><path fill="url(#SVGfJHKreRU)" d="M7 11H4v7h3a1 1 0 0 0 1-1v-5a1 1 0 0 0-1-1"/><path fill="url(#SVGfJHKreRU)" d="M13 11h3v7h-3a1 1 0 0 1-1-1v-5a1 1 0 0 1 1-1"/><defs><linearGradient id="SVG6sNa2bIm" x1="-3.714" x2="-1.292" y1="2" y2="11.178" gradientUnits="userSpaceOnUse"><stop stop-color="#b9c0c7"/><stop offset="1" stop-color="#70777d"/></linearGradient><linearGradient id="SVG2jUJYdbZ" x1="16.5" x2="16.5" y1="11" y2="18" gradientUnits="userSpaceOnUse"><stop stop-color="#0078d4"/><stop offset="1" stop-color="#2052cb"/></linearGradient><linearGradient id="SVGfJHKreRU" x1="14.25" x2="14.25" y1="11" y2="18" gradientUnits="userSpaceOnUse"><stop stop-color="#0fafff"/><stop offset="1" stop-color="#0067bf"/></linearGradient></defs></g></svg>
        </p>
      {:else if newEventType === EventType.MAX_FOCUS}
        <p title="High Focus (morph)" class="transition-all duration-700 hover:scale-125">
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGka68peCK)" d="M18.176 7.53a3.6 3.6 0 0 0-1.165-.206c-.032-.083-2.085-.123-2.434-.185c-2.662.995-8.521 5.08-11.576 5.08v.465a7.5 7.5 0 0 0 7.001 4.816a7.5 7.5 0 0 0 7.28-5.69c1.281-.084 2.343-.754 2.632-1.677a1.83 1.83 0 0 0-.184-1.508c-.283-.471-.893-.85-1.554-1.094"/><path fill="url(#SVGca7svbTj)" d="M18.176 7.53a3.6 3.6 0 0 0-1.165-.206c-.032-.083-2.085-.123-2.434-.185c-2.662.995-8.521 5.08-11.576 5.08v.465a7.5 7.5 0 0 0 7.001 4.816a7.5 7.5 0 0 0 7.28-5.69c1.281-.084 2.343-.754 2.632-1.677a1.83 1.83 0 0 0-.184-1.508c-.283-.471-.893-.85-1.554-1.094"/><path fill="url(#SVGb1CAndaw)" d="M18.176 7.53a3.6 3.6 0 0 0-1.165-.206c-.032-.083-2.085-.123-2.434-.185c-2.662.995-8.521 5.08-11.576 5.08v.465a7.5 7.5 0 0 0 7.001 4.816a7.5 7.5 0 0 0 7.28-5.69c1.281-.084 2.343-.754 2.632-1.677a1.83 1.83 0 0 0-.184-1.508c-.283-.471-.893-.85-1.554-1.094"/><path fill="url(#SVGmnBdqeXI)" d="M18.176 7.53a3.6 3.6 0 0 0-1.165-.206c-.032-.083-2.085-.123-2.434-.185c-2.662.995-8.521 5.08-11.576 5.08v.465a7.5 7.5 0 0 0 7.001 4.816a7.5 7.5 0 0 0 7.28-5.69c1.281-.084 2.343-.754 2.632-1.677a1.83 1.83 0 0 0-.184-1.508c-.283-.471-.893-.85-1.554-1.094"/><path fill="url(#SVG8APyAbci)" fill-opacity="0.5" d="M18.176 7.53a3.6 3.6 0 0 0-1.165-.206c-.032-.083-2.085-.123-2.434-.185c-2.662.995-8.521 5.08-11.576 5.08v.465a7.5 7.5 0 0 0 7.001 4.816a7.5 7.5 0 0 0 7.28-5.69c1.281-.084 2.343-.754 2.632-1.677a1.83 1.83 0 0 0-.184-1.508c-.283-.471-.893-.85-1.554-1.094"/><path fill="url(#SVGVJ1YudkF)" d="M18.176 7.53a3.6 3.6 0 0 0-1.165-.206c-.032-.083-2.085-.123-2.434-.185c-2.662.995-8.521 5.08-11.576 5.08v.465a7.5 7.5 0 0 0 7.001 4.816a7.5 7.5 0 0 0 7.28-5.69c1.281-.084 2.343-.754 2.632-1.677a1.83 1.83 0 0 0-.184-1.508c-.283-.471-.893-.85-1.554-1.094"/><path fill="url(#SVGjaJE5bGd)" d="M1.827 12.47c.351.13.74.195 1.164.206l.009.022v-.022c2.563 0 4.78-1.338 7-2.677c2.218-1.339 4.438-2.678 7.006-2.683A7.503 7.503 0 0 0 2.72 8.19C1.44 8.274.375 8.95.086 9.871A1.83 1.83 0 0 0 .27 11.38c.283.471.896.846 1.557 1.09"/><path fill="url(#SVGxOnP8RRz)" d="M1.827 12.47c.351.13.74.195 1.164.206l.009.022v-.022c2.563 0 4.78-1.338 7-2.677c2.218-1.339 4.438-2.678 7.006-2.683A7.503 7.503 0 0 0 2.72 8.19C1.44 8.274.375 8.95.086 9.871A1.83 1.83 0 0 0 .27 11.38c.283.471.896.846 1.557 1.09"/><path fill="url(#SVG3mpDHeMY)" d="M1.827 12.47c.351.13.74.195 1.164.206l.009.022v-.022c2.563 0 4.78-1.338 7-2.677c2.218-1.339 4.438-2.678 7.006-2.683A7.503 7.503 0 0 0 2.72 8.19C1.44 8.274.375 8.95.086 9.871A1.83 1.83 0 0 0 .27 11.38c.283.471.896.846 1.557 1.09"/><path fill="url(#SVGA5uZmeGT)" d="M1.827 12.47c.351.13.74.195 1.164.206l.009.022v-.022c2.563 0 4.78-1.338 7-2.677c2.218-1.339 4.438-2.678 7.006-2.683A7.503 7.503 0 0 0 2.72 8.19C1.44 8.274.375 8.95.086 9.871A1.83 1.83 0 0 0 .27 11.38c.283.471.896.846 1.557 1.09"/><defs><radialGradient id="SVGka68peCK" cx="0" cy="0" r="1" gradientTransform="matrix(-9.35759 -12.90423 12.79122 -9.27564 16.465 18.553)" gradientUnits="userSpaceOnUse"><stop stop-color="#8f77ff"/><stop offset=".457" stop-color="#775be3"/><stop offset=".656" stop-color="#6552d9"/></radialGradient><radialGradient id="SVGca7svbTj" cx="0" cy="0" r="1" gradientTransform="rotate(160.502 8.735 7.527)scale(3.47169 8.68803)" gradientUnits="userSpaceOnUse"><stop offset=".241" stop-color="#6e30c8"/><stop offset="1" stop-color="#6730c6" stop-opacity="0"/></radialGradient><radialGradient id="SVGb1CAndaw" cx="0" cy="0" r="1" gradientTransform="rotate(-124.74 12.726 7.385)scale(7.17869 27.5349)" gradientUnits="userSpaceOnUse"><stop stop-color="#f36284"/><stop offset="1" stop-color="#f36284" stop-opacity="0"/></radialGradient><radialGradient id="SVGmnBdqeXI" cx="0" cy="0" r="1" gradientTransform="matrix(-13.5506 0 0 -7.64451 20.71 7.14)" gradientUnits="userSpaceOnUse"><stop stop-color="#7d40c8"/><stop offset="1" stop-color="#7f45d2" stop-opacity="0"/></radialGradient><radialGradient id="SVG8APyAbci" cx="0" cy="0" r="1" gradientTransform="rotate(55.581 -10.278 19.358)scale(15.0359 45.6198)" gradientUnits="userSpaceOnUse"><stop offset=".265" stop-color="#0a26b5" stop-opacity="0"/><stop offset=".581" stop-color="#051d92"/></radialGradient><radialGradient id="SVGVJ1YudkF" cx="0" cy="0" r="1" gradientTransform="rotate(175.731 9.933 6.89)scale(3.53809 9.54879)" gradientUnits="userSpaceOnUse"><stop offset=".195" stop-color="#e173e7"/><stop offset=".901" stop-color="#e173e7" stop-opacity="0"/></radialGradient><radialGradient id="SVGjaJE5bGd" cx="0" cy="0" r="1" gradientTransform="matrix(12.59449 9.24328 -9.2108 12.55023 4.919 3.12)" gradientUnits="userSpaceOnUse"><stop stop-color="#3dd3dc"/><stop offset="1" stop-color="#4290f0"/></radialGradient><radialGradient id="SVGxOnP8RRz" cx="0" cy="0" r="1" gradientTransform="matrix(3.29865 -4.23435 10.32087 8.04019 0 8.18)" gradientUnits="userSpaceOnUse"><stop offset=".285" stop-color="#3a80e1"/><stop offset="1" stop-color="#488ae5" stop-opacity="0"/></radialGradient><radialGradient id="SVG3mpDHeMY" cx="0" cy="0" r="1" gradientTransform="matrix(2.61929 -.46475 1.42188 8.01365 0 6.58)" gradientUnits="userSpaceOnUse"><stop offset=".232" stop-color="#3dd3dc"/><stop offset="1" stop-color="#3dd3dc" stop-opacity="0"/></radialGradient><radialGradient id="SVGA5uZmeGT" cx="0" cy="0" r="1" gradientTransform="rotate(67.947 .107 4.89)scale(10.3999 31.8176)" gradientUnits="userSpaceOnUse"><stop offset=".576" stop-color="#26cfdb" stop-opacity="0"/><stop offset="1" stop-color="#19d9e7"/></radialGradient></defs></g></svg>
        </p>
      {/if}
    </button>
  </div>
  <input
    type="range"
    name="duration"
    bind:value={newEventStop}
    step="300"
    min={newEventStart}
    max={evening.getTime() / 1000}
    class="my-3 w-full"
  />
  <button
    onclick={async () =>
      Exec(
        async () => {
          await PlanningClient().upsert(
            create(UpsertRequestSchema, {
              event: newEvent,
            }),
          );
        },
        undefined,
        processing => (loading = processing),
      )}
    class="flex flex-row gap-2 justify-center items-center p-3 w-full rounded-xl transition-all duration-700 cursor-pointer hover:scale-105 glass"
  >
    {#if loading}
      <!-- prettier-ignore -->
      <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
    {:else}
      <span>Add Event</span>
      <!-- prettier-ignore -->
      <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g fill="none"><path fill="url(#SVG3sxK6chn)" d="M12.793 1.383a1 1 0 0 0-1.579 0L9.801 3.2a.25.25 0 0 1-.291.079L7.378 2.41a1 1 0 0 0-1.367.79l-.315 2.28a.25.25 0 0 1-.213.213l-2.28.315a1 1 0 0 0-.79 1.367l.868 2.132a.25.25 0 0 1-.079.291l-1.816 1.413a1 1 0 0 0 0 1.579l1.816 1.413a.25.25 0 0 1 .079.291l-.867 2.132a1 1 0 0 0 .79 1.367l2.279.315a.25.25 0 0 1 .213.213l.315 2.28a1 1 0 0 0 1.367.79l2.132-.868a.25.25 0 0 1 .291.079l1.413 1.816a1 1 0 0 0 1.579 0l1.413-1.816a.25.25 0 0 1 .291-.079l2.131.867a1 1 0 0 0 1.368-.79l.315-2.279a.25.25 0 0 1 .213-.213l2.28-.315a1 1 0 0 0 .789-1.367l-.867-2.132a.25.25 0 0 1 .079-.291l1.816-1.413a1 1 0 0 0 0-1.579l-1.816-1.413a.25.25 0 0 1-.079-.291l.867-2.132a1 1 0 0 0-.79-1.367l-2.279-.315a.25.25 0 0 1-.213-.213l-.315-2.28a1 1 0 0 0-1.367-.79l-2.132.868a.25.25 0 0 1-.291-.079z"/><path fill="url(#SVG2VNJKcHX)" fill-opacity="0.95" d="M12 7a.75.75 0 0 1 .75.75v3.5h3.5a.75.75 0 0 1 0 1.5h-3.5v3.5a.75.75 0 0 1-1.5 0v-3.5h-3.5a.75.75 0 0 1 0-1.5h3.5v-3.5A.75.75 0 0 1 12 7"/><defs><radialGradient id="SVG3sxK6chn" cx="0" cy="0" r="1" gradientTransform="matrix(-23.9474 -42.34411 40.5584 -22.9375 26.245 26.212)" gradientUnits="userSpaceOnUse"><stop stop-color="#ffc470"/><stop offset=".251" stop-color="#ff835c"/><stop offset=".55" stop-color="#f24a9d"/><stop offset=".814" stop-color="#b339f0"/></radialGradient><linearGradient id="SVG2VNJKcHX" x1="16.305" x2="5.813" y1="19.823" y2="13.027" gradientUnits="userSpaceOnUse"><stop offset=".024" stop-color="#ffc8d7"/><stop offset=".807" stop-color="#fff"/></linearGradient></defs></g></svg>
    {/if}
  </button>
</div>
