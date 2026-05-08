package modules

// 引入模块
// NOTE: app_bot imports bot_api, so Go guarantees bot_api init runs first.
// Auth DB fallback in bot_api covers any construction-time race with app_bot's registry setup.
import (
	_ "github.com/Mininglamp-OSS/octo-server/modules/app_bot"
	_ "github.com/Mininglamp-OSS/octo-server/modules/backup"
	_ "github.com/Mininglamp-OSS/octo-server/modules/base"
	_ "github.com/Mininglamp-OSS/octo-server/modules/bot_api"
	_ "github.com/Mininglamp-OSS/octo-server/modules/botfather"
	_ "github.com/Mininglamp-OSS/octo-server/modules/category"
	_ "github.com/Mininglamp-OSS/octo-server/modules/channel"
	_ "github.com/Mininglamp-OSS/octo-server/modules/common"
	_ "github.com/Mininglamp-OSS/octo-server/modules/file"
	_ "github.com/Mininglamp-OSS/octo-server/modules/group"
	_ "github.com/Mininglamp-OSS/octo-server/modules/message"
	_ "github.com/Mininglamp-OSS/octo-server/modules/notify"
	_ "github.com/Mininglamp-OSS/octo-server/modules/oidc"
	_ "github.com/Mininglamp-OSS/octo-server/modules/openapi"
	_ "github.com/Mininglamp-OSS/octo-server/modules/qrcode"
	_ "github.com/Mininglamp-OSS/octo-server/modules/report"
	_ "github.com/Mininglamp-OSS/octo-server/modules/robot"
	_ "github.com/Mininglamp-OSS/octo-server/modules/search"
	_ "github.com/Mininglamp-OSS/octo-server/modules/space"
	_ "github.com/Mininglamp-OSS/octo-server/modules/statistics"
	_ "github.com/Mininglamp-OSS/octo-server/modules/thread"
	_ "github.com/Mininglamp-OSS/octo-server/modules/user"
	_ "github.com/Mininglamp-OSS/octo-server/modules/voice"
	_ "github.com/Mininglamp-OSS/octo-server/modules/webhook"
	_ "github.com/Mininglamp-OSS/octo-server/modules/workplace"
)
