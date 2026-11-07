import { NextResponse } from 'next/server'
import { thunderIDProxy, createRouteMatcher, } from '@thunderid/nextjs/server'
import { parseScopes } from '@/lib/granted-scopes'
import { ACCESS_DENIED_PATH, ADMIN_HOME_PATH, gateDecision } from '@/lib/access-gate'

const isKpiRoute = createRouteMatcher(['/kpis', '/kpis/*'])

export default thunderIDProxy(async (thunderid, request) => {
  const isSignedIn = thunderid.isSignedIn()
  const session = isSignedIn ? await thunderid.getSession() : undefined

  const decision = gateDecision({
    isSignedIn,
    grantedScopes: parseScopes(session?.scopes),
    pathname: request.nextUrl.pathname,
    isProtectedPath: isKpiRoute(request),
    hasAuthCode: request.nextUrl.searchParams.has('code'),
  })

  switch (decision) {
    // Not signed in on a protected page: send to sign-in. protectRoute also lets the OAuth callback through.
    case 'sign-in':
      return await thunderid.protectRoute()
    // Signed in without app access: end the session before anything renders, so nothing inside shows.
    case 'deny':
      return NextResponse.redirect(new URL(ACCESS_DENIED_PATH, request.url))
    case 'to-admin':
      return NextResponse.redirect(new URL(ADMIN_HOME_PATH, request.url))
    case 'to-home':
      return NextResponse.redirect(new URL('/', request.url))
  }
})

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|.*\\.(?:svg|png|jpg|jpeg|gif|webp|avif|ico)$).*)',
  ],
}
