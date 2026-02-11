python_sources(name="lib")

python_tests(
    name="tests",
    dependencies=["//:reqs#httpx"],
)

pex_binary(
    name="bin",
    entry_point="app.py",
)
