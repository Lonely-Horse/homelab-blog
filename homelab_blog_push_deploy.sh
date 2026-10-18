#!/usr/bin/bash
set -Eeuo pipefail

Work_dir="${HOME}/go_learn/homelab-blog"
cd ${Work_dir}

version="${1:-}"
commit="${2:-}"

if [ -z "$version" ]; then
    version="v$(date +%y.%m.%d)"
    echo "未指定版本号，自动生成: ${version}"
fi

if [[ ! "$version" =~ ^v[0-9]{2}\.[0-9]{2}\.[0-9]{2}$ ]]; then
    echo "非法版本号: ${version}（期望格式: v26.10.18）"
    exit 1
fi

if [ -z "$commit" ]; then
    commit="$(git -C "${Work_dir}" rev-parse HEAD)"
    echo "未指定 commit，使用当前 HEAD: ${commit}"
fi

if [[ ! "$commit" =~ ^[0-9a-f]{7,40}$ ]]; then
    echo "非法 commit SHA: ${commit}"
    exit 1
fi

Image_address="100.87.250.21:5000"
Image_name="homelab-blog"
Image_latest="${Image_address}/${Image_name}:latest"
Image_version="${Image_address}/${Image_name}:${version}"
MP9_HOST="lonelyhorse@100.94.116.106"
MP9_DEPLOY_SCRIPT="${HOME}/server/homelab-blog/pull_deploy.sh"
mp9_blog_address="${HOME}/server/homelab-blog"
ssh_key="${HOME}/.ssh/ser7_to_mp9"
ssh_opts="-i ${ssh_key} -o BatchMode=yes -o StrictHostKeyChecking=accept-new"

Registry_container="local_registry_cache"
Registry_data="${HOME}/docker_registry_data"

fetch_commit() {
    for i in 1 2 3; do
        echo "拉取远程 main，尝试${i}/3"
        if git fetch origin main && git cat-file -e "${commit}^{commit}" 2>/dev/null; then
            return 0
        fi

        if [ "${i}" -lt 3 ]; then
            echo "拉取失败，等待5s，继续尝试"
            sleep 5
        fi
    done
    echo "远程拉取失败，请稍后再次尝试"
    return 1
}

fetch_commit

build_dir="/tmp/my_blog_build_${version}"

cleanup(){
    docker start "${Registry_container}" >/dev/null 2>&1 || true
    if [ -n "${build_dir}" ] && [ -d "${build_dir}" ]; then
        echo "清理临时推送目录: ${build_dir}"
        git -C "${Work_dir}" worktree remove --force "${build_dir}" 2>/dev/null || true
        rm -rf "${build_dir}" 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo "清理旧的临时目录"
if [ -d "${build_dir}" ]; then
    git -C "${Work_dir}" worktree remove --force "${build_dir}" 2>/dev/null || true 
    rm -rf "${build_dir}" 2>/dev/null || true
fi

echo "创建临时构建目录: ${build_dir}"
git -C "${Work_dir}" worktree add --detach "${build_dir}" "${commit}"
cd "${build_dir}"

echo "现在开始原地构建镜像,并打好标签"
docker build \
  -f docker/Dockerfile \
  -t "${Image_latest}" \
  -t "${Image_version}" \
  .

echo "现在开始推送镜像到本地regustry仓库"

docker push ${Image_latest}
docker push ${Image_version}

echo "镜像推送完毕"

echo "清理构建机上的旧镜像与过期构建缓存"
docker image prune -f
docker builder prune -f --filter until=168h

scp ${ssh_opts} ./pull_deploy.sh "${MP9_HOST}:${mp9_blog_address}/"

echo "触发 T630 部署..."
ssh ${ssh_opts} "$MP9_HOST" "bash ${MP9_DEPLOY_SCRIPT} '${version}'"

echo "回收 registry 中无 tag 的旧镜像（期间 registry 短暂不可用）"
docker stop "${Registry_container}" >/dev/null
if ! timeout 600 docker run --rm \
    -v "${Registry_data}:/var/lib/registry" \
    registry:2 garbage-collect -m /etc/docker/registry/config.yml; then
    echo "GC 失败，跳过本次回收"
fi
docker start "${Registry_container}" >/dev/null
