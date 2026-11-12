import { NextResponse } from 'next/server'
import { thunderIDProxy, createRouteMatcher, } from '@thunderid/nextjs/server'
import { parseScopes } from '@/lib/granted-scopes'
import { ACCESS_DENIED_PATH, adminHomePath, gateDecision } from '@/lib/access-gate'

// The admin pages: signed-out visitors are sent to sign in, users without access are denied.
const isAdminRoute = createRouteMatcher(['/kpis', '/kpis/*', '/data-sources', '/data-sources/*'])

export default thunderIDProxy(async (thunderid, request) => {
  const isSignedIn = thunderid.isSignedIn()
  const session = isSignedIn ? await thunderid.getSession() : undefined

  const grantedScopes = parseScopes(session?.scopes)
  const decision = gateDecision({
    isSignedIn,
    grantedScopes,
    pathname: request.nextUrl.pathname,
    isProtectedPath: isAdminRoute(request),
    hasAuthCode: request.nextUrl.searchParams.has('code'),
  })

  switch (decision) {
    case 'sign-in':
      return await thunderid.protectRoute({ redirect: '/' })
    // Signed in with no page scope: end the session before anything renders, so nothing inside shows.
    case 'deny':
      return NextResponse.redirect(new URL(ACCESS_DENIED_PATH, request.url))
    // The first admin page the user can open (KPIs, else Data Sources).
    case 'to-admin':
      return NextResponse.redirect(new URL(adminHomePath(grantedScopes) ?? ACCESS_DENIED_PATH, request.url))
    case 'to-home':
      return NextResponse.redirect(new URL('/', request.url))
  }
})

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|.*\\.(?:svg|png|jpg|jpeg|gif|webp|avif|ico)$).*)',
  ],
}
