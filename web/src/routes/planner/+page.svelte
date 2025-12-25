<script>
  import { create } from '@bufbuild/protobuf';
  import { PlanningClient } from '$lib/client/client.svelte';
  import {
    DeleteRequestSchema,
    GetRequestSchema,
    UpsertRequestSchema,
  } from '$lib/sdk/v1/scheduler/planning/planning_pb';
  import { Exec } from '$lib/error/error.svelte';
  import { EventSchema, EventType } from '$lib/sdk/v1/scheduler/event_pb';
  import { browser } from '$app/environment';
  import Event from './Event.svelte';
  import { goto } from '$app/navigation';
  import { Code, ConnectError } from '@connectrpc/connect';
  import EventTypeIcon from '$lib/components/EventTypeIcon.svelte';
  import { onMount } from 'svelte';

  const kitchenFormatter = new Intl.DateTimeFormat(undefined, {
    hour: 'numeric',
    minute: '2-digit',
    hour12: false,
  });

  const dayFormatter = new Intl.DateTimeFormat(undefined, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  });

  const durationFormatter = new Intl.DurationFormat(undefined, {
    style: 'narrow',
  });

  // factor applied to event seconds to get the pixels on the canvas.
  let shrinkFactor = $state(0.02);

  let loading = $state(false);

  let initialLoad = $state(false);

  /** @type {import("$lib/sdk/v1/scheduler/event_pb").Event[]}*/
  let events = $state([]);

  let day = $state(new Date());

  // morningOffset defines when the calendar day starts. Changes to this value only apply for empty calendars.
  let morningOffset = $state(0);

  let visualMorning = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), morningOffset),
  );

  let morning = $derived(new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 0));

  let evening = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 23, 59, 59),
  );

  // immutablePivot defines the pivot from where items are considered immutable.
  // items <= pivot cannot be changed by the planner anymore.
  let immutablePivot = $derived(
    events.findLastIndex(event => {
      return event.immutable;
    }),
  );

  // immutableTime defines the time before which no event should be allocated.
  // this is effectively the "start of the calendar" for any writes.
  let immutableTime = $derived.by(() => {
    if (immutablePivot >= 0 && events.length > immutablePivot)
      return events[immutablePivot].stopTime;
    else return visualMorning.getTime() / 1000;
  });

  async function loadEvents() {
    await Exec(
      async () => {
        const response = await PlanningClient().get(
          create(GetRequestSchema, {
            since: BigInt(morning.getTime() / 1000),
            until: BigInt(evening.getTime() / 1000),
          }),
        );
        events = response.events;
        if (!initialLoad) {
          // initially define when the calendar starts (if an event is already present, it starts there)
          if (events.length > 0) {
            morningOffset = new Date(Number(events[0].startTime) * 1000).getHours();
          } else if (browser) {
            morningOffset = Number(localStorage.getItem(`default_day_start`) || 6) || 6;
          }
        }
        initialLoad = true;
      },
      async () => {
        goto('/login');
      },
      processing => (loading = processing),
    );
  }

  $effect(() => {
    loadEvents();
  });

  let editMode = $state(false);

  let editableEvent = $state(
    create(EventSchema, {
      type: EventType.CHILL,
      name: '',
      musicUrl: '',
    }),
  );

  $effect(() => {
    if (browser) {
      editableEvent.musicUrl =
        localStorage.getItem(`default_music_${editableEvent.type.toString()}`) ?? '';
    }
  });

  $effect(() => {
    if (!editMode) {
      if (events.length > 0) editableEvent.startTime = events.at(-1)?.stopTime || BigInt(0);
      else editableEvent.startTime = BigInt(visualMorning.getTime() / 1000);
      editableEvent.stopTime = editableEvent.startTime + BigInt(3600);
    }
  });

  // updateEvents applies the user modified event list to the database.
  // the server automatically drops the old events that are still referenced by the event.id
  // and creates new events with an id of start_time.
  async function updateEvents() {
    snapAlignEvents();
    const updates = [];
    for (const [i, event] of events.entries()) {
      if (i <= immutablePivot) continue;
      if (event.id === event.startTime.toString()) continue; // optimize lookups by omitting unchanged events
      updates.push(event);
    }
    if (updates.length > 0) {
      await PlanningClient().upsert(create(UpsertRequestSchema, { events: updates }));
    }
  }

  // snapAlignEvents sorts events and ensures they align in one single block from morning - events[-1].
  // this ensures the expected zen "no-pause" behavior is enforced (no overlapping events and no empty spaces in the calendar)
  function snapAlignEvents() {
    events = events.sort((a, b) => Number(a.startTime - b.startTime));
    events.forEach((event, i, events) => {
      if (i <= immutablePivot) return;

      const duration = event.stopTime - event.startTime;
      if (i < 1) event.startTime = BigInt(immutableTime);
      else event.startTime = events[i - 1].stopTime;
      event.stopTime = event.startTime + duration;
    });
  }

  /** @type {HTMLDivElement|undefined} */
  let timeline = $state();

  let timelineHead = $state(Date.now());

  $effect(() => {
    timeline?.scrollIntoView({
      behavior: 'smooth',
      block: 'center',
      inline: 'center',
    });
  });

  onMount(() => {
    let animateFrame = 0;
    function updateTimeline() {
      timelineHead = Date.now();
      animateFrame = requestAnimationFrame(updateTimeline);
    }
    animateFrame = requestAnimationFrame(updateTimeline);
    return () => {
      cancelAnimationFrame(animateFrame);
    };
  });

  /** @type {HTMLDivElement|undefined} */
  let trashZone = $state();

  let maybeDrag = $state(0);
  let maybeDragY = $state(0);

  /** @type {import("$lib/sdk/v1/scheduler/event_pb").Event | undefined} */
  let dragged = $state(undefined);

  let dragWidth = $state(300);
  let dragX = $state(0);
  let dragY = $state(0);
  let initialDragY = $state(0);

  onMount(() => {
    // the following interceptors are required because the browser dragging api is completely retarded.
    // first of all they must be implemented in js because they must not be "passive"
    // (which makes the browser bypass the eventbubbling in the decision phase (to get smoother scrolling)).
    // besides we want to allow scroll OR dragndrop so we cannot preventDefault ontouchstart (as we maybe need default scroll behavior).
    // therefore we just hijack and cancel operations that could be used by the browser to cancel the event (scroll or contextmenu).
    /** @param {TouchEvent} e  */
    function interceptStart(e) {
      if (dragged) e.preventDefault();
    }
    /** @param {TouchEvent} e  */
    function interceptMove(e) {
      if (maybeDrag || dragged) e.preventDefault();
    }
    /** @param {PointerEvent} e  */
    function interceptContext(e) {
      if (maybeDrag || dragged) e.preventDefault();
    }
    window.addEventListener('touchstart', interceptStart, { passive: false });
    window.addEventListener('touchmove', interceptMove, { passive: false });
    window.addEventListener('contextmenu', interceptContext, { passive: false });
    return () => {
      window.removeEventListener('touchstart', interceptStart);
      window.removeEventListener('touchmove', interceptMove);
      window.removeEventListener('contextmenu', interceptContext);
    };
  });

  /** @param {PointerEvent} e @param {import("$lib/sdk/v1/scheduler/event_pb").Event} event  */
  function handleDown(e, event) {
    maybeDragY = e.y;
    maybeDrag = setTimeout(() => {
      maybeDrag = 0;
      if (e.target instanceof Element) {
        e.target?.setPointerCapture(e.pointerId);
      }
      dragged = event;
      dragX = e.x - dragWidth / 2;
      dragY = e.y - (Number(dragged.stopTime - dragged.startTime) * shrinkFactor) / 2;
      initialDragY = dragY;
    }, 100);
  }

  /** @param {PointerEvent} e */
  async function handleUp(e) {
    clearTimeout(maybeDrag);
    maybeDrag = 0;

    e.preventDefault();
    e.stopPropagation();
    const event = dragged;
    dragged = undefined;
    if (!event) return;

    if (e.target instanceof Element && e.target.hasPointerCapture(e.pointerId)) {
      try {
        e.target?.releasePointerCapture(e.pointerId);
      } catch {
        /* ignore weird legacy browser specific failures */
      }
    }

    const zone = document.elementFromPoint(e.x, e.y);
    if (zone === trashZone) {
      await Exec(
        async () => {
          await PlanningClient().delete(create(DeleteRequestSchema, { id: event?.id }));
          await loadEvents();
        },
        undefined,
        processing => (loading = processing),
      );
      await Exec(
        async () => {
          await updateEvents();
          await loadEvents();
        },
        undefined,
        processing => (loading = processing),
      );
    } else {
      await Exec(
        async () => {
          if (event.startTime <= immutableTime)
            throw new ConnectError('cannot move event to the past', Code.OutOfRange);
          await updateEvents();
        },
        undefined,
        processing => (loading = processing),
      );
      await Exec(
        async () => await loadEvents(),
        undefined,
        processing => (loading = processing),
      );
    }
  }

  /** @param {PointerEvent} e */
  function handleMove(e) {
    if (Math.abs(maybeDragY - e.y) > 8) {
      clearTimeout(maybeDrag);
      maybeDrag = 0;
    }

    if (dragged) {
      e.preventDefault();
      e.stopPropagation();
      dragX = e.x - dragWidth / 2;
      dragY = e.y - (Number(dragged.stopTime - dragged.startTime) * shrinkFactor) / 2;
      dragged.startTime += BigInt(Math.floor((dragY - initialDragY) / shrinkFactor));
      dragged.stopTime += BigInt(Math.floor((dragY - initialDragY) / shrinkFactor));
      initialDragY = dragY;
    }
  }
