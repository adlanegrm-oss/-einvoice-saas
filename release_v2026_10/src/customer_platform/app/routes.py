from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/app", tags=["Customer App"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/app — Portail Client Conformité</h2>"