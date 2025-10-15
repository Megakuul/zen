import { Code, ConnectError } from '@connectrpc/connect';
import { toast } from 'svelte-sonner';

/** 
 * Exec wraps a function and reports thrown ConnectError exceptions
 * @param unauth controls whether unauthenticated responses emit an error or not
 * @param processing state variable that can be used to check if the operation is processing (ref)
 */
export async function Exec(fn: () => Promise<void>, unauth?: boolean, processing?: boolean): Promise<void> {
  processing = true
  try {
    const result = await fn()
    processing = false
    return result
  } catch (e: unknown) {
    const err = ConnectError.from(e)
    if (unauth || err.code !== Code.Unauthenticated) {
      toast.error(err.name, {
        description: String(err.message)
      })
    }
    processing = false
    return undefined;
  }
} 