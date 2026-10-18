#!/usr/bin/bash
set -Eeuo pipefail

Work_dir="${HOME}/go_learn/homelab-blog"
cd ${Work_dir}

version="${1:-}"

if [ -z "$version" ]; then
    read -rp "请输入版本号:" version
fi

if [ -z "$version" ]; then
    echo "版本号不能为空"
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
ssh_opts="-i ${ssh_key} -o BatchMode=yes"

fetch_tag() {
    for i in 1 2 3; do
         echo "拉取远程tag，尝试${i}/3"
        if git fetch origin tag "${version}"; then
            return 0
        fi

        if [ "${i}" -lt 3 ]; then
	    echo "拉取tag失败，等待5s，继续尝试"
	    sleep 5
        fi
    done
    echo "远程拉取失败，清稍后再次尝试"
    return 1
}

fetch_tag

build_dir="/tmp/my_blog_build_${version}"

cleanup(){
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
git -C "${Work_dir}" worktree add --detach "${build_dir}" "refs/tags/${version}"
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

scp ./pull_deploy.sh "${MP9_HOST}:${mp9_blog_address}/"

echo "触发 T630 部署..."
ssh ${ssh_opts} "$MP9_HOST" "bash ${MP9_DEPLOY_SCRIPT} '${version}'"
