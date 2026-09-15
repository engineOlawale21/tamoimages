# Browser session strategy

The browser must not persist access or refresh tokens in local storage, session storage, IndexedDB, URLs, or readable cookies.

The web application uses a backend-for-frontend session boundary. After successful identity verification, the server stores or exchanges credentials server-side and issues an opaque `__Host-tamo-session` cookie with `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, no `Domain`, bounded lifetime, and rotation after authentication or privilege changes. Production mutation requests also require same-origin checks and CSRF protection appropriate to the endpoint.

Middleware uses cookie presence only for early navigation redirects. It is not an authorization decision. Server routes validate the session and role before returning protected data, and backend APIs independently validate bearer claims and resource ownership.

Unauthenticated requests redirect to `/login` with a same-origin `returnTo` path. Expired sessions are cleared and treated as unauthenticated. Authenticated users with the wrong role receive a forbidden response and are not redirected into a different privileged area. Logout invalidates server state and expires the cookie.
