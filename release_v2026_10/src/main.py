"""
Point d'entrée principal de l'application
"""
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
import os

from .api.v1.compliance.router import router as compliance_router
from .api.middleware.security import SecurityAndTenantMiddleware
from .customer_platform.app.routes import router as app_router
from .customer_platform.partner.routes import router as partner_router
from .company_platform.admin.routes import router as admin_router
from .company_platform.ops.routes import router as ops_router
from .company_platform.mco.routes import router as mco_router
from .company_platform.rnd.routes import router as rnd_router
from .company_platform.deploy.routes import router as deploy_router

app = FastAPI(title="E-Invoice Compliance & Validation Platform", version="2026.10-compliance-core")
app.add_middleware(SecurityAndTenantMiddleware)

app.include_router(compliance_router)
app.include_router(app_router)
app.include_router(partner_router)
app.include_router(admin_router)
app.include_router(ops_router)
app.include_router(mco_router)
app.include_router(rnd_router)
app.include_router(deploy_router)

@app.get("/", response_class=HTMLResponse)
async def landing():
    landing_path = os.path.join(os.path.dirname(__file__), "landing_page", "index.html")
    with open(landing_path, "r", encoding="utf-8") as f:
        return f.read()

@app.get("/healthz")
async def healthz():
    return {"status": "HEALTHY", "platform": "E-Invoice Compliance Core"}