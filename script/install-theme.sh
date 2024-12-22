#!/bin/sh

if [ $# -lt 1 ]; then
    echo "Usage: $0 <theme-repo>"
    exit 1
fi

GITHUB_RAW_URL="raw.githubusercontent.com/chenx-dust/nezha-compat/compat"
NZ_DASHBOARD_PATH="/dashboard"

_repo=$1
_version=$(wget -qO- --timeout=10 "https://api.github.com/repos/${_repo}/releases/latest" 2>/dev/null | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g')

if [ -z "$_version" ]; then
    echo "获取版本号失败，请检查本机能否链接 https://api.github.com/repos/${_repo}/releases/latest"
    exit 1
else
    echo "${_repo} 最新版本为: ${_version}"
fi

NZ_DASH_URL="https://github.com/${_repo}/releases/download/${_version}/dist.zip"

TMP_DIR=$(mktemp -d)
echo "下载主题文件..."
wget -qO ${TMP_DIR}/dist.zip "${NZ_DASH_URL}" >/dev/null 2>&1
unzip -qq -o ${TMP_DIR}/dist.zip -d ${TMP_DIR}
if [ $? -ne 0 ]; then
    echo "解压主题文件失败，请检查仓库地址是否正确"
    exit 1
fi
# fix viewpassword.html
wget -qO ${TMP_DIR}/viewpassord.html https://${GITHUB_RAW_URL}/resource/template/theme-default/viewpassword.html >/dev/null 2>&1
sed -i "s|theme-default|theme-custom|g" ${TMP_DIR}/viewpassord.html

if [ -d "${NZ_DASHBOARD_PATH}/resource/template/theme-custom" ] || [ -d "${NZ_DASHBOARD_PATH}/resource/static/custom" ]; then
    echo "您可能已经安装过自定义主题，重复安装会覆盖现有主题，请注意备份。"
    printf "是否继续? [Y/n] "
    read -r input
    case $input in
    [yY][eE][sS] | [yY])
        ;;
    [nN][oO] | [nN])
        echo "退出安装"
        exit 0
        ;;
    *)
        echo "继续安装"
        ;;
    esac
fi

echo "清理原来的主题文件..."
if [ -d "${NZ_DASHBOARD_PATH}/resource/template/theme-custom" ]; then
    rm -rf "${NZ_DASHBOARD_PATH}/resource/template/theme-custom"
fi
if [ -d "${NZ_DASHBOARD_PATH}/resource/static/custom" ]; then
    rm -rf "${NZ_DASHBOARD_PATH}/resource/static/custom"
fi
mkdir -p "${NZ_DASHBOARD_PATH}/resource/template/theme-custom" "${NZ_DASHBOARD_PATH}/resource/static/custom" >/dev/null 2>&1

echo "替换静态文件..."
ASSETS_FILES=$(find ${TMP_DIR}/dist/ -type f)
for filename in ${ASSETS_FILES}; do
    NOW_FILE=$(echo $filename | sed "s|^\\${TMP_DIR}/dist/|/|")
    echo "replacing \"${NOW_FILE}\" to \"/static${NOW_FILE}\" ..."
    for file in ${ASSETS_FILES}; do
        sed -i "s|${NOW_FILE}|/static${NOW_FILE}|g" $file
    done
done

echo "安装模板文件..."
cat <<EOF > ${NZ_DASHBOARD_PATH}/resource/template/theme-custom/home.html
{{define "theme-custom/home"}}
$(cat ${TMP_DIR}/dist/index.html)
{{end}}
EOF
cp ${TMP_DIR}/viewpassord.html ${NZ_DASHBOARD_PATH}/resource/template/theme-custom/viewpassword.html
cp -rf ${TMP_DIR}/dist/* ${NZ_DASHBOARD_PATH}/resource/static/custom/

echo "清理临时文件..."
rm ${NZ_DASHBOARD_PATH}/resource/static/custom/index.html
rm -rf ${TMP_DIR}

echo
echo "${_repo} 主题安装成功"
echo "为了更好的体验，建议打开设置中的 使用界面主题处理无路由情况"
