import { getAccessToken } from "@/lib/with-auth";
import { hasScope, scopesFromAccessToken } from "@/lib/granted-scopes";

/**
 * Whether the signed-in user's access token grants `required`. For server components; false when
 * there is no session. Decides what the server renders, so nothing unauthorised reaches the browser.
 */
export async function hasServerPermission(required: string): Promise<boolean> {
  return hasScope(scopesFromAccessToken(await getAccessToken()), required);
}
