package bot_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Mininglamp-OSS/octo-lib/pkg/wkhttp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// setCommands handles POST /v1/bot/setCommands.
func (ba *BotAPI) setCommands(c *wkhttp.Context) {
	var req struct {
		Commands []struct {
			Command     string `json:"command"`
			Description string `json:"description"`
		} `json:"commands"`
	}
	if err := c.BindJSON(&req); err != nil {
		ba.Error("数据格式有误", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误"))
		return
	}

	robotID := getRobotIDFromContext(c)

	if req.Commands == nil {
		req.Commands = make([]struct {
			Command     string `json:"command"`
			Description string `json:"description"`
		}, 0)
	}
	commandsJSON, err := json.Marshal(req.Commands)
	if err != nil {
		ba.Error("序列化命令列表失败", zap.Error(err))
		c.ResponseError(errors.New("序列化命令列表失败"))
		return
	}

	err = ba.db.updateBotCommands(robotID, string(commandsJSON))
	if err != nil {
		ba.Error("保存命令列表失败", zap.Error(err))
		c.ResponseError(errors.New("保存命令列表失败"))
		return
	}

	c.ResponseOK()
}

// spaceUIDPattern matches space-prefixed UIDs: s{digits}_{baseUID}
var spaceUIDPattern = regexp.MustCompile(`^s\d+_(.+)$`)

// stripSpacePrefix extracts the base UID from a space-prefixed UID.
func stripSpacePrefix(uid string) string {
	if m := spaceUIDPattern.FindStringSubmatch(uid); len(m) == 2 {
		return m[1]
	}
	return uid
}

// getUserInfo handles GET /v1/bot/user/info?uid=xxx.
func (ba *BotAPI) getUserInfo(c *wkhttp.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		c.ResponseError(errors.New("uid参数不能为空"))
		return
	}

	bareUID := stripSpacePrefix(uid)

	userResp, err := ba.userService.GetUser(bareUID)
	if err != nil || userResp == nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "用户不存在"})
		return
	}

	cfg := ba.ctx.GetConfig()
	apiURL := cfg.External.BaseURL
	if strings.TrimSpace(apiURL) == "" {
		apiURL = fmt.Sprintf("http://%s:8090", cfg.External.IP)
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":    userResp.UID,
		"name":   userResp.Name,
		"avatar": fmt.Sprintf("%s/users/%s/avatar", apiURL, userResp.UID),
	})
}
