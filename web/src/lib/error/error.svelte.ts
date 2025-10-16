import { Code, ConnectError } from '@connectrpc/connect';
import { toast } from 'svelte-sonner';

/** 
 * Exec wraps a function and reports thrown ConnectError exceptions
 * @param unauth provides a function that handles "unauthenticated" errors, if omitted "unauthenticated" is considered an error.
 * @param processing state variable that can be used to check if the operation is processing (ref)
 */
export async function Exec(fn: () => Promise<void>, unauth?: () => Promise<void>, processing?: boolean): Promise<void> {
  processing = true
  try {
    const result = await fn()
    processing = false
    return result
  } catch (e: unknown) {
    const err = ConnectError.from(e)
    if (unauth !== undefined && err.code === Code.Unauthenticated) {
      await unauth()
    } else {
      toast.error(err.name, {
        description: String(err.message)
      })
    }
    processing = false
    return undefined;
  }
} 