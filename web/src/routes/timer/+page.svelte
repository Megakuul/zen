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

  // state used to retrigger activeEvent and nextEvent state calculations.
  let now = $state(BigInt(Date.now()) / 1000n);

  /** @type {import('$lib/sdk/v1/scheduler/event_pb').Event[]|undefined}*/
  let events = $state();

  let activeEvent = $derived.by(() => {
    if (!events) return undefined;
    for (const event of events) {
      if (event.startTime < now && now > event.stopTime) {
        return event;
      }
    }
    return undefined;
  });
  let nextEvent = $derived.by(() => {
    if (!events) return undefined;
    let time = now;
    if (activeEvent) {
      time = activeEvent.stopTime;
    }
    for (const event of events) {
      if (event.startTime > time) {
        return event;
      }
    }
    return undefined;
  });

  let active = $derived(
    (activeEvent?.timerStartTime ?? 0 < 1) && !(activeEvent?.timerStopTime ?? 0 < 1),
  );

  async function loadEvents() {
    await Exec(
      async () => {
        const response = await PlanningClient().get(create(GetRequestSchema, {}));
        events = response.events;
      },
      async () => {
        goto('/login');
      },
      loading,
    );
  }

  onMount(() => {
    loadEvents();

    const interval = setInterval(() => {
      now = BigInt(Date.now()) / 1000n;
    }, 1000);
    return () => clearInterval(interval);
  });
</script>

<div class="flex flex-col gap-8 p-2 w-screen text-base rounded-2xl sm:p-8 sm:text-4xl h-dvh">
  {#if activeEvent}
    <button
      onclick={async () => {
        await Exec(
          async () => {
            await TimingClient().start(
              create(StartRequestSchema, {
                startTime: BigInt(Date.now()) / 1000n,
              }),
            );
            await loadEvents();
          },
          undefined,
          loading,
        );
      }}
      class="w-full h-3/6 rounded-2xl cursor-pointer glass"
    >
    </button>
  {:else}
    <div class="flex flex-col gap-3 justify-center items-center p-8 w-full h-3/6 rounded-2xl glass">
      <h1 class="text-center">Nothing planned at the moment...</h1>
      <a href="/planner" class="p-4 w-full font-bold text-center rounded-xl glass">Plan now!</a>
      <p class="text-xs sm:text-base text-slate-400/40">
        Can't I get some time off? read <a href="/getting-started" class="underline">this</a>
      </p>
    </div>
  {/if}
  {#if nextEvent}
    <div class="w-full h-1/6 rounded-2xl glass"></div>
  {/if}
</div>

