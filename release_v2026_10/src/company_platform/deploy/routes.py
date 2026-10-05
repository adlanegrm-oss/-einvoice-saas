from fastapi import APIRouter
from fastapi.responses import HTMLResponse
router = APIRouter(prefix="/deploy", tags=["Company Deploy"])
@router.get("/dashboard", response_class=HTMLResponse)
async def dash(): return "<h2>/deploy — Plan de contrôle des releases</h2>"