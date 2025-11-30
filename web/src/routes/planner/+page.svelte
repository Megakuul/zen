<script>
  import { create } from '@bufbuild/protobuf';
  import { PlanningClient } from '$lib/client/client.svelte';
  import { GetRequestSchema } from '$lib/sdk/v1/scheduler/planning/planning_pb';
  import { Exec } from '$lib/error/error.svelte';

  /** @type {import("$lib/sdk/v1/scheduler/event_pb").Event[]}*/
  let events = $state([]);

  let loading = $state(false);

  let day = $state(new Date());

  let morning = $derived(
    new Date(Date.UTC(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate())),
  );

  let evening = $derived(
    new Date(Date.UTC(day.getUTCFullYear(), day.getUTCMonth(), day.getUTCDate(), 23, 59, 59, 999)),
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
</script>

<div>
  {#each events as event}
    <div>{event.name}</div>
  {/each}
</div>

