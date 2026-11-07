/**
 * Each resource gets `<resource>:create|update|delete`, matching the backend's
 * `RequireScope(scope + ":create")` etc. To add a resource, add one entry here:
 * it is then requested at sign-in and usable in permission checks via `RESOURCE_SCOPES`.
 */
export const RESOURCE_SCOPES = {
  employmentSectors: 'employment-sectors',
  experiences: 'experiences',
  educationLevels: 'education-levels',
  formalities: 'formalities',
  genders: 'genders',
  vocationalEducations: 'vocational-educations',
  majorGroups: 'major-groups',
  subMajorGroups: 'sub-major-groups',
  minorGroups: 'minor-groups',
  unitGroups: 'unit-groups',
  occupationGroups: 'occupation-groups',
  industrySectors: 'industry-sectors',
  industryDivisions: 'industry-divisions',
  industryGroups: 'industry-groups',
  industryClasses: 'industry-classes',
  industrySubClasses: 'industry-sub-classes',
} as const

export type ResourceScope = (typeof RESOURCE_SCOPES)[keyof typeof RESOURCE_SCOPES]

/** The actions each resource scope is split into, e.g. `experiences:update`. */
export const CRUD_ACTIONS = {
  create: 'create',
  update: 'update',
  delete: 'delete',
} as const

export type CrudAction = (typeof CRUD_ACTIONS)[keyof typeof CRUD_ACTIONS]

/** Every action, in order: create, update, delete. */
export const ALL_CRUD_ACTIONS: readonly CrudAction[] = Object.values(CRUD_ACTIONS)

/** How each action is worded to users, matching the UI's "+ Add", "Edit" and "Delete" buttons. */
export const CRUD_ACTION_LABELS: Record<CrudAction, string> = {
  create: 'add',
  update: 'edit',
  delete: 'delete',
}

/** The permission string the backend checks for an action on a resource, e.g. `experiences:update`. */
export const scopeFor = (resource: ResourceScope, action: CrudAction) => `${resource}:${action}`

// build the appScopes string to send to the IDP
export const appScopes = [
  'openid',
  'profile',
  ...Object.values(RESOURCE_SCOPES).flatMap((resource) => ALL_CRUD_ACTIONS.map((action) => scopeFor(resource, action))),
]

export const scopeString = appScopes.join(' ')
