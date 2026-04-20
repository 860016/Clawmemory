from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user
from app.services.daily_report_service import DailyReportService
from app.services.license_service import is_feature_enabled
from datetime import datetime
import asyncio

router = APIRouter(prefix="/api/v1/reports", tags=["daily_reports"])


@router.get("/daily")
def list_reports(
    limit: int = Query(30, ge=1, le=90),
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """获取日报列表"""
    service = DailyReportService(db)
    reports = service.list_reports(limit)
    return [
        {
            "id": r.id,
            "report_date": r.report_date,
            "summary": r.summary,
            "highlights": r.highlights or [],
            "stats": r.stats or {},
            "generated_at": str(r.generated_at) if r.generated_at else None,
            "is_pushed": r.is_pushed,
        }
        for r in reports
    ]


@router.get("/daily/{date_str}")
def get_report(
    date_str: str,
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """获取指定日期的日报"""
    service = DailyReportService(db)
    report = service.get_report(date_str)
    if not report:
        raise HTTPException(status_code=404, detail="Report not found")
    return {
        "id": report.id,
        "report_date": report.report_date,
        "summary": report.summary,
        "highlights": report.highlights or [],
        "knowledge_gained": report.knowledge_gained or [],
        "pending_tasks": report.pending_tasks or [],
        "tomorrow_suggestions": report.tomorrow_suggestions or [],
        "stats": report.stats or {},
        "ai_model": report.ai_model,
        "generated_at": str(report.generated_at) if report.generated_at else None,
        "is_pushed": report.is_pushed,
        "push_time": str(report.push_time) if report.push_time else None,
    }


@router.post("/daily/generate")
async def generate_report(
    date: str | None = Query(None),
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """手动生成日报（Pro 功能）"""
    if not is_feature_enabled("ai_extract"):
        raise HTTPException(status_code=403, detail="Pro feature: daily report")

    service = DailyReportService(db)

    if date:
        target = datetime.strptime(date, "%Y-%m-%d")
    else:
        target = datetime.now()

    report = await service.generate_report(target)

    if not report:
        return {"message": "No data for this date, report skipped"}

    return {
        "message": "Report generated",
        "report_date": report.report_date,
        "summary": report.summary,
    }


@router.get("/daily/stats")
def get_stats_summary(
    days: int = Query(7, ge=1, le=90),
    _=Depends(get_current_user),
    db: Session = Depends(get_db),
):
    """获取日报统计摘要"""
    service = DailyReportService(db)
    return service.get_stats_summary(days)
