go_package()

go_binary(name="bin")

docker_image(
    name="docker",
    repository="ghcr.io/{env.GITHUB_REPOSITORY_OWNER}/userapi",
    image_tags=["{env.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":bin"],
)
