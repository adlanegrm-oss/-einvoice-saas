# Headers de sécurité recommandés

À appliquer soit dans le reverse-proxy (Caddyfile déjà préparé), soit via middleware Go.

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
Permissions-Policy: geolocation=(), microphone=(), camera=(), payment=()
X-XSS-Protection: 0
Cache-Control: no-store (sur les réponses authentifiées)
```

Pour Caddy, voir Caddyfile.example (déjà inclus).
