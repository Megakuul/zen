import { create } from '@bufbuild/protobuf';
import { GetRequestSchema } from "$lib/sdk/v1/manager/authentication/authentication_pb";
import { AuthenticationClient } from './client.svelte';



/**
 * Login to the system (usually needed if GetToken() fails with status 16 "UNAUTHENTICATED").
 * Provide a message channel as verifier to send a code to the user (e.g. email:salami.brot@gmail.com).
 * After the user received a message, insert the message code to the verifier (e.g. code:1234-6789)
 * @param verifier channel to send verification code OR verification code
 */
export async function Login(verifier: string) {
	const response = await AuthenticationClient().get(create(GetRequestSchema, {
    verifier: verifier,
		autoRefresh: true,
	}))
	if (response.token) {
		localStorage.setItem("auth_token", response.token)
	}
}

/**
 * Retrieves the auth token from localstore or fetch a new one from the refresh_token.
 * @returns auth token
 */
export async function GetToken(): Promise<string> {
	const token = localStorage.getItem("auth_token")
	if (token) {
		return token
	}
	const response = await AuthenticationClient().get(create(GetRequestSchema, {
    verifier: "",
		autoRefresh: true,
	}))
	localStorage.setItem("auth_token", response.token)
	return response.token
}