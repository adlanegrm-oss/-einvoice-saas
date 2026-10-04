from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/admin", tags=["Company Admin"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/admin — Business Control Center</h2>"