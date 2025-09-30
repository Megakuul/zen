<script>
  import { create } from '@bufbuild/protobuf';
  import {PlannerClient} from "$lib/client/client.svelte"
  import { GetRequestSchema } from "$lib/sdk/v1/scheduler/planning/planning_pb";

  /** @type {import("$lib/sdk/v1/scheduler/event_pb").Event[]}*/
  let events = $state([]);

  async function fetchEvents() {
    try {
      const response = await PlannerClient().get(create(GetRequestSchema, {
        day: BigInt(Date.now() / 1000),
      }))
      events = response.events
    } catch (err) {
      console.error(err)
    }
  }

  $effect.root(() => {
    fetchEvents()
  })
</script>

<div>
  {#each events as event}
    <div>{event.name}</div>
  {/each}
</div>