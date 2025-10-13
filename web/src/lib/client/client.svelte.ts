import { createClient, type Transport } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ManagementService } from "$lib/sdk/v1/manager/management/management_pb";
import { AuthenticationService } from "$lib/sdk/v1/manager/authentication/authentication_pb";
import { PlanningService } from "$lib/sdk/v1/scheduler/planning/planning_pb";
import { TimingService } from "$lib/sdk/v1/scheduler/timing/timing_pb";
import { GetToken } from './auth.svelte';

let transportUrl = $state("")

export function SetUrl(url: string) {
	transportUrl = url
}

let transport: Transport = $derived(createConnectTransport({
	baseUrl: transportUrl,
	interceptors: [(next) => async (req) => {
		req.header.set("authorization", await GetToken())
		return await next(req)
	}],
}))

let authTransport: Transport = $derived(createConnectTransport({
	baseUrl: transportUrl,
}))

let authenticationClient = $derived(createClient(AuthenticationService, authTransport))
let managementClient = $derived(createClient(ManagementService, transport))
let planningClient = $derived(createClient(PlanningService, transport))
let timingClient = $derived(createClient(TimingService, transport))

export let AuthenticationClient = () => authenticationClient
export let ManagementClient = () => managementClient
export let PlanningClient = () => planningClient
export let TimingClient = () => timingClient