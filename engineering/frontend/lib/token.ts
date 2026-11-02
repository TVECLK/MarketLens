'use server'

import { thunderid } from '@thunderid/nextjs/server'


export async function getAccessToken(): Promise<string | undefined> {
  try {
    const sdk = await thunderid()
    const sessionId = await sdk.getSessionId()
    if (!sessionId) return undefined
    const token = await sdk.getAccessToken(sessionId)

    return token
  } catch {
    return undefined
  }
}