#!/bin/bash

# 判断环境和下载路径
# if [[ "$ENV" == "development" ]]; then
#   MAC_URL="http://$NEXUS_HOST_PORT/repository/raw-hosted/intel-ai-pc/oadin/dist/mac/$MAC_FILE_NAME"
#   WIN_URL="http://$NEXUS_HOST_PORT/repository/raw-hosted/intel-ai-pc/oadin/dist/win/$WIN_FILE_NAME"
# else
#   MAC_URL="https://oss-aipc.dcclouds.com/oadin/dist/macos/$MAC_FILE_NAME"
#   WIN_URL="https://oss-aipc.dcclouds.com/oadin/dist/windows/$WIN_FILE_NAME"
# fi

# 如果$MAC_FILE_NAME和$WIN_FILE_NAME都包含test，则使用测试地址
if [[ "$MAC_FILE_NAME" == *"test"* && "$WIN_FILE_NAME" == *"test"* ]]; then
  MAC_URL="http://10.3.70.145:32018/repository/raw-hosted/intel-ai-pc/oadin/dist/mac/$MAC_FILE_NAME"
  WIN_URL="http://10.3.70.145:32018/repository/raw-hosted/intel-ai-pc/oadin/dist/win/$WIN_FILE_NAME"
else
  MAC_URL="https://oss-aipc.dcclouds.com/oadin/dist/macos/$MAC_FILE_NAME"
  WIN_URL="https://oss-aipc.dcclouds.com/oadin/dist/windows/$WIN_FILE_NAME"
fi

BODY=$(cat <<EOF
{"content":"🎉 oadin：\n 🍎macDownloadUrl: $MAC_URL \n 🌊winDownloadUrl: $WIN_URL"}
EOF
)

# 发送通知
curl -X POST -H "Content-Type: application/json" -d "$BODY" "$HOOK_BRIDGE_URL?projectFlag=$PROJECT_FLAG"