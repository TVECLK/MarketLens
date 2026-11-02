import { thunderIDProxy, createRouteMatcher, } from '@thunderid/nextjs/server'

const isProtectedRoute = createRouteMatcher(['/kpis', '/kpis/*'])

export default thunderIDProxy(async (thunderid, request) => {
  if (isProtectedRoute(request)) {
    return await thunderid.protectRoute()
  }
})

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|.*\\.(?:svg|png|jpg|jpeg|gif|webp|avif|ico)$).*)',
  ],
}