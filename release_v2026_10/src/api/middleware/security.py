from fastapi import Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

class SecurityAndTenantMiddleware(BaseHTTPMiddleware):
    INTERNAL_PREFIXES = ('/admin', '/ops', '/mco', '/rnd', '/deploy')

    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        if any(path.startswith(prefix) for prefix in self.INTERNAL_PREFIXES):
            auth_header = request.headers.get('Authorization', '')
            internal_role = request.headers.get('X-Internal-Role', '')
            mfa_verified = request.headers.get('X-MFA-Verified', 'false').lower() == 'true'

            if not auth_header.startswith('Bearer sec_internal_') or internal_role not in ['SRE', 'COMPLIANCE_OFFICER', 'ADMIN_EXEC', 'DEVOPS']:
                return JSONResponse(status_code=403, content={'detail': 'Accès refusé : Authentification interne requise.'})
            if not mfa_verified:
                return JSONResponse(status_code=401, content={'detail': 'MFA Obligatoire pour accéder aux plans d\'administration interne.'})

        response = await call_next(request)
        response.headers['X-Content-Type-Options'] = 'nosniff'
        response.headers['X-Frame-Options'] = 'DENY'
        return response
