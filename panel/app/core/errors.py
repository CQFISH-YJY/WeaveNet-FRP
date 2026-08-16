"""统一业务错误与响应格式。

遵循设计文档第 12 章：成功 {code:0, message:"ok", data}；
失败含 business_code 业务码。
"""
from __future__ import annotations

from typing import Any

from fastapi import FastAPI, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException


class BizError(Exception):
    """业务错误。

    属性：
        http_code: HTTP 状态码
        business_code: 业务码
        message: 中文提示
    """

    def __init__(self, http_code: int, business_code: int, message: str) -> None:
        self.http_code = http_code
        self.business_code = business_code
        self.message = message
        super().__init__(message)


class AuthError(BizError):
    """认证失败。"""

    def __init__(self, business_code: int = 0, message: str = "未登录或登录已过期") -> None:
        super().__init__(401, business_code, message)


class ForbiddenError(BizError):
    """无权限。"""

    def __init__(self, business_code: int = 0, message: str = "没有权限执行此操作") -> None:
        super().__init__(403, business_code, message)


class NotFoundError(BizError):
    """资源不存在。"""

    def __init__(self, business_code: int = 0, message: str = "资源不存在") -> None:
        super().__init__(404, business_code, message)


class ConflictError(BizError):
    """业务冲突。"""

    def __init__(self, business_code: int, message: str) -> None:
        super().__init__(409, business_code, message)


class RateLimitedError(BizError):
    """限流触发。"""

    def __init__(self, business_code: int = 0, message: str = "请求过于频繁，请稍后再试") -> None:
        super().__init__(429, business_code, message)


def success(data: Any = None, message: str = "ok", http_code: int = 200) -> JSONResponse:
    """统一成功响应。"""
    return JSONResponse(
        status_code=http_code,
        content={"code": 0, "message": message, "data": data},
    )


def register_exception_handlers(app: FastAPI) -> None:
    """注册全局异常处理器。"""

    @app.exception_handler(BizError)
    async def biz_error_handler(request: Request, exc: BizError) -> JSONResponse:  # noqa: ANN001
        return JSONResponse(
            status_code=exc.http_code,
            content={
                "code": exc.http_code,
                "business_code": exc.business_code or None,
                "message": exc.message,
                "data": None,
            },
        )

    @app.exception_handler(StarletteHTTPException)
    async def http_error_handler(
        request: Request, exc: StarletteHTTPException  # noqa: ANN001
    ) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content={
                "code": exc.status_code,
                "business_code": None,
                "message": str(exc.detail),
                "data": None,
            },
        )

    @app.exception_handler(RequestValidationError)
    async def validation_error_handler(
        request: Request, exc: RequestValidationError  # noqa: ANN001
    ) -> JSONResponse:
        first = exc.errors()[0] if exc.errors() else {}
        loc = ".".join(str(x) for x in first.get("loc", []))
        msg = first.get("msg", "参数校验失败")
        return JSONResponse(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            content={
                "code": 422,
                "business_code": None,
                "message": f"参数错误: {loc} {msg}",
                "data": None,
            },
        )

    @app.exception_handler(Exception)
    async def unhandled_error_handler(
        request: Request, exc: Exception  # noqa: ANN001
    ) -> JSONResponse:
        # 内部错误返回统一格式，日志记录完整堆栈
        import logging

        logging.getLogger("weavenet").exception("Unhandled error")
        return JSONResponse(
            status_code=500,
            content={
                "code": 500,
                "business_code": 9001,
                "message": "服务器内部错误，请稍后重试",
                "data": None,
            },
        )
