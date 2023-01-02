import { RouteRecord } from 'vue-router'

function sortRoutes (a: RouteRecord, b: RouteRecord): number {
  if (a.name === 'index') {
    return -1
  }

  if (b.name === 'index') {
    return 1
  }

  return a.name!.toString().localeCompare(b.name!.toString())
}

export function getSortedRoutes () {
  const routes = useRouter().getRoutes()
  routes.sort(sortRoutes)
  return routes
}
