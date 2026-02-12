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
pex_binary(
    name="server",
    entry_point="uvicorn",
    dependencies=[":lib", "//:reqs#uvicorn"],
)

docker_image(
    name="docker",
    repository="ghcr.io/{build_args.GITHUB_REPOSITORY_OWNER}/product-service",
    image_tags=["{build_args.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":server"],
)
