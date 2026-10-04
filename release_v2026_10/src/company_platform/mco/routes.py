from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/mco", tags=["Company MCO"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/mco — Reliability & DR Platform</h2>"