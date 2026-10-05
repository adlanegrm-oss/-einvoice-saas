from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/partner", tags=["Partner Platform"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/partner — Plateforme API & Intégrateurs</h2>"