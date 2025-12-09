<script>
  import { onMount } from 'svelte';
  import { create } from '@bufbuild/protobuf';
  import { goto } from '$app/navigation';
  import { fade } from 'svelte/transition';
  import { Exec } from '$lib/error/error.svelte';
  import { PlanningClient, TimingClient } from '$lib/client/client.svelte';
  import { GetRequestSchema } from '$lib/sdk/v1/scheduler/planning/planning_pb';
  import { StartRequestSchema } from '$lib/sdk/v1/scheduler/timing/timing_pb';

  let loading = $state(false);

  let day = new Date();

  let morning = $derived(new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate()));

  let evening = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 23, 59, 59),
  );

  /** @type {import('$lib/sdk/v1/scheduler/event_pb').Event[]}*/
  let events = $state([]);

  let activeEventIdx = $derived.by(() => {
    for (let i = 0; i < events.length; i++) {
      if (events[i].timerStartTime && !events[i].timerStopTime) {
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
      },
      async () => {
        goto('/login');
      },
      processing => (loading = processing),
    );
  }

  let elapsed = $state();
  let animateFrame = 0;

  const duration = { hours: 1, minutes: 46, seconds: 40 };

  const counterFormat = new Intl.DurationFormat('en', { style: 'digital' });

  onMount(() => {
    loadEvents();

    function updateCounter() {
      if (activeEvent.timerStartTime) {
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

<div
  class="flex flex-col gap-8 p-2 w-screen text-base rounded-2xl sm:p-8 sm:text-4xl h-[80dvh] max-w-[1000px]"
>
  <div class="flex flex-col gap-5 justify-center items-center p-8 w-full h-3/6 rounded-2xl glass">
    {#if prevEvent}
      {prevEvent.name}
    {/if}
    {#if activeEvent}
      <button
        onclick={async () => {
          await Exec(
            async () => {
              await TimingClient().start(
                create(StartRequestSchema, {
                  id: activeEvent.id,
                }),
              );
              await loadEvents();
            },
            undefined,
            processing => (loading = processing),
          );
        }}
        class="w-full h-3/6 rounded-2xl cursor-pointer glass"
      >
        {activeEvent.name}
      </button>
      <div>{counterFormat.format({ seconds: elapsed / 1000 })}</div>
    {/if}
    {#if nextEvent}
      {nextEvent.name}
      <div class="w-full h-1/6 rounded-2xl glass"></div>
    {/if}
    {#if !activeEvent && !prevEvent && !nextEvent}
      <h1 class="text-center">Nothing planned at the moment...</h1>
      <a href="/planner" class="p-4 w-full font-bold text-center rounded-xl max-w-[600px] glass"
        >Plan now!</a
      >
      <p class="text-xs sm:text-base text-slate-400/40">
        Can't I get some time off? read <a href="/getting-started" class="underline">this</a>
      </p>
    {/if}
  </div>
</div>
