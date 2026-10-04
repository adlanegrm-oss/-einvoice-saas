from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/rnd", tags=["Company R&D"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/rnd — Laboratoire de conformité réglementaire</h2>"