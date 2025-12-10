<script>
  import { GetChangeDecorator } from '$lib/color/color';
  import EventTypeIcon from '$lib/components/EventTypeIcon.svelte';
  import { EventType } from '$lib/sdk/v1/scheduler/event_pb';

  /**
    @type {{
      event: import("$lib/sdk/v1/scheduler/event_pb").Event
      height: number
      immutable?: boolean
      [key: string]: any
    }}
  */
  let { event, immutable, height, ...others } = $props();

  const kitchenFormatter = new Intl.DateTimeFormat(undefined, {
    hour: 'numeric',
    minute: '2-digit',
    hour12: false,
  });

  const scoreFormatter = new Intl.NumberFormat(undefined, {
    signDisplay: 'always',
  });
</script>

<div
  style="height: {height}px"
  class="flex overflow-hidden flex-col w-full text-center rounded-xl select-none glass"
  class:justify-center={height < 40}
  class:opacity-40={immutable}
  {...others}
>
  <div class="flex flex-row gap-2 justify-start items-center p-2 w-full text-xs sm:text-base">
    <p class="mr-auto font-bold {immutable ? 'line-through' : ''}">{event.name}</p>
    <span class="text-slate-400/40">
      {kitchenFormatter.format(Number(event.startTime) * 1000)} -
      {kitchenFormatter.format(Number(event.stopTime) * 1000)}
    </span>
    <EventTypeIcon
      type={event.type}
      class="w-3 h-3 sm:w-5 sm:h-5"
      svgClass="w-3 h-3 sm:w-5 sm:h-5"
    />
    {#if event.timerStartTime > 0 && event.timerStopTime <= 0}
      <!-- prettier-ignore -->
      <svg class="w-3 h-3 sm:w-5 sm:h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" d="M12,1A11,11,0,1,0,23,12,11,11,0,0,0,12,1Zm0,20a9,9,0,1,1,9-9A9,9,0,0,1,12,21Z"/><rect width="2" height="7" x="11" y="6" fill="currentColor" rx="1"><animateTransform attributeName="transform" dur="9s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></rect><rect width="2" height="9" x="11" y="11" fill="currentColor" rx="1"><animateTransform attributeName="transform" dur="0.75s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/></rect></svg>
    {:else if event.timerStartTime > 0 && event.timerStopTime > 0}
      <p class={GetChangeDecorator(event.ratingChange)}>
        {scoreFormatter.format(event.ratingChange)}
      </p>
    {/if}
  </div>
</div>
