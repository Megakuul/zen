import { Code, ConnectError } from '@connectrpc/connect';
import { toast } from 'svelte-sonner';

/**
 * Exec wraps a function and reports thrown ConnectError exceptions
 * @param unauth provides a function that handles "unauthenticated" errors, if omitted "unauthenticated" is considered an error.
 * @param change callback used to retrieve the state of the operation (processing or not-processing).
 */
export async function Exec(
  fn: () => Promise<void>,
  unauth?: () => Promise<void>,
  callback?: (processing: boolean) => void,
): Promise<void> {
  if (callback) callback(true);
  try {
    const result = await fn();
    if (callback) callback(false);
    return result;
  } catch (e: unknown) {
    const err = ConnectError.from(e);
    if (unauth !== undefined && err.code === Code.Unauthenticated) {
      await unauth();
    } else {
      toast.error(err.name, {
        description: String(err.message),
      });
    }
    if (callback) callback(false);
    return undefined;
  }
}

