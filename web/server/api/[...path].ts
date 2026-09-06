export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const path = event.path || '/'
  const method = getMethod(event)

  return $fetch(`${config.backendBaseUrl}${path}`, {
    method,
    body: ['GET', 'HEAD'].includes(method) ? undefined : await readBody(event),
  })
})