</script>

<svelte:head>
  <title>Planner | Zen</title>
  <link rel="canonical" href="https://zen.megakuul.com/planner" />
  <meta property="og:title" content="Zen Planner" />
  <meta property="og:type" content="website" />
  <meta property="og:image" content="https://zen.megakuul.com/favicon.svg" />
</svelte:head>

<svelte:window onpointerup={handleUp} onpointercancel={handleUp} onpointermove={handleMove} />

<div class="flex flex-col gap-2 items-center text-base sm:text-4xl max-h-[90dvh]">
  <div
    class="flex flex-row gap-3 justify-center items-center mb-1 font-bold sm:mb-5 text-slate-100/60"
  >
    <button
      class="p-1 w-20 text-center rounded-xl transition-all cursor-pointer hover:bg-slate-500/20"
      onclick={() => {
        const previous = new Date(day);
        previous.setDate(day.getDate() - 1);
        initialLoad = false;
        day = previous;
      }}>&lt;</button
    >
    <span>
      {dayFormatter.format(day)}
    </span>
    <button
      class="p-1 w-20 text-center rounded-xl transition-all cursor-pointer hover:bg-slate-500/20"
      onclick={() => {
        const next = new Date(day);
        next.setDate(day.getDate() + 1);
        initialLoad = false;
        day = next;
      }}>&gt;</button
    >
  </div>
  <div class="flex flex-row gap-4 items-center">
    <input
      type="text"
      autocorrect="off"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      bind:value={editableEvent.name}
      placeholder="Next Event"
      class="p-2 text-center rounded-xl sm:p-4 sm:max-w-full glass max-w-40 focus:outline-0"
      onkeydown={/** @type {KeyboardEvent} e */ e => {
        if (e.key === 'Enter') {
          e.preventDefault();
          e.currentTarget.blur();
        }
      }}
    />
    <button
      aria-label="type"
      class="p-3 text-center rounded-xl cursor-pointer sm:p-5 glass"
      onclick={() => {
        if (editableEvent.type + 1 > (Number(Object.values(EventType).at(-1)) ?? 0)) {
          editableEvent.type = EventType.CHILL;
        } else {
          editableEvent.type++;
        }
      }}
    >
      <EventTypeIcon type={editableEvent.type} startup={events.length < 1} />
    </button>
    <button
      onclick={async () =>
        Exec(
          async () => {
            await PlanningClient().upsert(
              create(UpsertRequestSchema, {
                events: [editableEvent],
              }),
            );
            editableEvent.name = '';
            await loadEvents();
          },
          undefined,
          processing => (loading = processing),
        )}
      class="p-3 text-center rounded-xl cursor-pointer sm:p-5 glass"
    >
      {#if loading}
        <!-- prettier-ignore -->
        <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
      {:else}
        <!-- prettier-ignore -->
        <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGbTzkqb2t)" d="M3 6h14v8.5a2.5 2.5 0 0 1-2.5 2.5h-9A2.5 2.5 0 0 1 3 14.5z"/><path fill="url(#SVGBQaaD6nJ)" d="M3 6h14v8.5a2.5 2.5 0 0 1-2.5 2.5h-9A2.5 2.5 0 0 1 3 14.5z"/><path fill="url(#SVG8pHbKctU)" fill-opacity="0.3" d="M3 6h14v8.5a2.5 2.5 0 0 1-2.5 2.5h-9A2.5 2.5 0 0 1 3 14.5z"/><path fill="url(#SVGflld8dUk)" d="M17 5.5A2.5 2.5 0 0 0 14.5 3h-9A2.5 2.5 0 0 0 3 5.5V7h14z"/><path fill="url(#SVGVoDsPd0j)" d="M19 14.5a4.5 4.5 0 1 0-9 0a4.5 4.5 0 0 0 9 0"/><path fill="url(#SVGpL9XBbGr)" fill-rule="evenodd" d="M16.854 12.646a.5.5 0 0 1 0 .708l-3 3a.5.5 0 0 1-.708 0l-1-1a.5.5 0 0 1 .708-.708l.646.647l2.646-2.647a.5.5 0 0 1 .708 0" clip-rule="evenodd"/><defs><linearGradient id="SVGbTzkqb2t" x1="8" x2="11.5" y1="6" y2="17" gradientUnits="userSpaceOnUse"><stop stop-color="#b3e0ff"/><stop offset="1" stop-color="#8cd0ff"/></linearGradient><linearGradient id="SVGBQaaD6nJ" x1="11.5" x2="13.5" y1="10.5" y2="19.5" gradientUnits="userSpaceOnUse"><stop stop-color="#dcf8ff" stop-opacity="0"/><stop offset="1" stop-color="#ff6ce8" stop-opacity="0.7"/></linearGradient><linearGradient id="SVGflld8dUk" x1="3.563" x2="4.904" y1="3" y2="9.816" gradientUnits="userSpaceOnUse"><stop stop-color="#0094f0"/><stop offset="1" stop-color="#2764e7"/></linearGradient><linearGradient id="SVGVoDsPd0j" x1="10.321" x2="16.532" y1="11.688" y2="18.141" gradientUnits="userSpaceOnUse"><stop stop-color="#52d17c"/><stop offset="1" stop-color="#22918b"/></linearGradient><linearGradient id="SVGpL9XBbGr" x1="12.938" x2="13.946" y1="12.908" y2="17.36" gradientUnits="userSpaceOnUse"><stop stop-color="#fff"/><stop offset="1" stop-color="#e3ffd9"/></linearGradient><radialGradient id="SVG8pHbKctU" cx="0" cy="0" r="1" gradientTransform="rotate(90 -.5 15)scale(6.5)" gradientUnits="userSpaceOnUse"><stop offset=".535" stop-color="#4a43cb"/><stop offset="1" stop-color="#4a43cb" stop-opacity="0"/></radialGradient></defs></g></svg>
      {/if}
    </button>
  </div>
  {#if editableEvent.name}
    {@const start = Number(editableEvent.startTime) * 1000}
    {@const stop = Number(editableEvent.stopTime) * 1000}
    {@const diff = new Date(stop - start)}
    <span class="mt-2 text-slate-400/80">
      {kitchenFormatter.format(start)} -
      {kitchenFormatter.format(stop)}
      (<span class="text-slate-200/80">
        {durationFormatter.format({
          hours: diff.getUTCHours(),
        })}
        {durationFormatter.format({
          minutes: diff.getUTCMinutes(),
        })}
      </span>)
    </span>
  {/if}
  <input
    type="range"
    name="duration"
    oninput={e => (editableEvent.stopTime = BigInt(e.currentTarget.value || 0))}
    step="300"
    min={Number(editableEvent.startTime)}
    max={evening.getTime() / 1000}
    class="w-full"
  />

  <div class="flex flex-row justify-between w-full h-[60dvh]">
    <div class="w-full h-full overflow-scroll-hidden">
      {#if !initialLoad}
        <div class="flex justify-center items-center w-full h-full">
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
        </div>
      {:else}
        <div
          style="height: {((+evening - +visualMorning + 1 * 60 * 60 * 1000) / 1000) *
            shrinkFactor}px"
          class="flex relative flex-col gap-1 items-center pr-4 pl-10 my-3 w-full"
        >
          {#each { length: evening.getHours() - visualMorning.getHours() }, i}
            {@const hour = (1 + i) * 60 * 60 * 1000}
            <div
              style="top: {(hour / 1000) * shrinkFactor}px"
              class="flex absolute right-0 flex-row gap-2 items-center w-full"
            >
              <span class="text-xs select-none text-slate-50/20">
                {kitchenFormatter.format(new Date(visualMorning.getTime() + hour))}
              </span>
              <hr class="w-full rounded-full border-none h-[2px] bg-slate-50/5" />
            </div>
          {/each}

          {#if day.getUTCDay() === new Date().getUTCDay()}
            <div
              bind:this={timeline}
              style="top: {((timelineHead - visualMorning.getTime()) / 1000) * shrinkFactor}px"
              class="flex absolute right-0 left-8 flex-row justify-center items-center z-[5]"
            >
              <hr class="w-full rounded-full border-none h-[3px] bg-red-800/40" />
              <div class="w-3 h-2 rounded-2xl bg-red-800/40"></div>
            </div>
          {/if}

          {#each events as event, i}
            {#if event.startTime >= BigInt(visualMorning.getTime() / 1000)}
              {@const immutable = i <= immutablePivot}
              <div
                style={event.id === dragged?.id
                  ? `position: fixed; width: ${dragWidth}px; top: ${dragY}px; left: ${dragX}px; z-index: 10`
                  : 'width: 100%;'}
                role="row"
                tabindex={0}
                onpointerdown={e => {
                  if (!immutable) {
                    handleDown(e, event);
                  }
                }}
              >
                <Event
                  {event}
                  {immutable}
                  startup={i < 1}
                  height={Number(event.stopTime - event.startTime) * shrinkFactor - 4}
                />
                {#if dragged && !immutable && dragged.startTime > event.startTime && (events.length - 1 === i || dragged.startTime < events[i + 1].startTime)}
                  <hr class="mt-1 w-full rounded-full border-2 border-slate-100/40" />
                {/if}
              </div>
            {/if}
          {/each}
          {#if editableEvent.name}
            <Event
              event={editableEvent}
              startup={events.length < 1}
              height={Number(editableEvent.stopTime - editableEvent.startTime) * shrinkFactor}
            />
          {/if}
        </div>
      {/if}
    </div>
    <input
      type="range"
      name="scale"
      bind:value={shrinkFactor}
      onchange={() => {
        timeline?.scrollIntoView({
          behavior: 'smooth',
          block: 'center',
          inline: 'center',
        });
      }}
      step="any"
      min={0.007}
      max={0.04}
      class="ml-3 [writing-mode:vertical-lr]"
    />
  </div>

  {#if dragged}
    <div
      bind:this={trashZone}
      class="flex fixed right-0 bottom-0 left-0 flex-row justify-center p-2 w-full text-center rounded-xl transition-all select-none sm:p-4 z-[100] bg-slate-500/80"
    >
      <!-- prettier-ignore -->
      <svg class="w-5 h-5 pointer-events-none sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGbTzkqb2t)" d="M17 6H3v8.5A2.5 2.5 0 0 0 5.5 17h9a2.5 2.5 0 0 0 2.5-2.5z"/><path fill="url(#SVGBQaaD6nJ)" d="M17 6H3v8.5A2.5 2.5 0 0 0 5.5 17h9a2.5 2.5 0 0 0 2.5-2.5z"/><path fill="url(#SVG8pHbKctU)" fill-opacity="0.3" d="M17 6H3v8.5A2.5 2.5 0 0 0 5.5 17h9a2.5 2.5 0 0 0 2.5-2.5z"/><path fill="url(#SVGflld8dUk)" d="M17 5.5A2.5 2.5 0 0 0 14.5 3h-9A2.5 2.5 0 0 0 3 5.5V7h14z"/><path fill="url(#SVGKG7lOdgM)" d="M19 14.5a4.5 4.5 0 1 1-9 0a4.5 4.5 0 0 1 9 0"/><path fill="url(#SVGCByFXbpl)" fill-rule="evenodd" d="M12.646 12.646a.5.5 0 0 1 .708 0l1.146 1.147l1.146-1.147a.5.5 0 0 1 .708.708L15.207 14.5l1.147 1.146a.5.5 0 0 1-.708.708L14.5 15.207l-1.146 1.147a.5.5 0 0 1-.708-.708l1.147-1.146l-1.147-1.146a.5.5 0 0 1 0-.708" clip-rule="evenodd"/><defs><linearGradient id="SVGbTzkqb2t" x1="8" x2="11.5" y1="6" y2="17" gradientUnits="userSpaceOnUse"><stop stop-color="#b3e0ff"/><stop offset="1" stop-color="#8cd0ff"/></linearGradient><linearGradient id="SVGBQaaD6nJ" x1="11.5" x2="13.5" y1="10.5" y2="19.5" gradientUnits="userSpaceOnUse"><stop stop-color="#dcf8ff" stop-opacity="0"/><stop offset="1" stop-color="#ff6ce8" stop-opacity="0.7"/></linearGradient><linearGradient id="SVGflld8dUk" x1="3.563" x2="4.904" y1="3" y2="9.816" gradientUnits="userSpaceOnUse"><stop stop-color="#0094f0"/><stop offset="1" stop-color="#2764e7"/></linearGradient><linearGradient id="SVGKG7lOdgM" x1="11.406" x2="17.313" y1="10.563" y2="19.281" gradientUnits="userSpaceOnUse"><stop stop-color="#f83f54"/><stop offset="1" stop-color="#ca2134"/></linearGradient><linearGradient id="SVGCByFXbpl" x1="12.977" x2="14.771" y1="14.652" y2="16.518" gradientUnits="userSpaceOnUse"><stop stop-color="#fdfdfd"/><stop offset="1" stop-color="#fecbe6"/></linearGradient><radialGradient id="SVG8pHbKctU" cx="0" cy="0" r="1" gradientTransform="rotate(90 -.5 15)scale(6.5)" gradientUnits="userSpaceOnUse"><stop offset=".535" stop-color="#4a43cb"/><stop offset="1" stop-color="#4a43cb" stop-opacity="0"/></radialGradient></defs></g></svg>
    </div>
  {/if}
</div>
