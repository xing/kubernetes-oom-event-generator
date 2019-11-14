#! /bin/bash

set -ev

app_name="$1"
image="$2"

if [[ -z "$app_name" || -z "$image" ]]; then
  >&2 echo "Need to provide app_name image"
  exit 1
fi

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push "$image"

current_version=$(docker run --rm "$image" --version | grep -oE "$app_name [^ ]+" | cut -d ' ' -f2)
current_version_commit=$(git rev-parse "$current_version")
head_commit=$(git rev-parse HEAD)

if [[ "$head_commit" = "$current_version_commit" ]]; then
  docker tag "$image" "$image:$current_version"
  docker push "$image:$current_version"
fi
