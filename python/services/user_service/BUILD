python_sources(name="lib")

python_tests(
    name="tests",
    dependencies=["//:reqs#httpx"],
)

pex_binary(
    name="bin",
    entry_point="app.py",
)

# Docker-specific binary: entry point is uvicorn (python -m uvicorn).
# Explicitly depends on :lib (app sources) and uvicorn (not inferred because app.py
# doesn't import uvicorn directly).
pex_binary(
    name="server",
    entry_point="uvicorn",
    dependencies=[":lib", "//:reqs#uvicorn"],
)

docker_image(
    name="docker",
    repository="ghcr.io/{env.GITHUB_REPOSITORY_OWNER}/user-service",
    image_tags=["{env.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":server"],
)
