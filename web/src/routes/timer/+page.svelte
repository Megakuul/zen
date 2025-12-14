<script>
  import { onMount } from 'svelte';
  import { create } from '@bufbuild/protobuf';
  import { goto } from '$app/navigation';
  import { Exec } from '$lib/error/error.svelte';
  import { ManagementClient, PlanningClient, TimingClient } from '$lib/client/client.svelte';
  import { GetRequestSchema as PlanningGetSchema } from '$lib/sdk/v1/scheduler/planning/planning_pb';
  import { GetRequestSchema as ManagementGetSchema } from '$lib/sdk/v1/manager/management/management_pb';
  import { StartRequestSchema, StopRequestSchema } from '$lib/sdk/v1/scheduler/timing/timing_pb';
  import EventTypeIcon from '$lib/components/EventTypeIcon.svelte';
  import { GetChangeTextDecorator } from '$lib/color/color';
  import Streak from '$lib/components/Streak.svelte';
  import Fireworks from '@fireworks-js/svelte';
  import Countup from '$lib/components/Countup.svelte';

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

  let ratingChange = $state(0);

  /** @type {Fireworks} */
  let fw;
  $effect(() => {
    const fireworks = fw.fireworksInstance();
    if (ratingChange > 0) {
      fireworks.start();
      setTimeout(() => {
        fireworks.waitStop();
      }, 1000);
    } else fireworks.waitStop();
  });

  let day = new Date();

  let morning = $derived(new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate()));

  let evening = $derived(
    new Date(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 23, 59, 59),
  );

  /** @type {import('$lib/sdk/v1/manager/user_pb').User|undefined}*/
  let user = $state();

  /** @type {import('$lib/sdk/v1/scheduler/event_pb').Event[]}*/
  let events = $state([]);

  let activeEventIdx = $derived.by(() => {
    for (let i = 0; i < events.length; i++) {
      if (i !== 0 && !events[i].timerStopTime) {
        return i;
      }
    }
    return NaN;
  });
  let activeEvent = $derived(events[activeEventIdx]);
  let prevEvent = $derived(events[activeEventIdx - 1]);
  let nextEvent = $derived(events[activeEventIdx + 1]);

  async function loadEvents() {
    await Exec(
      async () => {
        const eventResponse = await PlanningClient().get(
          create(PlanningGetSchema, {
            since: BigInt(morning.getTime() / 1000),
            until: BigInt(evening.getTime() / 1000),
          }),
        );
        events = eventResponse.events;
        const userResponse = await ManagementClient().get(create(ManagementGetSchema, {}));
        user = userResponse.user;
        ratingChange = 0;
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
      if (!ratingChange) await loadEvents();
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

<Fireworks
  bind:this={fw}
  autostart={false}
  class="fixed inset-0 w-full h-full"
  options={{
    opacity: 0.1,
    hue: {
      max: 186,
      min: 128,
    },
    decay: {
      max: 0.03,
      min: 0.001,
    },
    intensity: 30,
    traceSpeed: 1,
    particles: 80,
    friction: 0.99,
    acceleration: 1.1,
  }}
/>

<div
  class="flex flex-col gap-8 p-2 w-screen text-base rounded-2xl sm:p-8 sm:text-4xl h-[85dvh] max-w-[1000px]"
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
          class="flex flex-row gap-2 justify-start items-center p-4 w-full h-1/6 rounded-2xl opacity-40 glass"
        >
          <EventTypeIcon type={prevEvent.type} />
          <span class="overflow-hidden line-through text-nowrap">{prevEvent.name}</span>
          <span class="brightness-200">
            (<span class={GetChangeTextDecorator(prevEvent.ratingChange)}>
              {scoreFormatter.format(prevEvent.ratingChange)}
            </span>)
          </span>
          <span class="flex flex-row gap-1 ml-auto opacity-50">
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
                const response = await TimingClient().stop(
                  create(StopRequestSchema, { id: activeEvent.id }),
                );
                ratingChange = response.ratingChange;
                setTimeout(async () => await loadEvents(), 3000);
                if (nextEvent)
                  await TimingClient().start(create(StartRequestSchema, { id: nextEvent.id }));
              } else {
                await TimingClient().start(create(StartRequestSchema, { id: activeEvent.id }));
                await loadEvents();
              }
            }, undefined);
          }}
          class="flex overflow-hidden flex-col justify-between items-center w-full h-full rounded-2xl cursor-pointer glass"
        >
          {#if ratingChange}
            <div class="flex flex-row gap-2 items-center text-lg sm:text-xl text-slate-100/40">
              <EventTypeIcon type={activeEvent.type} svgClass="w-2 h-2 sm:w-4 sm:h-4" />
              <span>{activeEvent.name}</span>
            </div>
            {#if !activeEvent.timerStartTime}
              <p class="text-3xl font-bold sm:text-6xl text-slate-100/30">Start Event</p>
            {:else if elapsed}
              {@const elapsedDate = new Date(elapsed)}
              {@const expectedDate = new Date(
                Number(activeEvent.stopTime - activeEvent.startTime) * 1000,
              )}
              <div class="flex relative flex-col justify-center items-center w-full h-full">
                <svg
                  class="stroke-slate-100/20 [stroke-linecap:round] w-[240px] h-[240px] sm:w-[400px] sm:h-[400px]"
                >
                  <circle
                    class="[cx:120px] [cy:120px] [r:100px] sm:[cx:200px] sm:[cy:200px] sm:[r:180px]"
                    stroke-width="10"
                    fill="none"
                    pathLength={expectedDate.getTime()}
                    stroke-dasharray={expectedDate.getTime()}
                  />
                </svg>
                <svg
                  class="absolute top-1/2 left-1/2 translate-x-[-50%] translate-y-[-50%] stroke-slate-100/60 [stroke-linecap:round] w-[240px] h-[240px] sm:w-[400px] sm:h-[400px]"
                >
                  <circle
                    class="[cx:120px] [cy:120px] [r:100px] sm:[cx:200px] sm:[cy:200px] sm:[r:180px]"
                    stroke-width="10"
                    fill="none"
                    pathLength={expectedDate.getTime()}
                    stroke-dashoffset={+expectedDate / 4}
                    stroke-dasharray="{+elapsedDate} {+expectedDate - +elapsedDate}"
                  />
                </svg>
                <div
                  class="flex flex-col justify-center items-center absolute top-1/2 left-1/2 translate-x-[-50%] translate-y-[-50%]"
                >
                  {#if user?.streak}
                    <Streak streak={Number(user?.streak)} enabled={true} title="current streak" />
                  {/if}
                  <p class="text-3xl sm:text-6xl">
                    {counterFormatter.format({
                      hours: elapsedDate.getUTCHours(),
                      minutes: elapsedDate.getUTCMinutes(),
                      seconds: elapsedDate.getUTCSeconds(),
                    })}
                  </p>
                  <p class="sm:text-2xl text-1xl text-slate-200/40">
                    {counterFormatter.format({
                      hours: expectedDate.getUTCHours(),
                      minutes: expectedDate.getUTCMinutes(),
                      seconds: expectedDate.getUTCSeconds(),
                    })}
                  </p>
                </div>
              </div>
            {/if}
            <span class="flex flex-row gap-1 text-sm sm:text-lg text-slate-100/40">
              <span>{kitchenFormatter.format(new Date(Number(activeEvent.startTime) * 1000))}</span>
              <span>-</span>
              <span>{kitchenFormatter.format(new Date(Number(activeEvent.stopTime) * 1000))}</span>
            </span>
          {:else}
            <div
              class="flex justify-center items-center w-full h-full
                {ratingChange < 0 ? 'missed-magic' : 'success-magic'}"
            >
              <p
                class="text-4xl sm:text-8xl px-8 py-2 rounded-2xl shadow-inner min-w-64 shadow-slate-800/20 bg-stone-950/80 brightness-200 {GetChangeTextDecorator(
                  ratingChange,
                )}"
              >
                <Countup value={scoreFormatter.format(ratingChange)} />
              </p>
            </div>
          {/if}
        </button>
      {/if}
      {#if nextEvent}
        <div
          class="flex flex-row gap-2 justify-start items-center p-4 w-full h-1/6 rounded-2xl opacity-50 glass"
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
    {#if activeEvent && activeEvent.musicUrl}
      {@const hostname = URL.parse(activeEvent.musicUrl)?.hostname}
      <a
        class="flex flex-row gap-2 justify-center items-center p-2 w-full rounded-xl glass"
        target="_blank"
        rel="noopener noreferrer"
        href={activeEvent.musicUrl}
      >
        <span>Music</span>

        {#if hostname?.endsWith('spotify.com')}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g fill="none"><path fill="url(#SVG9DPAsdPE)" d="M9 7a1 1 0 0 1 .117 1.993L9 9H7a3 3 0 0 0-.176 5.995L7 15h2a1 1 0 0 1 .117 1.993L9 17H7a5 5 0 0 1-.217-9.995L7 7zm8 0a5 5 0 0 1 .217 9.995L17 17h-2a1 1 0 0 1-.117-1.993L15 15h2a3 3 0 0 0 .176-5.995L17 9h-2a1 1 0 0 1-.117-1.993L15 7zM7 11h10a1 1 0 0 1 .117 1.993L17 13H7a1 1 0 0 1-.117-1.993zh10z"/><defs><linearGradient id="SVG9DPAsdPE" x1="-4.429" x2="3.504" y1="2.625" y2="26.481" gradientUnits="userSpaceOnUse"><stop stop-color="#36dff1"/><stop offset="1" stop-color="#2764e7"/></linearGradient></defs></g></svg>
        {:else if hostname?.endsWith('youtube.com') || hostname?.endsWith('youtu.be')}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 180"><path fill="#f00" d="M250.346 28.075A32.18 32.18 0 0 0 227.69 5.418C207.824 0 127.87 0 127.87 0S47.912.164 28.046 5.582A32.18 32.18 0 0 0 5.39 28.24c-6.009 35.298-8.34 89.084.165 122.97a32.18 32.18 0 0 0 22.656 22.657c19.866 5.418 99.822 5.418 99.822 5.418s79.955 0 99.82-5.418a32.18 32.18 0 0 0 22.657-22.657c6.338-35.348 8.291-89.1-.164-123.134"/><path fill="#fff" d="m102.421 128.06l66.328-38.418l-66.328-38.418z"/></svg>
        {:else}
          <!-- prettier-ignore -->
          <svg class="w-5 h-5 sm:w-8 sm:h-8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><g fill="none"><path fill="url(#SVGuv0AxdpE)" d="M10 18a8 8 0 1 0 0-16a8 8 0 0 0 0 16"/><path fill="url(#SVGhWhXiclN)" fill-rule="evenodd" d="M7.853 2.291a7 7 0 0 0-.816 1.51c-.368.906-.65 1.995-.826 3.199H2.58q-.195.485-.33 1h3.84a22 22 0 0 0 .001 4h-3.84q.135.515.33 1h3.63c.176 1.204.458 2.293.826 3.199a7 7 0 0 0 .816 1.51A8 8 0 0 0 10 18a8 8 0 0 0 2.147-.291a7 7 0 0 0 .816-1.51c.368-.906.65-1.995.826-3.199h3.63q.195-.485.329-1h-3.84a21.6 21.6 0 0 0 0-4h3.84a8 8 0 0 0-.33-1H13.79c-.176-1.204-.458-2.293-.826-3.199a7 7 0 0 0-.816-1.51A8 8 0 0 0 10 2a8 8 0 0 0-2.147.291M7.223 7c.166-1.076.42-2.035.74-2.822c.298-.733.642-1.292 1.003-1.66C9.324 2.153 9.672 2 10 2s.676.153 1.034.518c.36.368.705.927 1.003 1.66c.32.787.574 1.746.74 2.822zM10 18c.328 0 .676-.153 1.034-.518c.36-.368.705-.927 1.003-1.66c.32-.787.574-1.746.74-2.822H7.223c.167 1.076.421 2.035.741 2.822c.298.733.642 1.292 1.003 1.66c.358.365.706.518 1.034.518m-3-8c0 .692.033 1.362.096 2h5.808A21 21 0 0 0 13 10c0-.692-.033-1.362-.096-2H7.096A21 21 0 0 0 7 10" clip-rule="evenodd"/><defs><radialGradient id="SVGhWhXiclN" cx="0" cy="0" r="1" gradientTransform="rotate(225 10.4 3.895)scale(12.7313)" gradientUnits="userSpaceOnUse"><stop stop-color="#25a2f0"/><stop offset=".974" stop-color="#3bd5ff"/></radialGradient><linearGradient id="SVGuv0AxdpE" x1="5.556" x2="17.111" y1="4.667" y2="15.333" gradientUnits="userSpaceOnUse"><stop stop-color="#29c3ff"/><stop offset="1" stop-color="#2052cb"/></linearGradient></defs></g></svg>
        {/if}
      </a>
    {/if}
  {/if}
</div>

<style>
  .success-magic {
    background-image: linear-gradient(
      to bottom right,
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 0) 50%,
      rgba(0, 255, 0, 0.1) 75%
    );
    background-size: 400% 400%;
    animation: move-magic 2s ease forwards;
  }

  .missed-magic {
    background-image: linear-gradient(
      to bottom right,
      rgba(0, 0, 0, 0) 0%,
      rgba(0, 0, 0, 0) 50%,
      rgba(255, 0, 0, 0.1) 75%
    );
    background-size: 400% 400%;
    background-position: 100% 100%;
  }

  @keyframes move-magic {
    0% {
      background-position: 0% 0%;
    }
    100% {
      background-position: 100% 100%;
    }
  }

  .success-magic * {
    animation: popup-magic 2s ease forwards;
  }
</style>
