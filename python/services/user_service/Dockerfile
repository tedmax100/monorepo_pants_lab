# Build context root = repository root (Pants sets this automatically).
# Pants places the PEX at address path (/ → . in directory name):
#   python.services.user_service/server.pex
FROM python:3.11-slim

WORKDIR /app

# Copy the self-contained PEX (includes FastAPI, uvicorn, pydantic, and app code)
COPY python.services.user_service/server.pex /app/server.pex

EXPOSE 8080

# ENTRYPOINT runs the PEX as "python -m uvicorn" (entry_point="uvicorn" in pex_binary).
# CMD supplies the app target and binding args passed to uvicorn.
# Module path: user_service.app because /python/services is the source root.
ENTRYPOINT ["python3", "/app/server.pex"]
CMD ["user_service.app:app", "--host", "0.0.0.0", "--port", "8080"]
