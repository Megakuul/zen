import { createClient, Code, ConnectError, type Interceptor, type Transport } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { PlannerService } from "$lib/sdk/v1/scheduler/planner_pb";
import { TimerService } from "$lib/sdk/v1/scheduler/timer_pb";

let transportUrl = $state("")
let transportToken = $state("")

export function SetUrl(url: string) {
	transportUrl = url
}

export function SetToken(token: string) {
  transportToken = token
}

let transport: Transport = $derived(createConnectTransport({
	baseUrl: transportUrl,
	interceptors: [(next) => async (req) => {
		req.header.set("authorization", transportToken)
		return await next(req)
	}]
}))

let plannerClient = $derived(createClient(PlannerService, transport))
let timerClient = $derived(createClient(TimerService, transport))

export const PlannerClient = () => plannerClient;
export const TimerClient = () => timerClient;