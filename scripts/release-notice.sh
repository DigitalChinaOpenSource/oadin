#!/bin/bash

# 判断环境和下载路径
if [[ "$ENV" == "deployment" ]]; then
  MAC_URL="http://$NEXUS_HOST_PORT/repository/raw-hosted/intel-ai-pc/oadin/releases/mac/$MAC_FILE_NAME"
  WIN_URL="http://$NEXUS_HOST_PORT/repository/raw-hosted/intel-ai-pc/oadin/releases/win/$WIN_FILE_NAME"
else
  MAC_URL="https://oss-aipc.dcclouds.com/oadin/releases/macos/$MAC_FILE_NAME"
  WIN_URL="https://oss-aipc.dcclouds.com/oadin/releases/windows/$WIN_FILE_NAME"
fi

BODY=$(cat <<EOF
{"content":"🎉 oadin：\n 🍎macDownloadUrl: $MAC_URL \n 🌊winDownloadUrl: $WIN_URL"}
EOF
)

# 发送通知
curl -X POST -H "Content-Type: application/json" -d "$BODY" "$HOOK_BRIDGE_URL?projectFlag=$PROJECT_FLAG"