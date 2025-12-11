<script>
  import { onMount } from 'svelte';
  import { create } from '@bufbuild/protobuf';
  import { goto } from '$app/navigation';
  import { Exec } from '$lib/error/error.svelte';
  import { PlanningClient, TimingClient } from '$lib/client/client.svelte';
  import { GetRequestSchema } from '$lib/sdk/v1/scheduler/planning/planning_pb';
  import { StartRequestSchema, StopRequestSchema } from '$lib/sdk/v1/scheduler/timing/timing_pb';
  import EventTypeIcon from '$lib/components/EventTypeIcon.svelte';
  import { GetChangeDecorator } from '$lib/color/color';

  const kitchenFormatter = new Intl.DateTimeFormat(undefined, {
    hour: 'numeric',
    minute: '2-digit',
    hour12: false,
  });

  const counterFormatter = new Intl.DurationFormat('en', {
    style: 'digital',
  });

  const scoreFormatter = new Intl.NumberFormat(undefined, {
    signDisplay: 'always',
  });

  let initialLoad = $state(false);

  let day = new Date();

  let morning = $derived(new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate()));

  let evening = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 23, 59, 59),
  );

  /** @type {import('$lib/sdk/v1/scheduler/event_pb').Event[]}*/
  let events = $state([]);

  let activeEventIdx = $derived.by(() => {
    for (let i = 0; i < events.length; i++) {
      if (!events[i].timerStopTime) {
        return i;
      }
    }
    return -1;
  });
  let activeEvent = $derived(events[activeEventIdx]);
  let prevEvent = $derived(events[activeEventIdx - 1]);
  let nextEvent = $derived(events[activeEventIdx + 1]);

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
        initialLoad = true;
      },
      async () => {
        goto('/login');
      },
    );
  }

  let elapsed = $state();
  let animateFrame = 0;

  onMount(() => {
    loadEvents();

    function updateCounter() {
      if (activeEvent?.timerStartTime) {
        elapsed = Date.now() - Number(activeEvent.timerStartTime) * 1000;
      }

      animateFrame = requestAnimationFrame(updateCounter);
    }
    animateFrame = requestAnimationFrame(updateCounter);

    const interval = setInterval(async () => {
      await loadEvents();
    }, 10000);
    return () => {
      cancelAnimationFrame(animateFrame);
      clearInterval(interval);
    };
  });
</script>

<svelte:head>
  <title>Timer | Zen</title>
  <link rel="canonical" href="https://zen.megakuul.com/timer" />
  <meta property="og:title" content="Zen Timer" />
  <meta property="og:type" content="website" />
  <meta property="og:image" content="https://zen.megakuul.com/favicon.svg" />
</svelte:head>

<div
  class="flex flex-col gap-8 p-2 w-screen text-base rounded-2xl sm:p-8 sm:text-4xl h-[80dvh] max-w-[1000px]"
>
  {#if !initialLoad}
    <div class="flex justify-center items-center w-full h-full">
      <!-- prettier-ignore -->
      <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g stroke="currentColor" stroke-width="1"><circle cx="12" cy="12" r="9.5" fill="none" stroke-linecap="round" stroke-width="3"><animate attributeName="stroke-dasharray" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0 150;42 150;42 150;42 150"/><animate attributeName="stroke-dashoffset" calcMode="spline" dur="1.5s" keySplines="0.42,0,0.58,1;0.42,0,0.58,1;0.42,0,0.58,1" keyTimes="0;0.475;0.95;1" repeatCount="indefinite" values="0;-16;-59;-59"/></circle><animateTransform attributeName="transform" dur="2s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></g></svg>
    </div>
  {:else}
    <div
      class="flex flex-col gap-5 justify-center items-center p-8 w-full h-full rounded-2xl glass"
    >
      {#if prevEvent}
        <div
          class="flex flex-row gap-2 justify-start items-center p-4 w-full h-1/6 rounded-2xl opacity-10 glass"
        >
          <EventTypeIcon type={prevEvent.type} />
          <span class="overflow-hidden line-through text-nowrap">{prevEvent.name}</span>
          <span>
            (<span class={GetChangeDecorator(prevEvent.ratingChange)}>
              {scoreFormatter.format(prevEvent.ratingChange)}
            </span>)
          </span>
          <span class="flex flex-row gap-1 ml-auto">
            <span>{kitchenFormatter.format(new Date(Number(prevEvent.startTime) * 1000))}</span>
            <span>-</span>
            <span>{kitchenFormatter.format(new Date(Number(prevEvent.stopTime) * 1000))}</span>
          </span>
        </div>
      {/if}
      {#if activeEvent}
        <button
          onclick={async () => {
            await Exec(async () => {
              if (activeEvent.timerStartTime) {
                await TimingClient().stop(create(StopRequestSchema, { id: activeEvent.id }));
                if (nextEvent)
                  await TimingClient().start(create(StartRequestSchema, { id: nextEvent.id }));
              } else {
                await TimingClient().start(create(StartRequestSchema, { id: activeEvent.id }));
              }
              await loadEvents();
            }, undefined);
          }}
          class="flex flex-col justify-between items-center w-full h-full rounded-2xl cursor-pointer glass"
        >
          <div class="flex flex-row gap-2 items-center text-lg sm:text-xl text-slate-100/40">
            <EventTypeIcon type={activeEvent.type} svgClass="w-2 h-2 sm:w-4 sm:h-4" />
            <span>{activeEvent.name}</span>
          </div>
          {#if !activeEvent.timerStartTime}
            <p class="text-3xl font-bold sm:text-6xl text-slate-100/30">Start Event</p>
          {:else if elapsed}
            {@const elapsedDate = new Date(elapsed)}
            <p class="text-3xl sm:text-6xl">
              {counterFormatter.format({
                hours: elapsedDate.getUTCHours(),
                minutes: elapsedDate.getUTCMinutes(),
                seconds: elapsedDate.getUTCSeconds(),
              })}
            </p>
          {/if}
          <span class="flex flex-row gap-1 text-sm sm:text-lg text-slate-100/40">
            <span>{kitchenFormatter.format(new Date(Number(activeEvent.startTime) * 1000))}</span>
            <span>-</span>
            <span>{kitchenFormatter.format(new Date(Number(activeEvent.stopTime) * 1000))}</span>
          </span>
        </button>
      {/if}
      {#if nextEvent}
        <div
          class="flex flex-row gap-2 justify-start items-center p-4 w-full h-1/6 rounded-2xl opacity-30 glass"
        >
          <EventTypeIcon type={nextEvent.type} />
          <span class="overflow-hidden text-nowrap">{nextEvent.name}</span>
          <span class="flex flex-row gap-1 ml-auto">
            <span>{kitchenFormatter.format(new Date(Number(nextEvent.startTime) * 1000))}</span>
            <span>-</span>
            <span>{kitchenFormatter.format(new Date(Number(nextEvent.stopTime) * 1000))}</span>
          </span>
        </div>
      {/if}
      {#if !activeEvent && !prevEvent && !nextEvent}
        <h1 class="text-center">Nothing planned at the moment...</h1>
        <a href="/planner" class="p-4 w-full font-bold text-center rounded-xl max-w-[600px] glass">
          Plan now!
        </a>
        <p class="text-xs sm:text-base text-slate-400/40">
          Can't I get some time off? read <a href="/getting-started" class="underline">this</a>
        </p>
      {/if}
    </div>
  {/if}
</div>
