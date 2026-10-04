from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/ops", tags=["Company Ops"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/ops — Compliance Operations Center</h2>"