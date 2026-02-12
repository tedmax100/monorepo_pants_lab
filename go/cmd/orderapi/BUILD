go_package()

go_binary(name="bin")

docker_image(
    name="docker",
    repository="ghcr.io/{build_args.GITHUB_REPOSITORY_OWNER}/orderapi",
    image_tags=["{env.IMAGE_TAG}"],
    registries=["@ghcr"],
    dependencies=[":bin"],
)
