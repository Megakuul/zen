<script>
  import { create } from '@bufbuild/protobuf';
  import {PlanningClient} from "$lib/client/client.svelte"
  import { GetRequestSchema } from "$lib/sdk/v1/scheduler/planning/planning_pb";
  import { onMount } from 'svelte';

  /** @type {import("$lib/sdk/v1/scheduler/event_pb").Event[]}*/
  let events = $state([]);

  async function fetchEvents() {
    try {
      const response = await PlanningClient().get(create(GetRequestSchema, {
        day: BigInt(Date.now() / 1000),
      }))
      events = response.events
    } catch (err) {
      console.error(err)
    }
  }

  onMount(() => {
    fetchEvents()
  })
</script>

<div>
  {#each events as event}
    <div>{event.name}</div>
  {/each}
</div>